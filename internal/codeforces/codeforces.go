package codeforces

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

const (
	ratingUpdateCheckInterval time.Duration = 30 * time.Minute
)

type guildProvider interface {
	getGuilds() []*discordgo.Guild
}

type Handler struct {
	db      *pgxpool.Pool
	discord *discordgo.Session

	Contests *contestService
	Pinger   *contestPinger
	Auth     *authService

	guilds []*discordgo.Guild
	mu     sync.RWMutex

	lbGuildData []lbGuildData
	lbMu        sync.RWMutex
}

func New(db *pgxpool.Pool, discord *discordgo.Session, guilds []*discordgo.Guild) (*Handler, error) {
	h := Handler{db: db, discord: discord, guilds: guilds}

	h.Contests = newContestService(discord)
	h.Contests.addListener(&h)

	h.Pinger = newPinger(discord, h.Contests, &h)

	h.Auth = newAuthService(db, discord)

	if err := h.Pinger.updatePingData(); err != nil {
		return nil, fmt.Errorf("initializing ping guild data: %w", err)
	}

	if err := h.updateLeaderboardGuildData(); err != nil {
		return nil, fmt.Errorf("initializing leaderboard guild data: %w", err)
	}

	return &h, nil
}

func (h *Handler) HandleCommand(args []string, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "contests":
		if err := h.Contests.updateContests(); err != nil {
			return fmt.Errorf("failed updating upcoming contests: %w", err)
		}

		err := h.Contests.listContests(m)
		if err != nil {
			return errors.Join(errors.New("listing future contests failed,"), err)
		}
	case "addDebugContest":
		err := h.Contests.addDebugContest(args, m)
		if err != nil {
			return errors.Join(errors.New("adding debug contest failed,"), err)
		}
	case "authenticate":
		err := h.Auth.authCommand(args, m)
		if err != nil {
			return fmt.Errorf("authentication command failed: %w", err)
		}
	default:
		err := utils.UnknownCommand(h.discord, m)
		return err
	}

	return nil
}

func (h *Handler) getGuilds() []*discordgo.Guild {
	h.mu.RLock()
	defer h.mu.RUnlock()

	res := make([]*discordgo.Guild, len(h.guilds))
	copy(res, h.guilds)
	return res
}

func (h *Handler) onContestEnd(c *contest) {
	ratingUpdates := h.startRatingUpdateCheck(c, ratingUpdateCheckInterval)
	for updated := range ratingUpdates {
		if updated {
			h.sendLeaderboardMessageAll(c)
		}
	}
}
