package guessTheFunction

import (
	"encoding/json"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const numberSamplesPerFunctionTest int = 100
const maxTolerableError float64 = 1e-5

type testData struct {
	Input    string `json:"input"`
	Expected Expr   `json:"expected"`
}

func (td *testData) MarshalJSON() ([]byte, error) {
	var jsonFormat struct {
		Input    string          `json:"input"`
		Expected json.RawMessage `json:"expected"`
	}
	jsonFormat.Input = td.Input
	data, err := MarshalExpr(td.Expected)
	if err != nil {
		return nil, err
	}
	jsonFormat.Expected = data

	return json.Marshal(jsonFormat)
}

func (td *testData) UnmarshalJSON(data []byte) error {
	var jsonFormat struct {
		Input    string          `json:"input"`
		Expected json.RawMessage `json:"expected"`
	}
	if err := json.Unmarshal(data, &jsonFormat); err != nil {
		return err
	}

	td.Input = jsonFormat.Input
	expr, err := UnmarshalExpr(jsonFormat.Expected)
	if err != nil {
		return err
	}
	td.Expected = expr
	return nil
}

func loadAllTestCases(dir string) ([]testData, error) {
	var allTests []testData

	// all files in dir
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".json" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var tests []testData
			if err := json.Unmarshal(data, &tests); err != nil {
				return err
			}

			allTests = append(allTests, tests...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return allTests, nil
}

func TestMakeNewFunction(t *testing.T) {
	functionDefinitions, err := loadAllTestCases("testdata")
	if err != nil {
		t.Fatalf("could not load test cases: %v", err)
	}

	for index, testData := range functionDefinitions {
		t.Run(testData.Input, func(t *testing.T) {
			parsedFunction, err := MakeNewFunction(testData.Input)
			if err != nil {
				t.Fatal("unexpected error on function idx,", index, "error,", err)
				return
			}

			expectedFunction := testData.Expected

			for i := 0; i < numberSamplesPerFunctionTest; i++ {
				x := rand.Float64() * 100
				difference := math.Abs(parsedFunction.Eval(x) - expectedFunction.Eval(x))

				if difference > maxTolerableError {
					data, err := MarshalExpr(parsedFunction)
					if err != nil {
						t.Logf("unpacking error %s", err)
					} else {
						t.Logf("decoded function: %s", string(data))
					}
					t.Fatalf("failed on test idx %d function, %s x: %f y: %f y_pred: %f", index, testData.Input, x, expectedFunction.Eval(x), parsedFunction.Eval(x))
				}
			}
		})
	}
}
