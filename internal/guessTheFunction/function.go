package guessTheFunction

// functionDefinition a string which defines the function, mostly simple pythonic syntax
// returns a Expr type, call .eval(x) to evaluate the function
func makeNewFunction(functionDefinition string) (expr, error) {
	tokens, lexError := tokenizeInput(functionDefinition)
	if lexError != nil {
		return nil, lexError
	}

	expr, astError := buildAST(tokens)
	if astError != nil {
		return nil, astError
	}

	return expr, nil
}
