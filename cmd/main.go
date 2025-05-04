package main

import (
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

	session.AddHandler(onMessageCreate)

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = session.Open()
	if err != nil {
		log.Fatal("Could not open session with token ", err)
	}

	defer func() {
		err = session.Close() // Close session when application exits
	}()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Don't react to messages from this bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if string(m.Content[0:utf8.RuneCountInString(prefix)]) != prefix {
		return
	}

	// Get message arguments separated by space
	args := strings.Split(m.Content, " ")
	command := strings.TrimPrefix(args[0], prefix)

	switch command {
	case "hello":
		err := messageCommands.Hello(s, m)
		if err != nil {
			log.Println("Hello command failed to execute, ", err)
		}

	case "cf":
		codeforces.HandleCodeforcesCommands(args, s, m)

	case "guessTheFunction":
		guessTheFunction.HandleGuessTheFunctionCommands(args, s, m)

	default:
		err := messageCommands.UnknownCommand(s, m)
		if err != nil {
			log.Println("Unknown command failed to execute, ", err)
		}
	}
}
