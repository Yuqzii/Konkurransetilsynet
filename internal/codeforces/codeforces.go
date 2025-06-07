package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	utilCommands "github.com/yuqzii/konkurransetilsynet/internal/utilCommands"
)

// The return object from the Codeforces API
type contestListAPI struct {
	Status   string    `json:"status"`
	Contests []contest `json:"result"`
	Comment  string    `json:"comment,omitempty"`
}

type contest struct {
	ID                    uint32 `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Phase                 string `json:"phase"`
	Frozen                bool   `json:"frozen"`
	DurationSeconds       uint32 `json:"durationSeconds"`
	Description           string `json:"description,omitempty"`
	Difficulty            uint8  `json:"difficulty,omitempty"`
	Kind                  string `json:"kind,omitempty"`
	Season                string `json:"season,omitempty"`
	StartTimeSeconds      uint32 `json:"startTimeSeconds,omitempty"`
	RelativeTimeSeconds   int32  `json:"relativeTimeSeconds,omitempty"`
	PreparedBy            string `json:"preparedBy,omitempty"`
	Country               string `json:"country,omitempty"`
	City                  string `json:"city,omitempty"`
	IcpcRegion            string `json:"icpcRegion,omitempty"`
	WebsiteURL            string `json:"websiteUrl,omitempty"`
	FreezeDurationSeconds uint32 `json:"freezeDurationSeconds,omitempty"`
}

// !! Lock mutex when accessing
type contestList struct {
	contests []contest
	mu       sync.RWMutex
}

// Only access this variable when passing as a argument from a top-level function (Init and HandleCommands)
var upcoming = contestList{}

func Init(s *discordgo.Session) error {
	startContestUpdate(&upcoming, 1*time.Hour)
	if err := updatePingData(s); err != nil {
		return err
	}
	startContestPingCheck(&upcoming, 1*time.Minute, s)
	return nil
}

func HandleCodeforcesCommands(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "contests":
		if err := updateUpcoming(&upcoming); err != nil {
			return err
		}

		err := listContests(&upcoming, s, m)
		if err != nil {
			return errors.Join(errors.New("listing future contests failed,"), err)
		}
	case "addDebugContest":
		err := addDebugContestCommand(args, s, m)
		if err != nil {
			return errors.Join(errors.New("adding debug contest failed,"), err)
		}
	case "authenticate":
		err := authCommand(args, s, m)
		if err != nil {
			return fmt.Errorf("authentication command failed: %w", err)
		}
	default:
		err := utilCommands.UnknownCommand(s, m)
		return err
	}

	return nil
}

// Start goroutine that updates upcomingContests
func startContestUpdate(listToUpdate *contestList, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := updateUpcoming(listToUpdate)
			if err != nil {
				log.Println("Failed to update upcoming contests:", err)
			}
		}
	}()
}

func addDebugContestCommand(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(args) != 5 {
		err := utilCommands.UnknownCommand(s, m)
		return err
	}

	startTime64, err := strconv.ParseUint(args[3], 10, 32)
	if err != nil {
		err = errors.Join(err, utilCommands.UnknownCommand(s, m))
		return err
	}
	startTime := uint32(startTime64)

	id64, err := strconv.ParseUint(args[4], 10, 32)
	if err != nil {
		err = errors.Join(err, utilCommands.UnknownCommand(s, m))
		return err
	}
	id := uint32(id64)

	addContest(&upcoming, id, args[2], startTime)
	return nil
}

func addContest(contests *contestList, id uint32, name string, startTime uint32) {
	// Copy contests into new slice to avoid concurrency issues when writing
	contests.mu.RLock()
	newContests := make([]contest, len(contests.contests))
	copy(newContests, contests.contests)
	contests.mu.RUnlock()

	// Find position to insert into slice to maintain sorted order
	i := sort.Search(len(newContests), func(i int) bool {
		return newContests[i].StartTimeSeconds >= startTime
	})
	// Insert new contest into slice at the correct position
	newContests = slices.Insert(newContests, i, contest{
		ID:               id,
		Name:             name,
		StartTimeSeconds: startTime,
	})

	// Update contests to our slice containing the new element
	contests.mu.Lock()
	contests.contests = newContests
	contests.mu.Unlock()
}

// Updates listToUpdate with upcoming contests from the Codeforces API
func updateUpcoming(listToUpdate *contestList) error {
	contests, err := getContests()
	if err != nil {
		return err
	}

	contests, err = filterUpcoming(contests)
	if err != nil {
		return err
	}

	listToUpdate.mu.Lock()
	listToUpdate.contests = contests
	listToUpdate.mu.Unlock()
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

// Gets all contests from the Codeforces API
func getContests() (contests []contest, err error) {
	res, err := http.Get("https://codeforces.com/api/contest.list")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = res.Body.Close()
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var contestList contestListAPI
	err = json.Unmarshal(body, &contestList)

	if contestList.Status == "FAILED" {
		return nil, errors.New(contestList.Comment)
	}

	return contestList.Contests, err
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
