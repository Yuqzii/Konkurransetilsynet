package utilcommands

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

func dumpLog(s *discordgo.Session, m *discordgo.MessageCreate) error {
	log, err := os.ReadFile("log.txt")
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s```", string(log)))
	return err
}
