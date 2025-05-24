package guessTheFunction

import (
	"errors"
	"fmt"
	"strconv"
)

// Order of precedence, lowest to highest
var precedenceTokenTypes = [3][]TokenType{
	{ADDITION_TOKEN, SUBTRACTION_TOKEN},
	{MULTIPLICATION_TOKEN, DIVISION_TOKEN},
	{POWER_TOKEN},
}

type exprConstructor func(left, right expr) expr

var tokenTypeToExprConstructMap = map[TokenType]exprConstructor{
	ADDITION_TOKEN:       func(l expr, r expr) expr { return add{Left: l, Right: r} },
	SUBTRACTION_TOKEN:    func(l expr, r expr) expr { return subtract{Left: l, Right: r} },
	MULTIPLICATION_TOKEN: func(l expr, r expr) expr { return multiply{Left: l, Right: r} },
	DIVISION_TOKEN:       func(l expr, r expr) expr { return divide{Left: l, Right: r} },
	POWER_TOKEN:          func(l expr, r expr) expr { return power{Left: l, Right: r} },
}

func canPrecedeBinaryOperator(t Token) bool {
	switch t.Type {
	case NUMBER_TOKEN, VARIABLE_TOKEN, RIGHT_PAREN_TOKEN:
		return true
	default:
		return false
	}
}

// Finds first index that matches one of the given operators
func findFirstBinaryOperatorOfType(operators []TokenType, tokens []Token) (int, bool) {
	depth := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		// Track parentheses depth
		switch tokens[i].Type {
		case RIGHT_PAREN_TOKEN:
			depth++
			continue
		case LEFT_PAREN_TOKEN:
			depth--
			continue
		}

		// Skip parentheses content
		if depth != 0 {
			continue
		}

		for _, operator := range operators {
			// Skip non-operators and operators not currently searched for
			if tokens[i].Type != operator {
				continue
			}

			// Ensure subtraction is binary. Ex: 2-x valid, -(x+3) invalid
			if operator == SUBTRACTION_TOKEN {
				if i == 0 {
					continue // Can't be binary at pos 0
				}

				if !canPrecedeBinaryOperator(tokens[i-1]) {
					continue
				}
			}

			return i, true
		}
	}
	return 0, false
}

func buildASTBinaryOperatorExpr(tokens []Token, idx int) (expr, error) {
	// Build left and right
	left, err := buildASTRecursive(tokens[:idx])
	if err != nil {
		return nil, err
	}
	right, err := buildASTRecursive(tokens[idx+1:])
	if err != nil {
		return nil, err
	}

	// Get constructor for type
	exprConstruct, ok := tokenTypeToExprConstructMap[tokens[idx].Type]
	if !ok {
		return nil, fmt.Errorf("unexpected token %v when building precedence tokens", tokens[idx])
	}

	return exprConstruct(left, right), nil
}

func buildSingleToken(t Token) (expr, error) {
	switch t.Type {
	case NUMBER_TOKEN:
		val, err := strconv.ParseFloat(t.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", t.Value)
		}
		return number{Value: val}, nil
	case VARIABLE_TOKEN:
		return variable{}, nil
	default:
		return nil, fmt.Errorf("unexpected token %v", t)
	}
}

func isValidEnclosedParen(tokens []Token) bool {
	// Check if enclosed in parentheses
	startParen := tokens[0].Type == LEFT_PAREN_TOKEN
	endParen := tokens[len(tokens)-1].Type == RIGHT_PAREN_TOKEN
	isEnclosed := startParen && endParen

	if !isEnclosed {
		return false
	}

	// Check all parentheses are closed after opened
	depth := 0
	validWrap := true
	for i, token := range tokens {
		switch token.Type {
		case LEFT_PAREN_TOKEN:
			depth++
		case RIGHT_PAREN_TOKEN:
			depth--
		}

		if depth == 0 && i != len(tokens)-1 {
			validWrap = false
			break
		}
	}

	return validWrap && depth == 0
}

func buildASTRecursive(tokens []Token) (expr, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens to parse")
	}

	// Strip a full ( ... )
	if isValidEnclosedParen(tokens) {
		return buildASTRecursive(tokens[1 : len(tokens)-1])
	}

	// Build expr for all binary operators
	for _, search_tokens := range precedenceTokenTypes {
		idx, found := findFirstBinaryOperatorOfType(search_tokens, tokens)
		if found {
			return buildASTBinaryOperatorExpr(tokens, idx)
		}
	}

	// Leading unary minus
	if tokens[0].Type == SUBTRACTION_TOKEN {
		right, err := buildASTRecursive(tokens[1:])
		if err != nil {
			return nil, err
		}
		return multiply{
			Left:  number{Value: -1},
			Right: right,
		}, nil
	}

	// Single token
	if len(tokens) == 1 {
		return buildSingleToken(tokens[0])
	}

	return nil, errors.New("could not parse tokens into AST")
}

func buildAST(tokens []Token) (expr, error) {
	AST, astErr := buildASTRecursive(tokens)

	if astErr != nil {
		return nil, fmt.Errorf("error generating AST from tokens %v: %w", tokens, astErr)
	}

	return AST, nil
}
