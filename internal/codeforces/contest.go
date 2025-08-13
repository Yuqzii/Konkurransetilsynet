package codeforces

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

func (s *Service) listContests(m *discordgo.MessageCreate) error {
	embed := discordgo.MessageEmbed{
		Title:     "Upcoming Codeforces contests",
		URL:       "https://codeforces.com/contests",
		Color:     0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add embed for each contest
	s.contestsMu.RLock()
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
	s.contestsMu.RUnlock()

	_, err := s.discord.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}

func (s *Service) StartContestUpdate(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := s.updateUpcoming()
			if err != nil {
				log.Println("Failed to update upcoming contests:", err)
			}
		}
	}()
}

// Updates Service.contests with upcoming contests from the Codeforces API.
// Calls onContestEnd for any contests that have ended.
func (s *Service) updateUpcoming() error {
	// Check if any contest has ended.
	t := time.Now().Unix()
	for _, c := range s.contests {
		hasEnded := t >= int64(c.StartTimeSeconds)+int64(c.DurationSeconds)
		if hasEnded {
			go s.onContestEnd(c)
		}
	}

	contests, err := getContests()
	if err != nil {
		return err
	}

	contests, err = filterUpcoming(contests)
	if err != nil {
		return err
	}

	s.contestsMu.Lock()
	s.contests = contests
	s.contestsMu.Unlock()
	return nil
}

func (s *Service) addContest(name string, id uint32, startTime uint32) {
	// Copy contests into new slice to avoid concurrency issues when writing
	s.contestsMu.RLock()
	newContests := make([]contest, len(s.contests))
	copy(newContests, s.contests)

	s.contestsMu.RUnlock()

	// Find position to insert into slice to maintain sorted order
	i := sort.Search(len(newContests), func(i int) bool {
		return newContests[i].StartTimeSeconds >= startTime
	})
	// Insert new contest into slice at the correct position
	newContests = slices.Insert(newContests, i, contest{
		ID:               id,
		Name:             name,
		StartTimeSeconds: startTime,
		DurationSeconds:  60,
		WebsiteURL:       "https://codeforces.com/contests",
	})

	// Update contests to our slice containing the new element
	s.contestsMu.Lock()
	s.contests = newContests
	s.contestsMu.Unlock()
}

func (s *Service) addDebugContest(args []string, m *discordgo.MessageCreate) error {
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
	return nil
}

// Filters out contests that have ended and sorts the result
func filterUpcoming(contests []contest) ([]contest, error) {
	// Find all current or future contests
	filtered := filterContests(contests, func(contest *contest) bool {
		return contest.Phase == "BEFORE" || contest.Phase == "CODING"
	})

	// Sort upcoming contests by starting time
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].StartTimeSeconds < filtered[j].StartTimeSeconds
	})

	return filtered, nil
}

// Filters contests based on the f function argument
func filterContests(contests []contest, f func(*contest) bool) (result []contest) {
	for _, contest := range contests {
		if f(&contest) {
			result = append(result, contest)
		}
	}
	return result
}

func (s *Service) onContestEnd(c contest) {
	ratingUpdated := s.startRatingUpdateCheck(&c, ratingUpdateCheckInterval)
	for updated := range ratingUpdated {
		if updated {
			s.sendLeaderboardMessageAll(&c)
		}
	}
}
