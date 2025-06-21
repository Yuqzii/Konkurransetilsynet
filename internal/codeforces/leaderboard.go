package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/yuqzii/konkurransetilsynet/internal/database"
)

const (
	channelName string = "cf-leaderboard"
)

type ratingChange struct {
	Handle    string `json:"handle"`
	OldRating int    `json:"oldRating"`
	NewRating int    `json:"newRating"`
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

func sendLeaderboardMessage(guildID string, channelID string, s *discordgo.Session) error {
	handles, err := getCodeforcesInGuild(guildID, s)
	if err != nil {
		return fmt.Errorf("getting codeforces handles in %s: %w", guildID, err)
	}

	if len(handles) == 0 {
		log.Printf("No connected Codeforces in guild %s. Not sending leaderboard message.", guildID)
		return nil
	}

	ratingChan := make(chan *ratingChange)
	for _, handle := range handles {
		go func() {
			rating, err := getRating(handle)
			if err != nil {
				log.Printf("Getting Codeforces rating from handle %s failed: %s", handle, err)
				return
			}
			ratingChan <- rating
		}()
	}

	var ratings []*ratingChange
	for range len(handles) {
		rating := <-ratingChan
		ratings = append(ratings, rating)
	}

	// Sort by new rating
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].NewRating > ratings[j].NewRating
	})

	messageStr := ""
	for i, rating := range ratings {
		messageStr += fmt.Sprintf("%d. %s (%d)\n", i+1, rating.Handle, rating.NewRating)
	}

	_, err = s.ChannelMessageSend(channelID, messageStr)
	if err != nil {
		return fmt.Errorf("sending leaderboard message: %w", err)
	}

	return nil
}

func getCodeforcesInGuild(guildID string, s *discordgo.Session) (result []string, err error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("getting guild from id %s: %w", guildID, err)
	}

	// Helper lambda to avoid duplicate code in member loop and owner
	getCodeforcesFromID := func(id string) error {
		handle, err := database.GetConnectedCodeforces(id)
		// ErrNoRows expected if the user has not connected their Codeforces
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("getting Codeforces handle of %s: %w", id, err)
		}

		if handle != "" {
			result = append(result, handle)
		}

		return nil
	}

	for _, member := range guild.Members {
		err = getCodeforcesFromID(member.User.ID)
		if err != nil {
			return nil, err
		}
	}

	// guild.Members does for some reason not include owner, this includes the owner in the leaderboard
	err = getCodeforcesFromID(guild.OwnerID)
	if err != nil {
		return nil, err
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
