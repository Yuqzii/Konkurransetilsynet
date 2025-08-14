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

	client      api
	Contests    *contestService
	Pinger      *contestPinger
	auth        *authService
	leaderboard *lbService

	guilds []*discordgo.Guild
	mu     sync.RWMutex
}

func NewHandler(db *pgxpool.Pool, discord *discordgo.Session, client api, guilds []*discordgo.Guild) (*Handler, error) {
	h := Handler{db: db, discord: discord, client: client, guilds: guilds}

	h.Contests = newContestService(discord, client)
	h.Contests.addListener(&h)

	h.Pinger = newPinger(discord, h.Contests, &h)

	h.auth = newAuthService(db, discord, client)

	h.leaderboard = newLeaderboardService(discord, client, &h)

	if err := h.Pinger.updatePingData(); err != nil {
		return nil, fmt.Errorf("initializing ping guild data: %w", err)
	}

	if err := h.leaderboard.updateData(); err != nil {
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
		err := h.auth.authCommand(args, m)
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
	ratingUpdates := h.leaderboard.startRatingUpdateCheck(c, ratingUpdateCheckInterval)
	for updated := range ratingUpdates {
		if updated {
			h.leaderboard.sendLeaderboardMessageAll(c)
		}
	}
}
