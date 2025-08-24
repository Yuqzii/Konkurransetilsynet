package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/bwmarrin/discordgo"
)

const (
	maxErr  float64 = 1e-5
	minX    float64 = -1000
	maxX    float64 = 1000
	samples uint16  = 100
)

func guess(def string, actual expr) (bool, error) {
	expr, err := makeNewFunction(def)
	if err != nil {
		return false, fmt.Errorf("parsing function [%s]: %w", def, err)
	}

	correct := true
	for range samples {
		x := rand.Float64()*(maxX+math.Abs(minX)) - minX

		y := expr.Eval(x)
		correctY := actual.Eval(x)

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
		"Submitted function: `%s`\nYour function: `%s`", activeRounds[channelID].functionDefinition, guessedFunc)
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}

func sendWrongGuessMsg(channelID string, s *discordgo.Session) error {
	msgStr := fmt.Sprintf("Your guess was incorrect :( (skill issue tbh).")
	_, err := s.ChannelMessageSend(channelID, msgStr)
	return err
}
