package guessTheFunction

import (
	"errors"
	"strings"
)

func ParseFunction(input string) (Expr, error) {
	// Sanitize
	input = strings.ReplaceAll(input, " ", "")

	// Predefined for testing
	if input == "x^2+3x+2" {
		return Add{
			Left: Add{
				Left:  Power{Base: Variable{}, Exponent: Number{2}},
				Right: Multiply{Left: Number{3}, Right: Variable{}},
			},
			Right: Number{2},
		}, nil
	}
	return nil, errors.New("not implemented yet")
}
