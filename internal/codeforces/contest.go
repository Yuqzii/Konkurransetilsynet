package codeforces

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

type contestFinishListener interface {
	onContestFinish(c *contest)
}

type contestProvider interface {
	getContests() []*contest
}

type contestService struct {
	discord *discordgo.Session
	client  api

	contestUpdateInterval time.Duration

	contests      []*contest
	endedContests map[uint32]struct{}
	mu            sync.RWMutex
	listeners     []contestFinishListener
}

type contestOption func(*contestService)

func newContestService(discord *discordgo.Session, client api, opts ...contestOption) *contestService {
	const defaultContestUpdateInterval time.Duration = 1 * time.Hour

	s := &contestService{
		discord:               discord,
		client:                client,
		contestUpdateInterval: defaultContestUpdateInterval,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithContestUpdateInterval(interval time.Duration) contestOption {
	return func(s *contestService) {
		s.contestUpdateInterval = interval
	}
}

func (s *contestService) StartContestUpdate(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := s.updateContests()
			if err != nil {
				log.Println("Failed to update upcoming contests:", err)
			}
		}
	}()
}

func (s *contestService) addListener(l contestFinishListener) {
	s.listeners = append(s.listeners, l)
}

func (s *contestService) listContests(m *discordgo.MessageCreate) error {
	embed := discordgo.MessageEmbed{
		Title:     "Upcoming Codeforces contests",
		URL:       "https://codeforces.com/contests",
		Color:     0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add embed for each contest
	s.mu.RLock()
	for _, contest := range s.contests {
		f := &discordgo.MessageEmbedField{
			Name:   contest.Name,
			Inline: false,
		}

		if contest.Phase == "BEFORE" {
			f.Value = fmt.Sprintf("Starts <t:%d:R>", contest.StartTimeSeconds)
		} else {
			f.Value = fmt.Sprintf("In progress, ends <t:%d:R>",
				contest.StartTimeSeconds+contest.DurationSeconds)
		}

		embed.Fields = append(embed.Fields, f)
	}
	s.mu.RUnlock()

	_, err := s.discord.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}

// Updates Service.contests with upcoming contests from the Codeforces API.
// Calls onContestEnd for any contests that have ended.
func (s *contestService) updateContests() error {
	// Check if any contest has ended.
	t := time.Now().Unix()
	for _, c := range s.contests {
		hasEnded := t >= int64(c.StartTimeSeconds)+int64(c.DurationSeconds)
		if hasEnded {
			s.endedContests[c.ID] = struct{}{}
		}
	}

	contests, err := s.client.getContests(context.TODO())
	if err != nil {
		return err
	}

	s.checkContestsFinish(contests)
	upcoming := filterUpcoming(contests)

	s.mu.Lock()
	s.contests = upcoming
	s.mu.Unlock()
	return nil
}

/* Checks all contests in the contests parameter against s.endedContests, and calls onContestFinish
 * for contests that are finished and exists in s.endedContests.
 */
func (s *contestService) checkContestsFinish(contests []*contest) {
	for _, c := range contests {
		s.mu.RLock()
		_, ok := s.endedContests[c.ID]
		s.mu.RUnlock()
		if ok && c.Phase == "FINISHED" {
			go s.onContestFinish(*c)

			s.mu.Lock()
			delete(s.endedContests, c.ID)
			s.mu.Unlock()
		}
	}
}

func (s *contestService) getContests() []*contest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*contest, len(s.contests))
	copy(res, s.contests)
	return res
}

func (s *contestService) addContest(name string, id uint32, startTime uint32) *contest {
	// Copy contests into new slice to avoid concurrency issues when writing
	s.mu.RLock()
	newContests := make([]*contest, len(s.contests))
	copy(newContests, s.contests)

	s.mu.RUnlock()

	// Find position to insert into slice to maintain sorted order
	i := sort.Search(len(newContests), func(i int) bool {
		return newContests[i].StartTimeSeconds >= startTime
	})

	newContest := &contest{
		ID:               id,
		Name:             name,
		StartTimeSeconds: startTime,
		DurationSeconds:  60,
		WebsiteURL:       "https://codeforces.com/contests",
	}
	// Insert new contest into slice at the correct position
	newContests = slices.Insert(newContests, i, newContest)
	// Update contests to our slice containing the new element
	s.mu.Lock()
	s.contests = newContests
	s.mu.Unlock()

	return newContest
}

func (s *contestService) addDebugContest(args []string, m *discordgo.MessageCreate) error {
	if len(args) != 5 {
		err := utils.UnknownCommand(s.discord, m)
		return err
	}

	name := args[2]

	startTime64, err := strconv.ParseUint(args[3], 10, 32)
	if err != nil {
		err = errors.Join(fmt.Errorf("parsing \"%s\" as int: %w", args[3], err),
			utils.UnknownCommand(s.discord, m))
		return err
	}
	startTime := uint32(startTime64)

	id64, err := strconv.ParseUint(args[4], 10, 32)
	if err != nil {
		err = errors.Join(fmt.Errorf("parsing \"%s\" as int: %w", args[3], err),
			utils.UnknownCommand(s.discord, m))
		return err
	}
	id := uint32(id64)

	s.addContest(name, id, startTime)

	err = s.listContests(m)
	if err != nil {
		return fmt.Errorf("listing contests: %w", err)
	}
	return nil
}

// Filters out contests that have ended and sorts the result
func filterUpcoming(contests []*contest) []*contest {
	// Find all current or future contests
	filtered := filterContests(contests, func(contest *contest) bool {
		return contest.Phase == "BEFORE" || contest.Phase == "CODING"
	})

	// Sort upcoming contests by starting time
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].StartTimeSeconds < filtered[j].StartTimeSeconds
	})

	return filtered
}

// Filters contests based on the f function argument
func filterContests(contests []*contest, f func(*contest) bool) (result []*contest) {
	for _, contest := range contests {
		if f(contest) {
			result = append(result, contest)
		}
	}
	return result
}

func (s *contestService) onContestFinish(c contest) {
	for _, l := range s.listeners {
		l.onContestFinish(&c)
	}
}
