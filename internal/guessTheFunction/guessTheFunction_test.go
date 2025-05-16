package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const xValuesLowerBound float64 = -1000
const xValuesUpperBound float64 = 1000
const numberSamplesPerFunctionTest uint16 = 100
const maxTolerableError float64 = 1e-5

func logFunctionDefinition(fn expr, t *testing.T) {
	data, err := MarshalExpr(fn)
	if err != nil {
		t.Logf("unpacking error %s", err)
	} else {
		t.Logf("function: %s", string(data))
	}
}

func assertFunctionsApproxEqual(parsedFunc expr, correctFunc expr, t *testing.T) {
	for range numberSamplesPerFunctionTest {
		// Sample in interval given
		x := rand.Float64()*(xValuesUpperBound+math.Abs(xValuesLowerBound)) - xValuesLowerBound

		y_correct := correctFunc.Eval(x)
		y_parsed := parsedFunc.Eval(x)

		// Find relative difference
		absolute_difference := math.Abs(y_parsed - y_correct)
		y_average := (y_parsed + y_correct) / 2.0
		relative_difference := absolute_difference / y_average

		if relative_difference > maxTolerableError {
			logFunctionDefinition(parsedFunc, t)
			logFunctionDefinition(correctFunc, t)

			t.Fatalf("functions did not produce same value, x: %f y: %f y_pred: %f", x, correctFunc.Eval(x), parsedFunc.Eval(x))
		}
	}
}

func Test_MakeNewFunction(t *testing.T) {
	for i, tc := range TestCases_FunctionParsing {
		t.Run(tc.Input, func(t *testing.T) {
			parsedFunc, err := makeNewFunction(tc.Input)
			if err != nil {
				t.Fatalf("unexpected error on function index, %d error, %s", i, err)
				return
			}

			assertFunctionsApproxEqual(parsedFunc, tc.Expected, t)
		})
	}
}

func Test_MarshalAndUnMarshal(t *testing.T) {
	for i, correctFunc := range TestCases_SavingLoading {
		tName := fmt.Sprintf("function index: %d", i)
		t.Run(tName, func(t *testing.T) {
			// Serialize to JSON
			jsonData, err := MarshalExpr(correctFunc)
			if err != nil {
				t.Fatalf("unexpected error on function index, %d error, %s", i, err)
				return
			}

			// Deserialize back
			parsedFunc, err := unmarshalExpr(jsonData)
			if err != nil {
				panic(err)
			}

			// Validate parsedFunc
			assertFunctionsApproxEqual(parsedFunc, correctFunc, t)
		})
	}
}
