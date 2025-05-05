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

type token struct {
	Type  int
	Value string
}

func lexFunctionDefinition(definition string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(definition) {
		ch := definition[i]

		if unicode.IsDigit(rune(ch)) {
			start := i
			// Find all consecutive number tokens
			for i < len(definition) && unicode.IsDigit(rune(definition[i])) {
				i++
			}
			tokens = append(tokens, token{Type: NUMBER_TOKEN, Value: definition[start:i]})
		} else {
			switch ch {
			case 'x':
				tokens = append(tokens, token{Type: VARIABLE_TOKEN, Value: "x"})
			case '+':
				tokens = append(tokens, token{Type: ADDITION_TOKEN, Value: "+"})
			case '-':
				tokens = append(tokens, token{Type: SUBTRACTION_TOKEN, Value: "-"})
			case '*':
				tokens = append(tokens, token{Type: MULTIPLICATION_TOKEN, Value: "*"})
			case '/':
				tokens = append(tokens, token{Type: DIVISION_TOKEN, Value: "/"})
			case '^':
				tokens = append(tokens, token{Type: POWER_TOKEN, Value: "^"})
			default:
				return nil, fmt.Errorf("unexpected char, %c in function definition, %s", ch, definition)
			}
		}
		i++
	}
	return tokens, nil
}

func generateAST(tokens []token) (Expr, error) {
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
			left, errLeft := generateAST(tokens[:i])
			if errLeft != nil {
				return nil, errors.New("unable to parse left in addition AST, good luck")
			}

			right, errRight := generateAST(tokens[i+1:])
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
	tokens, lexError := lexFunctionDefinition(input)
	if lexError != nil {
		log.Println("error running lexical analysis on input, ", input, ", error,", lexError)
		return nil, errors.New("error running lexical analysis on input")
	}

	// Generate AST
	AST, astError := generateAST(tokens)
	if astError != nil {
		log.Println("error generating AST from tokens, ", tokens, ", error,", astError)
		return nil, errors.New("error generating AST from tokens")
	}

	return AST, nil
}
