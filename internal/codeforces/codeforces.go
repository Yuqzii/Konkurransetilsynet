package codeforces

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Manager struct {
	upcomingContests []Contest
}

func (manager *Manager) HandleCodeforcesCommands(args []string, session *discordgo.Session,
	message *discordgo.MessageCreate) {
	if args[1] == "contests" {
		err := listFutureContests(manager, session, message)
		if err != nil {
			log.Println("Listing future Codeforces contests failed, ", err)
		}
	}
}
