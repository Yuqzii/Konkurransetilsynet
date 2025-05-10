package codeforces

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/yuqzii/konkurransetilsynet/internal"
)

// The return object from the Codeforces API
type contestListAPI struct {
	Status   string    `json:"status"`
	Contests []contest `json:"result"`
	Comment  string    `json:"comment,omitempty"`
}

type contest struct {
	Pinged                bool
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Phase                 string `json:"phase"`
	Frozen                bool   `json:"frozen"`
	DurationSeconds       int    `json:"durationSeconds"`
	Description           string `json:"description,omitempty"`
	Difficulty            int    `json:"difficulty,omitempty"`
	Kind                  string `json:"kind,omitempty"`
	Season                string `json:"season,omitempty"`
	StartTimeSeconds      int    `json:"startTimeSeconds,omitempty"`
	RelativeTimeSeconds   int    `json:"relativeTimeSeconds,omitempty"`
	PreparedBy            string `json:"preparedBy,omitempty"`
	Country               string `json:"country,omitempty"`
	City                  string `json:"city,omitempty"`
	IcpcRegion            string `json:"icpcRegion,omitempty"`
	WebsiteURL            string `json:"websiteUrl,omitempty"`
	FreezeDurationSeconds int    `json:"freezeDurationSeconds,omitempty"`
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
	if err := updatePingChannels(s); err != nil {
		return err
	}
	startContestPingCheck(&upcoming, 1*time.Second, s)
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
			return errors.Join(errors.New("Listing future contests failed"), err)
		}
	case "addDebugContest":
		if len(args) != 4 {
			err := messageCommands.UnknownCommand(s, m)
			return err
		}

		startTime, err := strconv.Atoi(args[3])
		if err != nil {
			err = errors.Join(err, messageCommands.UnknownCommand(s, m))
			return err
		}

		addContest(&upcoming, args[2], startTime)
	default:
		err := messageCommands.UnknownCommand(s, m)
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

func addContest(contests *contestList, name string, startTime int) {
	// Create new list of contests and the contest to it
	contests.mu.RLock()
	newContests := make([]contest, len(contests.contests))
	copy(contests.contests, newContests)
	contests.mu.RUnlock()

	newContests = append(newContests, contest{
		Name:             name,
		StartTimeSeconds: startTime,
		Pinged:           false,
	})

	// Update the shared value with our new one
	contests.mu.Lock()
	contests.contests = newContests
	contests.mu.Unlock()
}

// Updates upcoming contests with data from the Codeforces API
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
