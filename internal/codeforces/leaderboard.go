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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/yuqzii/konkurransetilsynet/internal/database"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

const (
	channelName string = "cf-leaderboard"
)

type ratingChangeAPIReturn struct {
	Status  string         `json:"status"`
	Comment string         `json:"comment"`
	Result  []ratingChange `json:"result"`
}

type ratingChange struct {
	Handle    string `json:"handle"`
	OldRating int    `json:"oldRating"`
	NewRating int    `json:"newRating"`
	discordID string
}

type lbGuildData struct {
	guildID   string
	channelID string
}

type lbGuildDataList struct {
	data []lbGuildData
	mu   sync.RWMutex
}

var guildData lbGuildDataList

func updateLeaderboardGuildData(s *discordgo.Session, guilds []*discordgo.Guild) error {
	channels, err := utils.CreateChannelIfNotExist(s, channelName, guilds)
	if err != nil {
		return err
	}

	var newData []lbGuildData
	for i := range channels {
		newData = append(newData, lbGuildData{guilds[i].ID, channels[i]})
	}

	guildData.mu.Lock()
	guildData.data = newData
	guildData.mu.Unlock()
	return nil
}

// @abstract	Sends a leaderboard message for every guild the bot is in.
func sendLeaderboardMessageAll(s *discordgo.Session, c *contest) {
	guildData.mu.RLock()
	defer guildData.mu.RUnlock()

	for _, data := range guildData.data {
		go func() {
			err := sendLeaderboardMessage(data.guildID, data.channelID, c, s)
			if err != nil {
				log.Printf("Error sending leaderboard message to all guilds (guild %s): %s", data.guildID, err)
			}
		}()
	}
}

func sendLeaderboardMessage(guildID string, channelID string, c *contest, s *discordgo.Session) error {
	ratings, err := getRatingsInGuild(guildID, s)
	if err != nil {
		return fmt.Errorf("getting ratings in guild %s: %w", guildID, err)
	}
	// Sort by new rating descending
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].NewRating > ratings[j].NewRating
	})

	guild, err := s.State.Guild(guildID)
	if err != nil {
		return fmt.Errorf("getting guild of ID %s: %w", guildID, err)
	}
	messageStr := fmt.Sprintf("## %s Codeforces leaderboard after [%s](%s)", guild.Name, c.Name, c.WebsiteURL)
	for i, rating := range ratings {
		messageStr += fmt.Sprintf("\n%d. <@%s> (%s): %d", i+1, rating.discordID, rating.Handle, rating.NewRating)
	}

	msgData := discordgo.MessageSend{
		Content: messageStr,
		Flags:   discordgo.MessageFlagsSuppressNotifications,
	}
	_, err = s.ChannelMessageSendComplex(channelID, &msgData)
	if err != nil {
		return fmt.Errorf("sending leaderboard message: %w", err)
	}

	return nil
}

// Sends true to the returned channel when the ratings have been updated
func startRatingUpdateCheck(c *contest, interval time.Duration) <-chan bool {
	updatedChan := make(chan bool)
	go func() {
		errCnt := 0
		const maxErrs uint8 = 3

		for {
			time.Sleep(interval)
			updated, err := hasUpdatedRating(c)
			if err != nil {
				errCnt++
				log.Printf("Failed to check Codeforces rating update (attempt %d of %d): %s", errCnt, maxErrs, err)
				if errCnt == int(maxErrs) {
					log.Println("Stopping Codeforces rating update check.")
					return
				}
			}
			if updated {
				updatedChan <- true
				close(updatedChan)
				return
			}
		}
	}()
	return updatedChan
}

func hasUpdatedRating(c *contest) (updated bool, err error) {
	apiStr := fmt.Sprintf("https://codeforces.com/api/contest.ratingChanges?contestId=%d", c.ID)
	res, err := http.Get(apiStr)
	if err != nil {
		return false, fmt.Errorf("getting rating change from Codeforces api: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var api ratingChangeAPIReturn
	err = json.Unmarshal(body, &api)
	if api.Status == "FAILED" {
		return false, errors.New(api.Comment)
	}

	// Codeforces returns an empty result before the ratings have updated
	return len(api.Result) != 0, err
}

func getRatingsInGuild(guildID string, s *discordgo.Session) ([]*ratingChange, error) {
	handles, ids, err := getCodeforcesInGuild(guildID, s)
	if err != nil {
		return nil, fmt.Errorf("getting Codeforces handles in %s: %w", guildID, err)
	}

	if len(handles) == 0 {
		return nil, errors.New("no connected Codeforces in the guild")
	}

	ratingChan := make(chan *ratingChange)
	var wg sync.WaitGroup
	for i := range handles {
		wg.Add(1)
		go func() {
			defer wg.Done()

			rating, err := getRating(handles[i])
			if err != nil {
				log.Printf("Getting Codeforces rating from handle %s failed: %s", handles[i], err)
				return
			}
			rating.discordID = ids[i]
			ratingChan <- rating
		}()
	}

	go func() {
		wg.Wait()
		close(ratingChan)
	}()

	var ratings []*ratingChange
	for rating := range ratingChan {
		ratings = append(ratings, rating)
	}

	return ratings, nil
}

func getCodeforcesInGuild(guildID string, s *discordgo.Session) (result []string, discordIDs []string, err error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting guild from id %s: %w", guildID, err)
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
			discordIDs = append(discordIDs, id)
		}

		return nil
	}

	for _, member := range guild.Members {
		err = getCodeforcesFromID(member.User.ID)
		if err != nil {
			return nil, nil, err
		}
	}

	// guild.Members does for some reason not include owner, this includes the owner in the leaderboard
	err = getCodeforcesFromID(guild.OwnerID)
	if err != nil {
		return nil, nil, err
	}

	return result, discordIDs, nil
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

	var api ratingChangeAPIReturn
	err = json.Unmarshal(body, &api)
	if api.Status == "FAILED" {
		return nil, errors.New(api.Comment)
	}

	return &api.Result[len(api.Result)-1], err
}
