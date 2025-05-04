package codeforces

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
)

type contestList struct {
	Status   string    `json:"status"`
	Contests []contest `json:"result"`
	Comment  string    `json:"comment,omitempty"`
}

type contest struct {
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

type Manager struct {
	upcomingContests []contest
}

func (manager *Manager) HandleCodeforcesCommands(args []string, session *discordgo.Session,
	message *discordgo.MessageCreate) {
	if args[1] == "contests" {
		err := manager.listFutureContests(session, message)
		if err != nil {
			log.Println("Listing future Codeforces contests failed, ", err)
		}
	}
}

func (manager *Manager) updateUpcomingContests() error {
	contests, err := getContests()
	if err != nil {
		return err
	}

	if contests.Status == "FAILED" {
		return errors.New(contests.Comment)
	}

	// Find all current or future contests
	var upcoming []contest
	for _, contest := range contests.Contests {
		if contest.Phase == "BEFORE" || contest.Phase == "CODING" {
			upcoming = append(upcoming, contest)
		}
	}

	// Sort upcoming contests by starting time
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].StartTimeSeconds < upcoming[j].StartTimeSeconds
	})

	manager.upcomingContests = upcoming
	return nil
}

func getContests() (contests *contestList, err error) {
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

	var contestList contestList
	err = json.Unmarshal(body, &contestList)

	return &contestList, err
}
