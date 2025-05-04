package codeforces

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

const pingTime int = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func (man *manager) startContestPingCheck() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)

			man.checkContestPing()
		}
	}()
}

func (man *manager) checkContestPing() {
	for _, contest := range man.upcomingContests {
		if contest.StartTimeSeconds - int(time.Now().Unix()) <= pingTime && !contest.Pinged {
			contestPing(&contest)
		}
	}
}

func contestPing(contest *contest) {
	contest.Pinged = true
}

func (man *manager) initPingChannel(session *discordgo.Session) error {
	const channelName string = "contest-pings"

	// Clear pingChannelIDs of possible existing IDs
	man.pingChannelIDs = nil

	for _, guild := range session.State.Guilds {
		channels, err := session.GuildChannels(guild.ID)
		if err != nil {
			return err
		}

		pingChannel := ""
		for _, channel := range channels {
			// Skip non-text channels
			if channel.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			if channel.Name == channelName {
				pingChannel = channel.ID
				break
			}
		}

		if pingChannel == "" {
			// The server does not have a ping channel
			newChannel, err := session.GuildChannelCreate(guild.ID, channelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				return err
			}
			
			pingChannel = newChannel.ID
		}

		man.pingChannelIDs = append(man.pingChannelIDs, pingChannel)
	}

	return nil
}
