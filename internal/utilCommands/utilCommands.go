package utilcommands

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	messageCommands "github.com/yuqzii/konkurransetilsynet/internal"
)

func HandleUtilCommands(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "log":
		err := dumpLog(s, m)
		if err != nil {
			return errors.Join(errors.New("failed to dump log,"), err)
		}
	default:
		err := messageCommands.UnknownCommand(s, m)
		if err != nil {
			return err
		}
	}

	return nil
}
