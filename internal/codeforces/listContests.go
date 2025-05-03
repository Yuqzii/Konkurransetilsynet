package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ContestList struct {
	Status   string    `json:"status"`
	Contests []Contest `json:"result"`
	Comment  string    `json:"comment,omitempty"`
}

type Contest struct {
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

func listFutureContests(session *discordgo.Session, message *discordgo.MessageCreate) error {
	contests, err := getFromAPI()
	if err != nil {
		return err
	}

	if contests.Status == "FAILED" {
		return errors.New(contests.Comment)
	}

	embed := discordgo.MessageEmbed{
		Title:     "Upcoming Codeforces contests",
		URL:       "https://codeforces.com/contests",
		Color:     0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Find all current or future contests
	var upcoming []Contest
	for _, contest := range contests.Contests {
		if contest.Phase == "BEFORE" || contest.Phase == "CODING" {
			upcoming = append(upcoming, contest)
		}
	}

	// Sort upcoming contests by starting time
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].StartTimeSeconds < upcoming[j].StartTimeSeconds
	})

	// Add embed for each contest
	for _, contest := range upcoming {
		if contest.Phase == "BEFORE" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   contest.Name,
				Value:  fmt.Sprintf("Starts <t:%d:F>", contest.StartTimeSeconds),
				Inline: true,
			})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: contest.Name,
				Value: fmt.Sprintf("In progress, ends <t:%d:F>",
					contest.StartTimeSeconds+contest.DurationSeconds),
				Inline: true,
			})
		}
	}

	_, err = session.ChannelMessageSendEmbed(message.ChannelID, &embed)
	return err
}

func getFromAPI() (contests *ContestList, err error) {
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

	var contestList ContestList
	err = json.Unmarshal(body, &contestList)

	return &contestList, err
}
