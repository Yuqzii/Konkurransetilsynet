package codeforces

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

const pingTime int = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func (man *manager) startContestPingCheck(session *discordgo.Session) {
	go func() {
		for {
			time.Sleep(1 * time.Minute)

			man.checkContestPing(session)
		}
	}()
}

func (man *manager) checkContestPing(session *discordgo.Session) {
	curTime := int(time.Now().Unix())
	for _, contest := range man.upcomingContests {
		shouldPing := contest.StartTimeSeconds-curTime <= pingTime
		if shouldPing && !contest.Pinged {
			log.Println("Pinging contest", contest.Name)
			err := man.contestPing(contest, session)
			if err != nil {
				log.Println("Automatic contest ping failed, ", err)
			}
		}
	}
}

func (man *manager) contestPing(contest *contest, session *discordgo.Session) error {
	man.mu.Lock()
	defer man.mu.Unlock()
	contest.Pinged = true

	for _, channel := range man.pingChannelIDs {
		// !!!! UPDATE FOR PRODUCTION, using temporary hardcoded role id
		_, err := session.ChannelMessageSend(channel,
			fmt.Sprint("<@&1369025298359648358>", contest.Name,
				fmt.Sprintf("is starting <t:%d:R>", contest.StartTimeSeconds)))
		if err != nil {
			return err
		}
	}

	return nil
}

func (man *manager) InitPingChannel(session *discordgo.Session) error {
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

			log.Println("Created ping channel, ", newChannel.ID)
			pingChannel = newChannel.ID
		}

		man.pingChannelIDs = append(man.pingChannelIDs, pingChannel)
		log.Println("Found ping channel, ", pingChannel)
	}

	return nil
}
