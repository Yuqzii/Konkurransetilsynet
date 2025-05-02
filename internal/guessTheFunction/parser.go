package guessTheFunction

import (
	"strings"
)

func ParseFunction(input string) (Expr, error) {
	// Sanitize
	input = strings.ReplaceAll(input, " ", "")

	return Variable{}, nil
}
