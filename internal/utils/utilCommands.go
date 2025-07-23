package utils

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func HandleUtilCommands(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "log":
		err := dumpLog(s, m)
		if err != nil {
			return errors.Join(errors.New("failed to dump log,"), err)
		}
	default:
		err := UnknownCommand(s, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func Hello(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "world!")
	return err
}

func UnknownCommand(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "unknown command")
	return err
}
