package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

type testData struct {
	input  string
	answer Function
}

const numberSamplesPerFunctionTest int = 100
const maxTolerableError float64 = 1e-5

func TestMakeNewFunction_Linear(t *testing.T) {
	functionDefinitions := [...]testData{
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

		{"x-210+x+1", Function{Add{
			Subtract{
				Variable{},
				Number{Value: 210},
			},
			Add{
				Variable{},
				Number{Value: 1},
			}},
		}},

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

	for index, testData := range functionDefinitions {
		parsedFunction, err := MakeNewFunction(testData.input)
		if err != nil {
			fmt.Println("unexpected error on function idx,", index, "error,", err)
			t.Fail()
			continue
		}

		expectedFunction := testData.answer

		for i := 0; i < numberSamplesPerFunctionTest; i++ {
			x := rand.Float64() * 100
			difference := math.Abs(parsedFunction.Eval(x) - expectedFunction.Eval(x))

			if difference > maxTolerableError {
				t.Logf("failed on test idx %d function, %s x: %f y: %f y_pred: %f", index, testData.input, x, expectedFunction.Eval(x), parsedFunction.Eval(x))
			}
		}

	}
}
