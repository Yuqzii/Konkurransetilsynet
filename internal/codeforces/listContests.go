package codeforces

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"fmt"
)

type ContestList struct {
	Status string   `json:"status"`
	Contests []Contest `json:"result"`
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

func ListFutureContests(session *discordgo.Session, message *discordgo.MessageCreate) error {
	contests, err := GetFromAPI()
	if err != nil {
		return err
	}

	// Find all contests that are not yet finished
	var futureContests []Contest
	for _, contest := range contests.Contests {
		if contest.Phase == "BEFORE" || contest.Phase == "CODING" {
			futureContests = append(futureContests, contest)
		}
	}

	embed := discordgo.MessageEmbed{
		Title: "Upcoming Codeforces contests",
		Description: "This is a list of all upcoming Codeforces contests",
		Color: 0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add an embed field for each upcoming contest
	for _, contest := range futureContests {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: contest.Name,
			Value: fmt.Sprint(contest.StartTimeSeconds),
			Inline: true,
		})
	}

	session.ChannelMessageSendEmbed(message.ChannelID, &embed)
	return nil
}

func GetFromAPI() (*ContestList, error) {
	res, err := http.Get("https://codeforces.com/api/contest.list")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response ContestList
	json.Unmarshal(body, &response)

	return &response, nil
}
