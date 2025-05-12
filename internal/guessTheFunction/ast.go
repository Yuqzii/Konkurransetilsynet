package guessTheFunction

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
)

type Expr interface {
	Eval(x float64) float64
}

type Number struct {
	Value float64
}

type Variable struct{}

type Add struct {
	Left, Right Expr
}

type Subtract struct {
	Left, Right Expr
}

type Multiply struct {
	Left, Right Expr
}

type Divide struct {
	Left, Right Expr
}

type Power struct {
	Base, Exponent Expr
}

func (n Number) Eval(x float64) float64 {
	return n.Value
}

func (v Variable) Eval(x float64) float64 {
	return x
}

func (a Add) Eval(x float64) float64 {
	return a.Left.Eval(x) + a.Right.Eval(x)
}

func (s Subtract) Eval(x float64) float64 {
	return s.Left.Eval(x) - s.Right.Eval(x)
}

func (m Multiply) Eval(x float64) float64 {
	return m.Left.Eval(x) * m.Right.Eval(x)
}

func (d Divide) Eval(x float64) float64 {
	return d.Left.Eval(x) / d.Right.Eval(x)
}

func (p Power) Eval(x float64) float64 {
	return math.Pow(p.Base.Eval(x), p.Exponent.Eval(x))
}

// Finds first index that matches one of the given operators
func findFirstBinaryOperatorOfType(operators []int, tokens []Token) int {
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
			// Skip non operators
			if tokens[i].Type != operator {
				continue
			}

			// If subtraction, make sure it's binary
			if operator == SUBTRACTION_TOKEN {
				if i == 0 {
					continue // can't be binary at pos 0
				}
				prev := tokens[i-1].Type
				if prev != NUMBER_TOKEN && prev != VARIABLE_TOKEN && prev != RIGHT_PAREN_TOKEN {
					continue
				}
			}

			return i
		}
	}
	// Didn't find any
	return -1
}

func buildASTRecursive(tokens []Token) (Expr, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens to parse")
	}

	// Strip a full ( ... )
	if tokens[0].Type == LEFT_PAREN_TOKEN && tokens[len(tokens)-1].Type == RIGHT_PAREN_TOKEN {
		depth := 0
		validWrap := true
		for i, token := range tokens {
			if token.Type == LEFT_PAREN_TOKEN {
				depth++
			} else if token.Type == RIGHT_PAREN_TOKEN {
				depth--
				if depth == 0 && i != len(tokens)-1 {
					validWrap = false
					break
				}
			}
		}
		if validWrap && depth == 0 {
			return buildASTRecursive(tokens[1 : len(tokens)-1])
		}
	}

	// Leading unary minus
	if tokens[0].Type == SUBTRACTION_TOKEN {
		right, err := buildASTRecursive(tokens[1:])
		if err != nil {
			return nil, err
		}
		return Multiply{
			Left:  Number{Value: -1},
			Right: right,
		}, nil
	}

	// order of precedence least to most
	precedenceTokenTypes := [][]int{
		{ADDITION_TOKEN, SUBTRACTION_TOKEN},
		{MULTIPLICATION_TOKEN, DIVISION_TOKEN},
		{POWER_TOKEN},
	}

	for _, search_tokens := range precedenceTokenTypes {
		idx := findFirstBinaryOperatorOfType(search_tokens, tokens)
		if idx != -1 {
			left, err := buildASTRecursive(tokens[:idx])
			if err != nil {
				return nil, err
			}
			right, err := buildASTRecursive(tokens[idx+1:])
			if err != nil {
				return nil, err
			}

			switch tokens[idx].Type {
			case ADDITION_TOKEN:
				return Add{Left: left, Right: right}, nil
			case SUBTRACTION_TOKEN:
				return Subtract{Left: left, Right: right}, nil
			case MULTIPLICATION_TOKEN:
				return Multiply{Left: left, Right: right}, nil
			case DIVISION_TOKEN:
				return Divide{Left: left, Right: right}, nil
			case POWER_TOKEN:
				return Power{Base: left, Exponent: right}, nil
			default:
				return nil, fmt.Errorf("unexpected token %v when building precedence tokens", tokens[idx])
			}
		}
	}

	// Single token
	if len(tokens) == 1 {
		switch tokens[0].Type {
		case NUMBER_TOKEN:
			val, err := strconv.ParseFloat(tokens[0].Value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number %q", tokens[0].Value)
			}
			return Number{Value: val}, nil
		case VARIABLE_TOKEN:
			return Variable{}, nil
		default:
			return nil, fmt.Errorf("unexpected token %v", tokens[0])
		}
	}

	return nil, errors.New("could not parse tokens into AST")
}

func BuildAST(tokens []Token) (Expr, error) {
	// Build AST
	AST, astError := buildASTRecursive(tokens)
	if astError != nil {
		log.Println("error generating AST from tokens, ", tokens, ", error,", astError)
		return nil, errors.New("error generating AST from tokens")
	}

	return AST, nil
}
