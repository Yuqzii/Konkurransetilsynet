package guessTheFunction

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	utilCommands "github.com/yuqzii/konkurransetilsynet/internal/utilCommands"
)

type gtfRound struct {
	def       string
	expr      expr
	channelID string
	lb        float64
	ub        float64
}

var activeRounds = make(map[string]gtfRound)

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
	// Delete start message so other users can't see the function
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		return fmt.Errorf("deleting start message: %w", err)
	}

	// Parse args
	funcDef, lwrBound, uprBound, err := parseGTFStartRoundArgs(args)
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	// Parse function
	funcExpr, err := makeNewFunction(funcDef)
	if err != nil {
		return fmt.Errorf("parsing function [%s]: %w", funcDef, err)
	}

	// Add to active rounds
	newRound := gtfRound{
		def:       funcDef,
		expr:      funcExpr,
		channelID: m.ChannelID,
		lb:        lwrBound,
		ub:        uprBound,
	}
	activeRounds[m.ChannelID] = newRound

	// Confirmation message
	_, err = s.ChannelMessageSend(m.ChannelID, "GTF Round started!")
	if err != nil {
		return fmt.Errorf("failed to send confirmation message, %w", err)
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
			return sendNoActiveRoundMsg(m.ChannelID, s)
		}
		y := r.expr.Eval(x)

		msgStr := fmt.Sprintf("f(%f) = %f", x, y)
		_, err = s.ChannelMessageSend(m.ChannelID, msgStr)
		if err != nil {
			return err
		}
	case "guess":
		guessFunc := args[2]
		correct, err := guess(guessFunc, activeRounds[m.ChannelID])
		if err != nil {
			if errors.Is(err, ErrLex) {
				err = errors.Join(err, sendLexErrMsg(m.ChannelID, s))
			} else if errors.Is(err, ErrBuildingAST) {
				err = errors.Join(err, sendASTErrMsg(m.ChannelID, s))
			}
			return fmt.Errorf("guessing function: %w", err)
		}

		if correct {
			return sendCorrectGuessMsg(m.ChannelID, guessFunc, s)
		} else {
			return sendWrongGuessMsg(m.ChannelID, s)
		}

	default:
		err := utilCommands.UnknownCommand(s, m)
		return err
	}

	return nil
}

func sendNoActiveRoundMsg(channelID string, s *discordgo.Session) error {
	msgStr := "There is not an active Guess the Function round in this channel.\n" +
		"Start a new one with `!gtf start [lower bound] [upper bound] [function definition]`."
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}

func sendLexErrMsg(channelID string, s *discordgo.Session) error {
	msgStr := "Could not perform lexical analysis on your guess. Make sure it only contains valid characters."
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}

func sendASTErrMsg(channelID string, s *discordgo.Session) error {
	msgStr := "Could not build an AST from your guess. (your function doesn't make sense, git gud)."
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}
