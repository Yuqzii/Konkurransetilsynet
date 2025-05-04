package codeforces

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (manager *manager) listFutureContests(session *discordgo.Session,
	message *discordgo.MessageCreate) error {

	err := manager.updateUpcomingContests()
	if err != nil {
		return err
	}

	embed := discordgo.MessageEmbed{
		Title:     "Upcoming Codeforces contests",
		URL:       "https://codeforces.com/contests",
		Color:     0x50e6ac,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add embed for each contest
	for _, contest := range manager.upcomingContests {
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

	_, err = session.ChannelMessageSendEmbed(message.ChannelID, &embed)
	return err
}

