package codeforces

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/yuqzii/konkurransetilsynet/internal"
)

type contestList struct {
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

type manager struct {
	upcomingContests []contest
	pingChannelIDs   []string
}

func NewManager(session *discordgo.Session) (*manager, error) {
	man := new(manager)
	man.startContestUpdate()
	man.startContestPingCheck(session)
	err := man.initPingChannel(session)
	if err != nil {
		return nil, err
	}
	return man, nil
}

func (man *manager) HandleCodeforcesCommands(args []string, session *discordgo.Session,
	message *discordgo.MessageCreate) {
	switch args[1] {
	case "contests":
		err := man.listFutureContests(session, message)
		if err != nil {
			log.Println("Listing future Codeforces contests failed, ", err)
		}
	case "addDebugContest":
		if len(args) != 4 {
			err := messageCommands.UnknownCommand(session, message)
			if err != nil {
				log.Println("UnknownCommand failed, ", err)
			}
			return
		}

		startTime, err := strconv.Atoi(args[3])
		if err != nil {
			err = messageCommands.UnknownCommand(session, message)
			if err != nil {
				log.Println("UnknownCommand failed, ", err)
			}
			return
		}

		man.addDebugContest(args[2], startTime)
	}
}

func (man *manager) addDebugContest(name string, startTime int) {
	man.upcomingContests = append(man.upcomingContests, contest{
		Name:             name,
		StartTimeSeconds: startTime,
		Pinged:           false,
	})
}

// Start goroutine that updates upcomingContests
func (man *manager) startContestUpdate() {
	go func() {
		for {
			// Update once every hour
			time.Sleep(1 * time.Hour)

			err := man.updateUpcomingContests()
			if err != nil {
				log.Println("Updating upcoming contests failed,", err)
			}
		}
	}()
}

func (man *manager) updateUpcomingContests() error {
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

	man.upcomingContests = upcoming
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
