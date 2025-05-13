package guessTheFunction

import (
	"math"
)

type Expr interface {
	Eval(x float64) float64
	isExpr()
}

type Number struct {
	Value float64 `json:"value"`
}

type Variable struct{}

type Add struct {
	Left  Expr `json:"left"`
	Right Expr `json:"right"`
}

type Subtract struct {
	Left  Expr `json:"left"`
	Right Expr `json:"right"`
}

type Multiply struct {
	Left  Expr `json:"left"`
	Right Expr `json:"right"`
}

type Divide struct {
	Left  Expr `json:"left"`
	Right Expr `json:"right"`
}

type Power struct {
	Base     Expr `json:"base"`
	Exponent Expr `json:"exponent"`
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

func (Number) isExpr() {}

func (Variable) isExpr() {}

func (Add) isExpr() {}

func (Subtract) isExpr() {}

func (Multiply) isExpr() {}

func (Divide) isExpr() {}

func (Power) isExpr() {}
