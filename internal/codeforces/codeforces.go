package codeforces

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleCodeforcesCommands(args []string, session *discordgo.Session,
	message *discordgo.MessageCreate) {
	if args[1] == "contests" {
		err := ListFutureContests(session, message)
		if err != nil {
			log.Println("Listing future Codeforces contests failed, ", err)
		}
	}
}
