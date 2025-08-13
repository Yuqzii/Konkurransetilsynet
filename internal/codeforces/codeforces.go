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
	contestUpdateInterval     time.Duration = 1 * time.Hour
	contestPingCheckInterval  time.Duration = 1 * time.Minute
	ratingUpdateCheckInterval time.Duration = 30 * time.Minute
)

type Service struct {
	db      *pgxpool.Pool
	discord *discordgo.Session
	guilds  []*discordgo.Guild

	contests   []contest
	contestsMu sync.RWMutex

	pingData  []pingData
	pingedIDs map[uint32]struct{}
	pingMu    sync.RWMutex

	lbGuildData []lbGuildData
	lbMu        sync.RWMutex
}

func New(db *pgxpool.Pool, discord *discordgo.Session, guilds []*discordgo.Guild) (*Service, error) {
	s := Service{db: db, discord: discord, guilds: guilds}

	if err := s.updatePingData(); err != nil {
		return nil, fmt.Errorf("initializing ping guild data: %w", err)
	}

	if err := s.updateLeaderboardGuildData(); err != nil {
		return nil, fmt.Errorf("initializing leaderboard guild data: %w", err)
	}

	return &s, nil
}

func (s *Service) HandleCommand(args []string, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "contests":
		if err := s.updateUpcoming(); err != nil {
			return fmt.Errorf("failed updating upcoming contests: %w", err)
		}

		err := s.listContests(m)
		if err != nil {
			return errors.Join(errors.New("listing future contests failed,"), err)
		}
	case "addDebugContest":
		err := s.addDebugContest(args, m)
		if err != nil {
			return errors.Join(errors.New("adding debug contest failed,"), err)
		}
	case "authenticate":
		err := s.authCommand(args, m)
		if err != nil {
			return fmt.Errorf("authentication command failed: %w", err)
		}
	default:
		err := utils.UnknownCommand(s.discord, m)
		return err
	}

	return nil
}
