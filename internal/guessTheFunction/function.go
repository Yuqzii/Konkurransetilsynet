package guessTheFunction

import "fmt"

// functionDefinition a string which defines the function, mostly simple pythonic syntax
// returns a Expr type, call .eval(x) to evaluate the function
func makeNewFunction(functionDefinition string) (expr, error) {
	tokens, err := tokenizeInput(functionDefinition)
	if err != nil {
		return nil, fmt.Errorf("tokenizing: %w", err)
	}

	expr, err := buildAST(tokens)
	if err != nil {
		return nil, err
	}

	return expr, nil
}
