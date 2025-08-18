package codeforces

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

const (
	lbChannelName string = "cf-leaderboard"
)

type lbGuildData struct {
	guildID   string
	channelID string
}

type lbService struct {
	discord *discordgo.Session
	client  api
	db      Repository
	guilds  guildProvider

	ratingUpdateInterval time.Duration

	data []lbGuildData
	mu   sync.RWMutex
}

type lbOption func(*lbService)

func newLeaderboardService(discord *discordgo.Session, client api, db Repository,
	guilds guildProvider, opts ...lbOption) *lbService {

	const defaultRatingUpdateInterval time.Duration = 30 * time.Minute

	s := &lbService{
		discord:              discord,
		client:               client,
		db:                   db,
		guilds:               guilds,
		ratingUpdateInterval: defaultRatingUpdateInterval,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithRatingUpdateInterval(interval time.Duration) lbOption {
	return func(s *lbService) {
		s.ratingUpdateInterval = interval
	}
}

// Sends a leaderboard message for every guild the bot is in.
func (s *lbService) sendLeaderboardMessageAll(c *contest) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, data := range s.data {
		go func() {
			err := s.sendLeaderboardMessage(data.guildID, data.channelID, c)
			if err != nil {
				log.Printf("Error when sending leaderboard message to all guilds (guild %s): %s",
					data.guildID, err)
			}
		}()
	}
}

func (s *lbService) sendLeaderboardMessage(guildID string, channelID string, c *contest) error {
	ratings, err := s.getRatingsInGuild(guildID)
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
func (s *lbService) startRatingUpdateCheck(c *contest) <-chan bool {
	updatedChan := make(chan bool)
	go func() {
		errCnt := 0
		const maxErrs uint8 = 3

		for {
			time.Sleep(s.ratingUpdateInterval)
			updated, err := s.client.hasUpdatedRating(context.TODO(), c)
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

func (s *lbService) getRatingsInGuild(guildID string) ([]*ratingChange, error) {
	handles, ids, err := s.getCodeforcesInGuild(guildID)
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

			rating, err := s.client.getRating(context.TODO(), handles[i])
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

func (s *lbService) getCodeforcesInGuild(guildID string) (result []string, discordIDs []string, err error) {
	guild, err := s.discord.Guild(guildID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting guild from id %s: %w", guildID, err)
	}

	// Helper lambda to avoid duplicate code in member loop and owner
	getCodeforcesFromID := func(id string) error {
		handle, err := s.db.GetConnectedCodeforces(context.TODO(), id)
		if err != nil && !errors.Is(err, ErrUserNotConnected) {
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

func (s *lbService) updateData() error {
	guilds := s.guilds.getGuilds()
	channels, err := utils.CreateChannelIfNotExist(s.discord, lbChannelName, guilds)
	if err != nil {
		return err
	}

	var newData []lbGuildData
	for i := range guilds {
		newData = append(newData, lbGuildData{guilds[i].ID, channels[i]})
	}

	s.mu.Lock()
	s.data = newData
	s.mu.Unlock()
	return nil
}
