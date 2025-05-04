package guessTheFunction

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleGuessTheFunctionCommands(args []string, session *discordgo.Session, message *discordgo.MessageCreate) {
	log.Println("received guessTheFunction command")
	// TODO: this does not support spaces in function definition
	function, parseError := MakeNewFunction(args[1])
	if parseError != nil {
		log.Println("error parsing function: ", parseError)
		return
	}

	// Testing purposes
	output := ""
	output += fmt.Sprintf("f(2) = %f", function.Eval(2)) + "\n"
	output += fmt.Sprintf("f(10) = %f", function.Eval(10))
	log.Println(output)

	_, messageError := session.ChannelMessageSend(message.ChannelID, output)
	if messageError != nil {
		log.Println("guessTheFunction command failed to execute, ", messageError)
		return
	}
}
