package guessTheFunction

import (
	"math"
)

type expr interface {
	Eval(x float64) float64
	isExpr()
}

type number struct {
	Value float64 `json:"value"`
}

type variable struct{}

type add struct {
	Left  expr `json:"left"`
	Right expr `json:"right"`
}

type subtract struct {
	Left  expr `json:"left"`
	Right expr `json:"right"`
}

type multiply struct {
	Left  expr `json:"left"`
	Right expr `json:"right"`
}

type divide struct {
	Left  expr `json:"left"`
	Right expr `json:"right"`
}

type power struct {
	Base     expr `json:"base"`
	Exponent expr `json:"exponent"`
}

func (n number) Eval(x float64) float64 {
	return n.Value
}

func (v variable) Eval(x float64) float64 {
	return x
}

func (a add) Eval(x float64) float64 {
	return a.Left.Eval(x) + a.Right.Eval(x)
}

func (s subtract) Eval(x float64) float64 {
	return s.Left.Eval(x) - s.Right.Eval(x)
}

func (m multiply) Eval(x float64) float64 {
	return m.Left.Eval(x) * m.Right.Eval(x)
}

func (d divide) Eval(x float64) float64 {
	return d.Left.Eval(x) / d.Right.Eval(x)
}

func (p power) Eval(x float64) float64 {
	return math.Pow(p.Base.Eval(x), p.Exponent.Eval(x))
}

func (number) isExpr() {}

func (variable) isExpr() {}

func (add) isExpr() {}

func (subtract) isExpr() {}

func (multiply) isExpr() {}

func (divide) isExpr() {}

func (power) isExpr() {}
