package guessTheFunction

import (
	"errors"
	"fmt"
	"log"
	"strconv"
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
)

type Token struct {
	Type  int
	Value string
}

func LexFunctionDefinition(definition string) ([]Token, error) {
	var tokens []Token
	i := 0
	for i < len(definition) {
		ch := definition[i]

		if unicode.IsDigit(rune(ch)) {
			start := i
			// Find all consecutive number tokens
			for i < len(definition) && unicode.IsDigit(rune(definition[i])) {
				i++
			}
			tokens = append(tokens, Token{Type: NUMBER_TOKEN, Value: definition[start:i]})
		} else if ch == 'x' {
			tokens = append(tokens, Token{Type: VARIABLE_TOKEN, Value: "x"})
		} else if ch == '+' {
			tokens = append(tokens, Token{Type: ADDITION_TOKEN, Value: "+"})
		} else if ch == '-' {
			tokens = append(tokens, Token{Type: SUBTRACTION_TOKEN, Value: "-"})
		} else if ch == '*' {
			tokens = append(tokens, Token{Type: MULTIPLICATION_TOKEN, Value: "*"})
		} else if ch == '/' {
			tokens = append(tokens, Token{Type: DIVISION_TOKEN, Value: "/"})
		} else if ch == '^' {
			tokens = append(tokens, Token{Type: POWER_TOKEN, Value: "^"})
		} else {
			return nil, fmt.Errorf("unexpected char, %c in function definition, %s", ch, definition)
		}
		i++
	}
	return tokens, nil
}

func GenerateAST(tokens []Token) (Expr, error) {
	// Generate base case for array of length 1
	if len(tokens) == 1 {
		token := tokens[0]

		switch token.Type {
		case VARIABLE_TOKEN:
			return Variable{}, nil
		case NUMBER_TOKEN:
			value, parsErr := strconv.ParseFloat(token.Value, 64)
			if parsErr != nil {
				return nil, fmt.Errorf("unable to parse float, %s", token.Value)
			}
			return Number{Value: value}, nil
		default:
			return nil, fmt.Errorf("unsupported token type in AST generation, %d", token.Type)
		}
	}

	i := 0
	for i < len(tokens) {
		currentToken := tokens[i]

		// Handle addition token, splits token array recursively
		if currentToken.Type == ADDITION_TOKEN && i > 0 && i < len(tokens)-1 {
			left, errLeft := GenerateAST(tokens[:i])
			if errLeft != nil {
				return nil, errors.New("unable to parse left in addition AST, good luck")
			}

			right, errRight := GenerateAST(tokens[i+1:])
			if errRight != nil {
				return nil, errors.New("unable to parse right in addition AST, good luck")
			}

			AST := Add{Left: left, Right: right}
			return AST, nil
		}
		i++
	}
	return nil, errors.New("unable to generate AST, good luck")
}

func ParseFunction(input string) (Expr, error) {
	// Sanitize
	input = strings.ReplaceAll(input, " ", "")

	// Find tokens
	tokens, lexError := LexFunctionDefinition(input)
	if lexError != nil {
		log.Fatal("error running lexical analysis on input, ", input, ", error,", lexError)
		return nil, errors.New("error running lexical analysis on input")
	}

	// Generate AST
	AST, astError := GenerateAST(tokens)
	if astError != nil {
		log.Fatal("error generating AST from tokens, ", tokens, ", error,", astError)
		return nil, errors.New("error generating AST from tokens")
	}

	return AST, nil
}
