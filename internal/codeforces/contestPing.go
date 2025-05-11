package codeforces

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type pingData struct {
	channel string
	role    string
}

// !! Lock mutex when accessing
type pingDataList struct {
	list []pingData
	mu   sync.RWMutex
}

var pingList = pingDataList{}

const pingTime int = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func startContestPingCheck(contests *contestList, interval time.Duration, session *discordgo.Session) {
	go func() {
		for {
			time.Sleep(interval)
			checkContestPing(contests, session)
		}
	}()
}

func checkContestPing(contests *contestList, session *discordgo.Session) {
	contests.mu.RLock()
	defer contests.mu.RUnlock()

	curTime := int(time.Now().Unix())
	for i, contest := range contests.contests {
		shouldPing := contest.StartTimeSeconds-curTime <= pingTime
		if shouldPing && !contest.Pinged {
			// Unlock reading to allow contestPing to write
			contests.mu.RUnlock()

			log.Println("Pinging contest", contest.Name)
			err := contestPing(contests, i, session)
			if err != nil {
				log.Println("Automatic contest ping failed:", err)
			}
		}
		// Lock again to ensure safe access on next iteration
		contests.mu.RLock()
	}
}

func contestPing(contests *contestList, idx int, session *discordgo.Session) error {
	contests.mu.Lock()
	contests.contests[idx].Pinged = true
	contests.mu.Unlock()

	contests.mu.RLock()
	defer contests.mu.RUnlock()
	pingList.mu.RLock()
	defer pingList.mu.RUnlock()
	for _, data := range pingList.list {
		// TODO: Find role id belonging to each server and ping it
		_, err := session.ChannelMessageSend(data.channel,
			fmt.Sprintf("<@&%s> **%s** is starting <t:%d:R>",
				data.role, contests.contests[idx].Name, contests.contests[idx].StartTimeSeconds))
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePingData(s *discordgo.Session) error {
	data, err := getPingData(s)
	if err != nil {
		return err
	}

	pingList.mu.Lock()
	defer pingList.mu.Unlock()

	pingList.list = data
	return nil
}

func getPingData(s *discordgo.Session) ([]pingData, error) {
	const channelName string = "contest-pings"
	const roleName string = "Contest Ping"

	var result []pingData

	for _, guild := range s.State.Guilds {
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			return nil, err
		}
		pingChannel := ""
		// Try to find a channel with the name
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
		// Create ping channel if server does not have one
		if pingChannel == "" {
			newChannel, err := s.GuildChannelCreate(guild.ID, channelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				return nil, err
			}

			log.Println("Created ping channel,", newChannel.ID)
			pingChannel = newChannel.ID
		}

		roles, err := s.GuildRoles(guild.ID)
		if err != nil {
			return nil, err
		}
		pingRole := ""
		// Try to find role with correct name
		for _, role := range roles {
			if role.Name == roleName {
				pingRole = role.ID
				break
			}
		}
		// Create ping role if server does not have one
		if pingRole == "" {
			newRole, err := s.GuildRoleCreate(guild.ID, &discordgo.RoleParams{
				Name:        roleName,
				Mentionable: boolPointer(true),
			})
			if err != nil {
				return nil, errors.Join(errors.New("failed to create ping role:"), err)
			}

			log.Println("Created ping role:", newRole.ID)
			pingRole = newRole.ID
		}

		result = append(result, pingData{pingChannel, pingRole})
		log.Println("Found ping channel,", pingChannel)
	}

	return result, nil
}

// GuildRoleCreate wants a *bool, which cannot be made easily without a helper function like this
func boolPointer(b bool) *bool {
	return &b
}
