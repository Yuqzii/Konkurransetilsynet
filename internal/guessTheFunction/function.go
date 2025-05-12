package guessTheFunction

import (
	"errors"
	"log"
)

// Function is a parsed, evaluable function
type Function struct {
	Expression Expr
}

func (f *Function) Eval(x float64) float64 {
	return f.Expression.Eval(x)
}

// MakeNewFunction
// functionDefinition a string which defines the function, mostly simple pythonic syntax
// returns a Function type, call .eval(x) to evaluate the function
func MakeNewFunction(functionDefinition string) (*Function, error) {
	tokens, lexError := TokenizeInput(functionDefinition)
	if lexError != nil {
		log.Println("error running lexical analysis on input, ", functionDefinition, ", error,", lexError)
		return nil, lexError
	}

	expr, astError := BuildAST(tokens)
	if astError != nil {
		log.Println("error  from tokens, ", tokens, ", error,", astError)
		return nil, errors.New("error building AST from tokens")
	}

	return &Function{Expression: expr}, nil
}
