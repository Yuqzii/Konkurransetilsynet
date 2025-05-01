package gjettFunksjonen

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
    expr, err := ParseFunction(functionDefinition)
    if err != nil {
        return nil, err
    }
    return &Function{Expression: expr}, nil
}
