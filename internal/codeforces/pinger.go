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
	discord  *discordgo.Session
	contests contestProvider
	guilds   guildProvider

	pingTime        time.Duration
	pingChannelName string
	pingRoleName    string

	pingData  []pingData
	pingedIDs map[uint32]struct{}
	mu        sync.RWMutex
}

type pingerOption func(*contestPinger)

func newPinger(discord *discordgo.Session, contests contestProvider,
	guilds guildProvider, opts ...pingerOption) *contestPinger {

	const (
		defaultPingTime        time.Duration = 1 * time.Hour
		defaultPingChannelName string        = "contest-pings"
		defaultPingRoleName    string        = "Contest Ping"
	)

	p := &contestPinger{
		discord:         discord,
		contests:        contests,
		guilds:          guilds,
		pingedIDs:       make(map[uint32]struct{}),
		pingTime:        defaultPingTime,
		pingChannelName: defaultPingChannelName,
		pingRoleName:    defaultPingRoleName,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithPingTime(t time.Duration) pingerOption {
	return func(p *contestPinger) {
		p.pingTime = t
	}
}

func WithPingChannelName(name string) pingerOption {
	return func(p *contestPinger) {
		p.pingChannelName = name
	}
}

func WithPingRoleName(name string) pingerOption {
	return func(p *contestPinger) {
		p.pingRoleName = name
	}
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
		shouldPing := c.StartTimeSeconds-curTime <= uint32(p.pingTime.Seconds())
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
	channels, err := utils.CreateChannelIfNotExist(p.discord, p.pingChannelName, guilds)
	if err != nil {
		return fmt.Errorf("finding/creating ping channel: %w", err)
	}

	roles, err := utils.CreateRoleIfNotExists(p.discord, p.pingRoleName, guilds)
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
