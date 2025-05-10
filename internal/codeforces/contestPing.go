package codeforces

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// !! Lock mutex when accessing
type pingChannelIDs struct {
	list []string
	mu   sync.RWMutex
}

var pingChannels = pingChannelIDs{}

const pingTime int = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func startContestPingCheck(interval time.Duration, session *discordgo.Session) {
	go func() {
		for {
			time.Sleep(interval)
			checkContestPing(session)
		}
	}()
}

func checkContestPing(session *discordgo.Session) {
	upcoming.mu.RLock()
	defer upcoming.mu.RUnlock()

	curTime := int(time.Now().Unix())
	for i, contest := range upcoming.contests {
		shouldPing := contest.StartTimeSeconds-curTime <= pingTime
		if shouldPing && !contest.Pinged {
			// Unlock reading to allow contestPing to write
			upcoming.mu.RUnlock() 

			log.Println("Pinging contest", contest.Name)
			err := contestPing(i, session)
			if err != nil {
				log.Println("Automatic contest ping failed:", err)
			}
		}
		// Lock again to ensure safe access on next iteration
		upcoming.mu.RLock()
	}
}

func contestPing(idx int, session *discordgo.Session) error {
	upcoming.mu.Lock()
	upcoming.contests[idx].Pinged = true
	upcoming.mu.Unlock()

	upcoming.mu.RLock()
	defer upcoming.mu.RUnlock()
	pingChannels.mu.RLock()
	defer pingChannels.mu.RUnlock()
	for _, channel := range pingChannels.list {
		// !!!! UPDATE FOR PRODUCTION, using temporary hardcoded role id
		_, err := session.ChannelMessageSend(channel,
			fmt.Sprint("@<role> ", upcoming.contests[idx].Name,
				fmt.Sprintf(" is starting <t:%d:R>", upcoming.contests[idx].StartTimeSeconds)))
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePingChannels(s *discordgo.Session) error {
	ids, err := getPingChannels(s)
	if err != nil {
		return err
	}

	pingChannels.mu.Lock()
	defer pingChannels.mu.Unlock()

	pingChannels.list = ids
	return nil
}

func getPingChannels(s *discordgo.Session) ([]string, error) {
	const channelName string = "contest-pings"

	var result []string

	for _, guild := range s.State.Guilds {
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			return nil, err
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
			newChannel, err := s.GuildChannelCreate(guild.ID, channelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				return nil, err
			}

			log.Println("Created ping channel,", newChannel.ID)
			pingChannel = newChannel.ID
		}

		result = append(result, pingChannel)
		log.Println("Found ping channel,", pingChannel)
	}

	return result, nil
}
