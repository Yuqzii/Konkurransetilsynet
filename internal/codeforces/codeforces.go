package codeforces

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

type guildProvider interface {
	getGuilds() []*discordgo.Guild
}

type Handler struct {
	discord *discordgo.Session
	guilds  []*discordgo.Guild
	mu      sync.RWMutex

	Contests    *contestService
	Pinger      *contestPinger
	auth        *authService
	leaderboard *lbService
}

var ErrUserNotConnected error = errors.New("user not connected")

type Repository interface {
	DiscordIDExists(ctx context.Context, discID string) (bool, error)
	AddCodeforcesUser(ctx context.Context, discID, handle string) error
	UpdateCodeforcesUser(ctx context.Context, discID, handle string) error
	GetConnectedCodeforces(ctx context.Context, discID string) (string, error)
}

func NewHandler(db Repository, discord *discordgo.Session, client api, guilds []*discordgo.Guild) (*Handler, error) {
	h := Handler{discord: discord, guilds: guilds}

	h.Contests = newContestService(discord, client)
	h.Contests.addListener(&h)

	h.Pinger = newPinger(discord, h.Contests, &h)

	h.auth = newAuthService(db, discord, client)

	h.leaderboard = newLeaderboardService(discord, client, db, &h, WithRatingUpdateInterval(30*time.Minute))

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
	case "leaderboard":
		// This is only for testing purposes
		c := h.Contests.addContest("Leaderboard Test Contest", 69, uint32(time.Now().Unix()))
		err := h.leaderboard.sendLeaderboardMessage(m.GuildID, m.ChannelID, c)
		if err != nil {
			return fmt.Errorf("sending test leaderboard message: %w", err)
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
	ratingUpdates := h.leaderboard.startRatingUpdateCheck(c)
	for updated := range ratingUpdates {
		if updated {
			h.leaderboard.sendLeaderboardMessageAll(c)
		}
	}
}
