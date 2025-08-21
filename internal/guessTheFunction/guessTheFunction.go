package guessTheFunction

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	utilCommands "github.com/yuqzii/konkurransetilsynet/internal/utilCommands"
)

type gtfRound struct {
	functionDefinition       string
	functionExpr             expr
	channelID                string
	functionDomainLowerBound float64
	functionDomainUpperBound float64
}

var activeRounds = make(map[string]gtfRound)
var ErrNoActiveRound = errors.New("no active round in current channel")

func parseGTFStartRoundArgs(args []string) (functionDefinition string, domainLowerBound float64, domainUpperBound float64, err error) {
	if len(args) < 4 {
		return "", 0, 0, fmt.Errorf("invalid format, too few arguments")
	}

	if domainLowerBound, err = strconv.ParseFloat(args[2], 64); err != nil {
		return "", 0, 0, fmt.Errorf("float parsing error, %w", err)
	}
	if domainUpperBound, err = strconv.ParseFloat(args[3], 64); err != nil {
		return "", 0, 0, fmt.Errorf("float parsing error, %w", err)
	}

	functionDefinition = strings.Join(args[4:], "")

	return functionDefinition, domainLowerBound, domainUpperBound, nil
}

func startGTFRound(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	log.Println("starting GTF round!")

	// Parse args
	funcDef, lwrBound, uprBound, argErr := parseGTFStartRoundArgs(args)
	if argErr != nil {
		return fmt.Errorf("failed to parse arguments, %w", argErr)
	}

	// Parse function
	funcExpr, prsErr := makeNewFunction(funcDef)
	if prsErr != nil {
		return fmt.Errorf("failed to parse function: [%s], %w", funcDef, prsErr)
	}

	// Add to active rounds
	newRound := gtfRound{
		functionDefinition:       funcDef,
		functionExpr:             funcExpr,
		channelID:                m.ChannelID,
		functionDomainLowerBound: lwrBound,
		functionDomainUpperBound: uprBound,
	}
	activeRounds[m.ChannelID] = newRound

	// Confirmation message
	_, msgErr := s.ChannelMessageSend(m.ChannelID, "GTF Round started!")
	if msgErr != nil {
		return fmt.Errorf("failed to send confirmation message, %w", msgErr)
	}

	return nil
}

func HandleGuessTheFunctionCommands(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	switch args[1] {
	case "start":
		err := startGTFRound(args, s, m)
		if err != nil {
			return err
		}
	case "query":
		// TESTING PURPOSES
		x, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return err
		}

		r, ok := activeRounds[m.ChannelID]
		if !ok {
			return ErrNoActiveRound
		}
		y := r.functionExpr.Eval(x)

		msgStr := fmt.Sprintf("f(%f) = %f", x, y)
		_, err = s.ChannelMessageSend(m.ChannelID, msgStr)
		if err != nil {
			return err
		}
	case "guess":

	default:
		err := utilCommands.UnknownCommand(s, m)
		return err
	}

	return nil
}
