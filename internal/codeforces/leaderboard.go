package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/yuqzii/konkurransetilsynet/internal/database"
)

const (
	channelName string = "cf-leaderboard"
)

type ratingChange struct {
	OldRating int `json:"oldRating"`
	NewRating int `json:"newRating"`
}

type lbGuildData struct {
	channels []string // Slice of channel IDs
	mu       sync.RWMutex
}

var guildData lbGuildData

func updateLeaderboardGuildData(s *discordgo.Session, guilds []*discordgo.Guild) error {
	channels, err := createChannelIfNotExist(s, channelName, guilds)
	if err != nil {
		return err
	}

	guildData.mu.Lock()
	guildData.channels = channels
	guildData.mu.Unlock()
	return nil
}

func getCodeforcesInGuild(guildID string, s *discordgo.Session) (result []string, err error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("getting guild from id %s: %w", guildID, err)
	}

	for _, member := range guild.Members {
		handle, err := database.GetConnectedCodeforces(member.User.ID)
		// ErrNoRows expected if the user has not connected their Codeforces
		if err != nil && err != pgx.ErrNoRows {
			return nil, fmt.Errorf("getting Codeforces handle of %s (%s): %w",
				member.User.ID, member.User.Username, err)
		}
		result = append(result, handle)
	}

	return result, nil
}

func getRating(handle string) (rating *ratingChange, err error) {
	apiStr := fmt.Sprintf("https://codeforces.com/api/user.rating?handle=%s", handle)
	res, err := http.Get(apiStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type apiReturn struct {
		Status  string         `json:"status"`
		Comment string         `json:"comment"`
		Result  []ratingChange `json:"result"`
	}
	var api apiReturn
	err = json.Unmarshal(body, &api)
	if api.Status == "FAILED" {
		return nil, errors.New(api.Comment)
	}

	return &api.Result[0], err
}
