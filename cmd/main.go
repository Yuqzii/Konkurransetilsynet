package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	codeforces "github.com/yuqzii/konkurransetilsynet/internal/codeforces"
	guessTheFunction "github.com/yuqzii/konkurransetilsynet/internal/guessTheFunction"
	utilCommands "github.com/yuqzii/konkurransetilsynet/internal/utilCommands"

	"github.com/bwmarrin/discordgo"
)

const prefix string = "!"

func main() {
	// Write to both stderr and log file
	logFile, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Panic(err)
		}
	}()
	mw := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(mw)

	token := os.Getenv("TOKEN")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Could not create bot, ", err)
	}

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

	err = codeforces.Init(session)
	if err != nil {
		log.Fatal("Could not initialize Codeforces package:", err)
	}

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
			err := utilCommands.Hello(session, message)
			if err != nil {
				log.Fatal("Hello command failed to execute, ", err)
			}

		case "cf":
			err := codeforces.HandleCodeforcesCommands(args, session, message)
			if err != nil {
				log.Println("Codeforces command failed:", err)
			}

		case "guessTheFunction":
			guessTheFunction.HandleGuessTheFunctionCommands(args, session, message)

		case "utils":
			err := utilCommands.HandleUtilCommands(args, session, message)
			if err != nil {
				log.Println("Utility command failed:", err)
			}

		default:
			err := utilCommands.UnknownCommand(session, message)
			if err != nil {
				log.Println("Unknown command failed to execute, ", err)
			}
		}
	})

	log.Println("Bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
