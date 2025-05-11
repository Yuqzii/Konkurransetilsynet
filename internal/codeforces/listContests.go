package codeforces

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func listContests(contests *contestList, s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := discordgo.MessageEmbed{
		Title:     "Upcoming Codeforces contests",
		URL:       "https://codeforces.com/contests",
		Color:     0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add embed for each contest
	contests.mu.RLock()
	for _, contest := range contests.contests {
		if contest.Phase == "BEFORE" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   contest.Name,
				Value:  fmt.Sprintf("Starts <t:%d:R>", contest.StartTimeSeconds),
				Inline: true,
			})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: contest.Name,
				Value: fmt.Sprintf("In progress, ends <t:%d:R>",
					contest.StartTimeSeconds+contest.DurationSeconds),
				Inline: true,
			})
		}
	}
	contests.mu.RUnlock()

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}
