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
	Status   string     `json:"status"`
	Contests []contest `json:"result"`
	Comment  string     `json:"comment,omitempty"`
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

var upcoming = contestList{}

func Init(s *discordgo.Session) error {
	startContestUpdate(1 * time.Hour)
	if err := updatePingChannels(s); err != nil {
		return err
	}
	startContestPingCheck(1 * time.Second, s)
	return nil
}

func HandleCodeforcesCommands(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "contests":
		err := listFutureContests(s, m)
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

		addDebugContest(args[2], startTime)
	default:
		err := messageCommands.UnknownCommand(s, m)
		return err
	}

	return nil
}

