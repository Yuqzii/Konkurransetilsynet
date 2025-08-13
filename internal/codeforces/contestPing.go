package codeforces

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yuqzii/konkurransetilsynet/internal/utils"
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

const (
	pingTime        uint32 = 1 * 3600 // 1 hour
	pingChannelName string = "contest-pings"
	pingRoleName    string = "Contest Ping"
)

// Start goroutine that checks whether it should issue a ping for upcoming contests
func (s *Service) startContestPingCheck(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := s.checkContestPing()
			if err != nil {
				log.Println("Automatic contest ping failed:", err)
			}
		}
	}()
}

func (s *Service) checkContestPing() error {
	s.contestsMu.RLock()
	defer s.contestsMu.RUnlock()

	curTime := uint32(time.Now().Unix())
	for i := range s.contests {
		shouldPing := s.contests[i].StartTimeSeconds-curTime <= pingTime
		_, isPinged := s.pingedIDs[s.contests[i].ID]
		if shouldPing && !isPinged {
			err := s.pingContest(i)
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

func (s *Service) pingContest(idx int) error {
	s.pingMu.RLock()
	defer s.pingMu.RUnlock()
	// Add contest ID to set
	s.pingedIDs[s.contests[idx].ID] = struct{}{}

	// Issue ping for every ping channel (essentially for every server)
	for _, data := range s.pingData {
		_, err := s.discord.ChannelMessageSend(data.channel,
			fmt.Sprintf("<@&%s> **%s** is starting <t:%d:R>",
				data.role, s.contests[idx].Name, s.contests[idx].StartTimeSeconds))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) updatePingData() error {
	channels, err := utils.CreateChannelIfNotExist(s.discord, pingChannelName, s.guilds)
	if err != nil {
		return fmt.Errorf("failed to find ping channel: %w", err)
	}

	roles, err := utils.CreateRoleIfNotExists(s.discord, pingRoleName, s.guilds)
	if err != nil {
		return fmt.Errorf("failed to find ping role: %w", err)
	}

	var newList []pingData
	for i := range s.guilds {
		newList = append(newList, pingData{channel: channels[i], role: roles[i]})
	}

	s.pingMu.Lock()
	s.pingData = newList
	s.pingMu.Unlock()
	return nil
}
