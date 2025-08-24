package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/bwmarrin/discordgo"
)

const (
	maxErr  float64 = 1e-5
	samples uint16  = 100
)

func guess(def string, round gtfRound) (bool, error) {
	expr, err := makeNewFunction(def)
	if err != nil {
		return false, fmt.Errorf("parsing function [%s]: %w", def, err)
	}

	correct := true
	for range samples {
		x := rand.Float64()*(round.ub+math.Abs(round.lb)) - round.lb

		y := expr.Eval(x)
		correctY := round.expr.Eval(x)

		absDiff := math.Abs(y - correctY)
		avg := (y + correctY) / 2
		relDiff := absDiff / avg

		if relDiff > maxErr {
			correct = false
			break
		}
	}

	return correct, nil
}

func sendCorrectGuessMsg(channelID, guessedFunc string, s *discordgo.Session) error {
	msgStr := fmt.Sprintf("Congratulations! You guessed the function!\n"+
		"Submitted function: `%s`\nYour function: `%s`", activeRounds[channelID].def, guessedFunc)
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}

func sendWrongGuessMsg(channelID string, s *discordgo.Session) error {
	msgStr := "Your guess was incorrect :( (skill issue tbh)."
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}
