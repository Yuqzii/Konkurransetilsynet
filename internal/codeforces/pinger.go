package codeforces

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

type pingData struct {
	channel string
	role    string
}

type contestPinger struct {
	discord   *discordgo.Session
	contests  contestProvider
	guilds    guildProvider
	pingData  []pingData
	pingedIDs map[uint32]struct{}
	mu        sync.RWMutex
}

const (
	pingTime        uint32 = 1 * 3600 // 1 hour
	pingChannelName string = "contest-pings"
	pingRoleName    string = "Contest Ping"
)

func newPinger(d *discordgo.Session, c contestProvider, g guildProvider) *contestPinger {
	return &contestPinger{discord: d, contests: c, guilds: g, pingedIDs: make(map[uint32]struct{})}
}

// Start goroutine that checks whether it should issue a ping for upcoming contests
func (p *contestPinger) StartContestPingCheck(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			err := p.checkContestPing()
			if err != nil {
				log.Println("Automatic contest ping failed:", err)
			}
		}
	}()
}

func (p *contestPinger) checkContestPing() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	curTime := uint32(time.Now().Unix())
	contests := p.contests.getContests()
	for _, c := range contests {
		shouldPing := c.StartTimeSeconds-curTime <= pingTime
		_, isPinged := p.pingedIDs[c.ID]
		if shouldPing && !isPinged {
			err := p.pingContest(c)
			if err != nil {
				return fmt.Errorf("pinging contest: %w", err)
			}
		} else if !shouldPing {
			// Contests are sorted, so no more contests should be pinged after
			// the first that should not
			break
		}
	}

	return nil
}

func (p *contestPinger) pingContest(c *contest) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// Add contest ID to set
	p.pingedIDs[c.ID] = struct{}{}

	// Issue ping for every ping channel (essentially for every server)
	for _, data := range p.pingData {
		_, err := p.discord.ChannelMessageSend(data.channel,
			fmt.Sprintf("<@&%s> **%s** is starting <t:%d:R>",
				data.role, c.Name, c.StartTimeSeconds))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *contestPinger) updatePingData() error {
	guilds := p.guilds.getGuilds()
	channels, err := utils.CreateChannelIfNotExist(p.discord, pingChannelName, guilds)
	if err != nil {
		return fmt.Errorf("finding/creating ping channel: %w", err)
	}

	roles, err := utils.CreateRoleIfNotExists(p.discord, pingRoleName, guilds)
	if err != nil {
		return fmt.Errorf("finding/creating ping role: %w", err)
	}

	var newList []pingData
	for i := range guilds {
		newList = append(newList, pingData{channel: channels[i], role: roles[i]})
	}

	p.mu.Lock()
	p.pingData = newList
	p.mu.Unlock()
	return nil
}
