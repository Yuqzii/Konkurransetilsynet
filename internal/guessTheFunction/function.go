package guessTheFunction

import (
	"errors"
	"log"
)

// functionDefinition a string which defines the function, mostly simple pythonic syntax
// returns a Function type, call .eval(x) to evaluate the function
func makeNewFunction(functionDefinition string) (expr, error) {
	tokens, lexError := tokenizeInput(functionDefinition)
	if lexError != nil {
		log.Println("error running lexical analysis on input, ", functionDefinition, ", error,", lexError)
		return nil, lexError
	}

	expr, astError := buildAST(tokens)
	if astError != nil {
		log.Println("error  from tokens, ", tokens, ", error,", astError)
		return nil, errors.New("error building AST from tokens")
	}

	return expr, nil
}
