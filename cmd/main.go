package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	codeforces "github.com/yuqzii/konkurransetilsynet/internal/codeforces"
	database "github.com/yuqzii/konkurransetilsynet/internal/database"
	guessTheFunction "github.com/yuqzii/konkurransetilsynet/internal/guessTheFunction"
	utilCommands "github.com/yuqzii/konkurransetilsynet/internal/utilCommands"

	"github.com/bwmarrin/discordgo"
)

const prefix string = "!"

func main() {
	logFile, err := enableLogFile()
	if err != nil {
		log.Fatal("Failed to enable logging to file: ", err)
	}
	// Close file when application exits
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Fatal("Failed to close log file: ", err)
		}
	}()

	// Connect to database
	conn, err := database.InitDB()
	if err != nil {
		log.Fatal("Could not connect to database: ", err)
	}
	log.Println("Connected to database.")
	// Close database when application exits
	defer func() {
		err := conn.Close(context.Background())
		if err != nil {
			log.Fatal("Failed to close database: ", err)
		}
	}()

	// Set up bot
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

func enableLogFile() (*os.File, error) {
	const fileName = "log.txt"
	// Write to both stderr and log file
	logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open log file %s,", fileName), err)
	}
	// Clear log file in case it is not empty
	err = os.Truncate(fileName, 0)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not truncate log file %s,", fileName), err)
	}
	mw := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(mw)
	return logFile, nil
}
