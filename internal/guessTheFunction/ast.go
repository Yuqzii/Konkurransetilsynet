package guessTheFunction

import "math"

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