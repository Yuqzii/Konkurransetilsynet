package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"
	"fmt"
	
	"github.com/bwmarrin/discordgo"

	"konkurransetilsynet/gjettFunksjonen"
)

const prefix string = "!"

func main() {
	token := os.Getenv("TOKEN")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			_, err := s.ChannelMessageSend(m.ChannelID, "world!")
			if err != nil {
				log.Fatal("Hello command failed to execute, ", err)
			}
		case "gjettFunksjonen":
			log.Println("recived gjettFunksjonen command")
			// predefined function for testing
			function, parseError := gjettFunksjonen.MakeNewFunction("x^2 + 3x + 2")
    		if parseError != nil {
    		    log.Fatal("parsed function: ", err)
    		}
			
			output := ""
			output += fmt.Sprintf("f(2) = %f", function.Eval(2))
			output += fmt.Sprintf("f(10) = %f", function.Eval(10))
			log.Println(output)
			
			_, messageError := s.ChannelMessageSend(m.ChannelID, output)
			if messageError != nil {
				log.Fatal("gjettFunksjonen command failed to execute, ", messageError)
			}
		}
	})

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = session.Open()
	if err != nil {
		log.Fatal(err)
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
