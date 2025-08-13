package codeforces

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/yuqzii/konkurransetilsynet/internal/database"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

const (
	lbChannelName string = "cf-leaderboard"
)

type lbGuildData struct {
	guildID   string
	channelID string
}

type lbGuildDataList struct {
	data []lbGuildData
	mu   sync.RWMutex
}

// @abstract	Sends a leaderboard message for every guild the bot is in.
func (s *Service) sendLeaderboardMessageAll(c *contest) {
	s.lbMu.RLock()
	defer s.lbMu.RUnlock()

	for i := range s.lbGuildData {
		go func() {
			err := s.sendLeaderboardMessage(i, c)
			if err != nil {
				log.Printf("Error when sending leaderboard message to all guilds (guild %s): %s",
					s.lbGuildData[i].guildID, err)
			}
		}()
	}
}

func (s *Service) sendLeaderboardMessage(idx int, c *contest) error {
	guildID := s.lbGuildData[idx].guildID
	channelID := s.lbGuildData[idx].channelID

	ratings, err := getRatingsInGuild(guildID, s.discord)
	if err != nil {
		return fmt.Errorf("getting ratings in guild %s: %w", guildID, err)
	}
	// Sort by new rating descending
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].NewRating > ratings[j].NewRating
	})

	guild, err := s.discord.State.Guild(guildID)
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
	_, err = s.discord.ChannelMessageSendComplex(channelID, &msgData)
	if err != nil {
		return fmt.Errorf("sending leaderboard message: %w", err)
	}

	return nil
}

// Sends true to the returned channel when the ratings have been updated
func (s *Service) startRatingUpdateCheck(c *contest, interval time.Duration) <-chan bool {
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

func (s *Service) updateLeaderboardGuildData() error {
	channels, err := utils.CreateChannelIfNotExist(s.discord, lbChannelName, s.guilds)
	if err != nil {
		return err
	}

	var newData []lbGuildData
	for i := range s.guilds {
		newData = append(newData, lbGuildData{s.guilds[i].ID, channels[i]})
	}

	s.lbMu.Lock()
	s.lbGuildData = newData
	s.lbMu.Unlock()
	return nil
}
