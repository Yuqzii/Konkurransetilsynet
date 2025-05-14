package guessTheFunction

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	NUMBER_TOKEN = iota
	VARIABLE_TOKEN
	ADDITION_TOKEN
	SUBTRACTION_TOKEN
	MULTIPLICATION_TOKEN
	DIVISION_TOKEN
	POWER_TOKEN
	LEFT_PAREN_TOKEN
	RIGHT_PAREN_TOKEN
)

type Token struct {
	Type  int
	Value string
}

func lexFunctionDefinition(definition string) ([]Token, error) {
	var tokens []Token
	i := 0
	for i < len(definition) {
		ch := definition[i]

		// Numbers (supporting decimals)
		if unicode.IsDigit(rune(ch)) || ch == '.' {
			start := i
			dotSeen := ch == '.'
			i++

			for i < len(definition) {
				if unicode.IsDigit(rune(definition[i])) {
					i++
				} else if definition[i] == '.' {
					if dotSeen {
						return nil, fmt.Errorf("multiple dots in number at position %d", i)
					}
					dotSeen = true
					i++
				} else {
					break
				}
			}

			if start == i-1 && definition[start] == '.' {
				return nil, fmt.Errorf("invalid standalone dot at position %d", start)
			}

			tokens = append(tokens, Token{Type: NUMBER_TOKEN, Value: definition[start:i]})
			continue
		}

		// Variables
		if ch == 'x' {
			tokens = append(tokens, Token{Type: VARIABLE_TOKEN, Value: "x"})
			i++
			continue
		}

		// Operators and parentheses
		switch ch {
		case '+':
			tokens = append(tokens, Token{Type: ADDITION_TOKEN, Value: "+"})
		case '-':
			tokens = append(tokens, Token{Type: SUBTRACTION_TOKEN, Value: "-"})
		case '*':
			tokens = append(tokens, Token{Type: MULTIPLICATION_TOKEN, Value: "*"})
		case '/':
			tokens = append(tokens, Token{Type: DIVISION_TOKEN, Value: "/"})
		case '^':
			tokens = append(tokens, Token{Type: POWER_TOKEN, Value: "^"})
		case '(':
			tokens = append(tokens, Token{Type: LEFT_PAREN_TOKEN, Value: "("})
		case ')':
			tokens = append(tokens, Token{Type: RIGHT_PAREN_TOKEN, Value: ")"})
		default:
			return nil, fmt.Errorf("unexpected character '%c' in input at position %d", ch, i)
		}
		i++
	}
	return tokens, nil
}

func tokenizeInput(input string) ([]Token, error) {
	// Sanitize
	input = strings.ReplaceAll(input, " ", "")

	// Find tokens
	tokens, lexError := lexFunctionDefinition(input)
	if lexError != nil {
		return nil, errors.New("error running lexical analysis on input")
	}

	return tokens, nil
}
