package guessTheFunction

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

var (
	ErrLex = errors.New("lexical analysis error")
)

const (
	NUMBER_TOKEN TokenType = iota
	VARIABLE_TOKEN
	ADDITION_TOKEN
	SUBTRACTION_TOKEN
	MULTIPLICATION_TOKEN
	DIVISION_TOKEN
	POWER_TOKEN
	LEFT_PAREN_TOKEN
	RIGHT_PAREN_TOKEN
)

func (t TokenType) String() string {
	switch t {
	case NUMBER_TOKEN:
		return "NUMBER"
	case VARIABLE_TOKEN:
		return "VARIABLE"
	case ADDITION_TOKEN:
		return "PLUS"
	case SUBTRACTION_TOKEN:
		return "MINUS"
	case MULTIPLICATION_TOKEN:
		return "TIMES"
	case DIVISION_TOKEN:
		return "DIVIDE"
	case POWER_TOKEN:
		return "POWER"
	case LEFT_PAREN_TOKEN:
		return "("
	case RIGHT_PAREN_TOKEN:
		return ")"
	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Type  TokenType
	Value string
}

var singleTokenTypeMap = map[byte]Token{
	'+': {Type: ADDITION_TOKEN, Value: "+"},
	'-': {Type: SUBTRACTION_TOKEN, Value: "-"},
	'*': {Type: MULTIPLICATION_TOKEN, Value: "*"},
	'/': {Type: DIVISION_TOKEN, Value: "/"},
	'^': {Type: POWER_TOKEN, Value: "^"},
	'(': {Type: LEFT_PAREN_TOKEN, Value: "("},
	')': {Type: RIGHT_PAREN_TOKEN, Value: ")"},
	'x': {Type: VARIABLE_TOKEN, Value: "x"},
}

func lexNumberString(definition string, i int) (*Token, int, error) {
	start := i
	ch := definition[i]
	dotSeen := ch == '.'
	i++

	for i < len(definition) {
		if unicode.IsDigit(rune(definition[i])) {
			i++
		} else if definition[i] == '.' {
			if dotSeen {
				return nil, 0, fmt.Errorf("multiple dots in number at position %d", i)
			}
			dotSeen = true
			i++
		} else {
			break
		}
	}

	onlyDot := start == i-1 && definition[start] == '.'
	if onlyDot {
		return nil, 0, fmt.Errorf("invalid standalone dot at position %d", start)
	}

	return &Token{Type: NUMBER_TOKEN, Value: definition[start:i]}, i, nil
}

func lexTokens(definition string) ([]Token, error) {
	var tokens []Token
	i := 0
	for i < len(definition) {
		ch := definition[i]

		// Parse numbers and floats
		if unicode.IsDigit(rune(ch)) || ch == '.' {
			token, endIdx, err := lexNumberString(definition, i)
			if err != nil {
				return nil, err
			}
			i = endIdx

			tokens = append(tokens, *token)
			continue
		}

		// Parse operators, parentheses, and variables
		token, ok := singleTokenTypeMap[ch]
		if !ok {
			return nil, fmt.Errorf("%w: unexpected character '%c' in input at position %d", ErrLex, ch, i)
		}
		tokens = append(tokens, token)

		i++
	}
	return tokens, nil
}

func tokenizeInput(input string) ([]Token, error) {
	// Sanitize
	input = strings.ReplaceAll(input, " ", "")

	// Find tokens
	tokens, err := lexTokens(input)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
