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
var pingedIDs = make(map[uint32]struct{}) // Set for pinged contests

const pingTime uint32 = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func startContestPingCheck(contests *contestList, interval time.Duration, session *discordgo.Session) {
	go func() {
		for {
			time.Sleep(interval)
			err := checkContestPing(contests, session)
			if err != nil {
				log.Println("Automatic contest ping failed:", err)
			}
		}
	}()
}

func checkContestPing(contests *contestList, session *discordgo.Session) error {
	contests.mu.RLock()
	defer contests.mu.RUnlock()

	curTime := uint32(time.Now().Unix())
	for i, contest := range contests.contests {
		shouldPing := contest.StartTimeSeconds-curTime <= pingTime
		_, isPinged := pingedIDs[contest.ID]
		if shouldPing && !isPinged {
			err := contestPing(contests, i, session)
			if err != nil {
				return errors.Join(errors.New("failed to ping contest,"), err)
			}
		} else if !shouldPing {
			// Contests are sorted, so no more contests should be pinged after
			// the first that should not
			break
		}
	}

	return nil
}

func contestPing(contests *contestList, idx int, session *discordgo.Session) error {
	contests.mu.RLock()
	defer contests.mu.RUnlock()

	// Add contest ID to set
	pingedIDs[contests.contests[idx].ID] = struct{}{}

	pingList.mu.RLock()
	defer pingList.mu.RUnlock()
	// Issue ping for every ping channel (essentially for every server)
	for _, data := range pingList.list {
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
	data, err := getPingData(s, "contest-pings", "Contest Ping")
	if err != nil {
		return err
	}

	pingList.mu.Lock()
	pingList.list = data
	pingList.mu.Unlock()
	return nil
}

func getPingData(s *discordgo.Session, channelName string, roleName string) ([]pingData, error) {
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
				Name: roleName,
			})
			if err != nil {
				return nil, errors.Join(errors.New("failed to create ping role,"), err)
			}

			pingRole = newRole.ID
		}

		result = append(result, pingData{pingChannel, pingRole})
	}

	return result, nil
}
