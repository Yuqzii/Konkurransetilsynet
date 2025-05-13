package guessTheFunction

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func Test_MarshalAndUnMarshal(t *testing.T) {
	// Defined as code since loading is being tested
	functionDefinitions := [...]Expr{
		Add{
			Variable{}, Number{Value: 10}},

		Add{
			Variable{}, Number{Value: 1}},

		Subtract{
			Variable{}, Number{Value: 31}},

		Subtract{
			Variable{}, Number{Value: 9}},

		Add{
			Variable{}, Add{Number{Value: 13}, Number{Value: 3}}},

		Add{
			Number{Value: 12}, Subtract{Variable{}, Number{Value: 2}},
		},

		Add{
			Subtract{
				Variable{},
				Number{Value: 210},
			},
			Add{
				Variable{},
				Number{Value: 1},
			}},

		Add{
			Multiply{Number{Value: 3}, Variable{}},
			Number{Value: 2},
		},

		Subtract{
			Multiply{Number{Value: 10}, Variable{}},
			Number{Value: 9},
		},

		Add{
			Multiply{Number{Value: -10}, Variable{}},
			Number{Value: 3},
		},

		Add{
			Multiply{Variable{}, Number{Value: -5}},
			Number{Value: 10},
		},

		Add{
			Number{Value: -1},
			Multiply{Number{Value: 7}, Variable{}},
		},

		Number{Value: -1},

		Add{
			Number{Value: 1},
			Multiply{Number{Value: -1}, Variable{}},
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
			loadedExpr, err := UnmarshalExpr(jsonData)
			if err != nil {
				panic(err)
			}

			for i := 0; i < numberSamplesPerFunctionTest; i++ {
				x := rand.Float64() * 100
				difference := math.Abs(loadedExpr.Eval(x) - functionExpr.Eval(x))

				if difference > maxTolerableError {
					t.Logf("failed on test index: %d, x: %f, y: %f, y_reconstruct: %f", index, x, functionExpr.Eval(x), loadedExpr.Eval(x))
				}
			}
		})
	}
}
