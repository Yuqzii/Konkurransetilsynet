package guessTheFunction

import (
	"math"
)

type expr interface {
	Eval(x float64) float64
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
	Left  expr `json:"left"`
	Right expr `json:"right"`
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
	return math.Pow(p.Left.Eval(x), p.Right.Eval(x))
}
