package messageCommands

import "github.com/bwmarrin/discordgo"

func Hello(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "world!")
	return err
}

func UnknownCommand(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "unknown comamnd")
	return err
}
