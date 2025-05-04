package codeforces

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Manager struct {
	upcomingContests []contest
}

func (manager *Manager) HandleCodeforcesCommands(args []string, session *discordgo.Session,
	message *discordgo.MessageCreate) {
	if args[1] == "contests" {
		err := manager.listFutureContests(session, message)
		if err != nil {
			log.Println("Listing future Codeforces contests failed, ", err)
		}
	}
}
