package guessTheFunction

import (
	"math"
	"math/rand"
	"testing"
)

type TestData struct {
	input  string
	answer Function
}

const numberSamplesPerFunctionTest int = 100
const maxTolerableError float64 = 1e-5

func TestMakeNewFunction_Linear(t *testing.T) {
	functionDefinitions := [...]TestData{
		{"x+10", Function{Add{
			Variable{}, Number{Value: 10}}}},

		{"x+1", Function{Add{
			Variable{}, Number{Value: 1}}}},

		{"x-31", Function{Subtract{
			Variable{}, Number{Value: 31}}}},

		{"x-9", Function{Subtract{
			Variable{}, Number{Value: 9}}}},

		{"x+13+3", Function{Add{
			Variable{}, Add{Number{Value: 13}, Number{Value: 3}}}}},

		{"12+x-2", Function{Add{
			Number{Value: 12}, Subtract{Variable{}, Number{Value: 2}},
		}}},

		{"x-210+x+1", Function{Subtract{
			Variable{},
			Add{
				Number{Value: 210},
				Add{Variable{}, Number{Value: 1}},
			},
		}}},

		{"3*x+2", Function{Add{
			Multiply{Number{Value: 3}, Variable{}},
			Number{Value: 2},
		}}},

		{"10*x-9", Function{Subtract{
			Multiply{Number{Value: 10}, Variable{}},
			Number{Value: 9},
		}}},

		{"-10*x+3", Function{Add{
			Multiply{Number{Value: -10}, Variable{}},
			Number{Value: 3},
		}}},

		{"x*-5+10", Function{Add{
			Multiply{Variable{}, Number{Value: -5}},
			Number{Value: 10},
		}}},

		{"-1+7*x", Function{Add{
			Number{Value: -1},
			Multiply{Number{Value: 7}, Variable{}},
		}}},

		{"-1", Function{Number{Value: -1}}},

		{"1+-x", Function{Add{
			Number{Value: 1},
			Multiply{Number{Value: -1}, Variable{}},
		}}},
	}

	returnedCorrectValues := true
	for index, testData := range functionDefinitions {
		parsedFunction, err := MakeNewFunction(testData.input)
		if err != nil {
			t.Fatal("unexpected error on function idx,", index, "error,", err)
		}

		expectedFunction := testData.answer
		returnedCorrectValues = true
		
		for i := 0; i < numberSamplesPerFunctionTest; i++ {
			x := rand.Float64() * 100
			difference := math.Abs(parsedFunction.Eval(x) - expectedFunction.Eval(x))

			if difference > maxTolerableError {
				returnedCorrectValues = false
			}
		}

		if !returnedCorrectValues {
			t.Fatal("failed on test idx", index, "function,", testData.input)
		}
	}
}
