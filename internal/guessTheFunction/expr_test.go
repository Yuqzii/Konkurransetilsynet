package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func Test_MarshalAndUnMarshal(t *testing.T) {
	// Defined as code since loading is being tested
	functionDefinitions := [...]expr{
		add{
			variable{}, number{Value: 10}},

		add{
			variable{}, number{Value: 1}},

		subtract{
			variable{}, number{Value: 31}},

		subtract{
			variable{}, number{Value: 9}},

		add{
			variable{}, add{number{Value: 13}, number{Value: 3}}},

		add{
			number{Value: 12}, subtract{variable{}, number{Value: 2}},
		},

		add{
			subtract{
				variable{},
				number{Value: 210},
			},
			add{
				variable{},
				number{Value: 1},
			}},

		add{
			multiply{number{Value: 3}, variable{}},
			number{Value: 2},
		},

		subtract{
			multiply{number{Value: 10}, variable{}},
			number{Value: 9},
		},

		add{
			multiply{number{Value: -10}, variable{}},
			number{Value: 3},
		},

		add{
			multiply{variable{}, number{Value: -5}},
			number{Value: 10},
		},

		add{
			number{Value: -1},
			multiply{number{Value: 7}, variable{}},
		},

		number{Value: -1},

		add{
			number{Value: 1},
			multiply{number{Value: -1}, variable{}},
		},
	}

	for index, functionExpr := range functionDefinitions {
		t.Run(fmt.Sprintf("function index:%d", index), func(t *testing.T) {
			// Serialize to JSON
			jsonData, err := MarshalExpr(functionExpr)
			if err != nil {
				t.Fatal("unexpected error on function index,", index, "error,", err)
			}

			// Deserialize back
			loadedExpr, err := unmarshalExpr(jsonData)
			if err != nil {
				panic(err)
			}

			for i := 0; i < numberSamplesPerFunctionTest; i++ {
				x := rand.Float64() * 100
				y_correct := functionExpr.Eval(x)
				y_parsed := loadedExpr.Eval(x)

				absolute_difference := math.Abs(y_parsed - y_correct)
				y_average := (y_parsed + y_correct) / 2.0

				relative_difference := absolute_difference / y_average

				if relative_difference > maxTolerableError {
					t.Logf("failed on test index: %d, x: %f, y: %f, y_reconstruct: %f", index, x, functionExpr.Eval(x), loadedExpr.Eval(x))
				}
			}
		})
	}
}
