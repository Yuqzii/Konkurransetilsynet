package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/yuqzii/konkurransetilsynet/internal"
	"github.com/yuqzii/konkurransetilsynet/internal/codeforces"
	"github.com/yuqzii/konkurransetilsynet/internal/guessTheFunction"

	"github.com/bwmarrin/discordgo"
)

const prefix string = "!"

func main() {
	token := os.Getenv("TOKEN")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Could not create bot, ", err)
	}

	cfManager := codeforces.MakeManager()

	session.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
		// Don't react to messages from this bot
		if message.Author.ID == session.State.User.ID {
			return
		}

		// Don't react to messages without the prefix
		if string(message.Content[0:utf8.RuneCountInString(prefix)]) != prefix {
			return
		}

		// Get message arguments separated by space
		args := strings.Split(message.Content, " ")
		command := strings.TrimPrefix(args[0], prefix)

		switch command {
		case "hello":
			err := messageCommands.Hello(session, message)
			if err != nil {
				log.Fatal("Hello command failed to execute, ", err)
			}

		case "cf":
			cfManager.HandleCodeforcesCommands(args, session, message)

		case "guessTheFunction":
			log.Println("recived guessTheFunction command")
			// predefined function for testing
			function, parseError := guessTheFunction.MakeNewFunction("x^2 + 3x + 2")
			if parseError != nil {
				log.Fatal("error parsing function: ", parseError)
			}

			output := ""
			output += fmt.Sprintf("f(2) = %f", function.Eval(2))
			output += fmt.Sprintf("f(10) = %f", function.Eval(10))
			log.Println(output)

			_, messageError := session.ChannelMessageSend(message.ChannelID, output)
			if messageError != nil {
				log.Fatal("guessTheFunction command failed to execute, ", messageError)
			}
		}
	})

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = session.Open()
	if err != nil {
		log.Fatal("Could not open session with token ", err)
	}

	// Close session when application exits
	defer func() {
		err = session.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

