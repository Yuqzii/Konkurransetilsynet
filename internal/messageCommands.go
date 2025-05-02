package messageCommands

import (
	"fmt"
	"log"

	"github.com/yuqzii/konkurransetilsynet/internal/guessTheFunction"

	"github.com/bwmarrin/discordgo"
)

func Hello(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID, "world!")
	return err
}

func GuessTheFunction(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	log.Println("recived guessTheFunction command")
	// TODO: this does not support spaces in function definition
	function, parseError := guessTheFunction.MakeNewFunction(args[1])
	if parseError != nil {
		log.Fatal("error parsing function: ", parseError)
		return fmt.Errorf("error parsing function: %s", parseError)
	}

	output := ""
	output += fmt.Sprintf("f(2) = %f", function.Eval(2)) + "\n"
	output += fmt.Sprintf("f(10) = %f", function.Eval(10))
	log.Println(output)

	_, messageError := s.ChannelMessageSend(m.ChannelID, output)
	if messageError != nil {
		log.Fatal("guessTheFunction command failed to execute, ", messageError)
		return fmt.Errorf("guessTheFunction command failed to execute, %s", messageError)
	}

	return nil
}
