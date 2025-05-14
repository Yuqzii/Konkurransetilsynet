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

func loadAllTestCases(dir string) ([]TestCase, error) {
	var allTests []TestCase

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

			var tests []TestCase
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
			parsedFunction, err := makeNewFunction(testData.Input)
			if err != nil {
				t.Fatal("unexpected error on function idx,", index, "error,", err)
				return
			}

			expectedFunction := testData.Expected

			for i := 0; i < numberSamplesPerFunctionTest; i++ {
				x := rand.Float64() * 100
				y_correct := expectedFunction.Eval(x)
				y_parsed := parsedFunction.Eval(x)

				absolute_difference := math.Abs(y_parsed - y_correct)
				y_average := (y_parsed + y_correct) / 2.0

				relative_difference := absolute_difference / y_average

				if relative_difference > maxTolerableError {
					data, err := MarshalExpr(parsedFunction)
					if err != nil {
						t.Logf("unpacking error %s", err)
					} else {
						t.Logf("decoded function: %s", string(data))
					}

					data, err = MarshalExpr(expectedFunction)
					if err != nil {
						t.Logf("unpacking error %s", err)
					} else {
						t.Logf("correct function: %s", string(data))
					}
					t.Fatalf("failed on test idx %d function, %s x: %f y: %f y_pred: %f", index, testData.Input, x, expectedFunction.Eval(x), parsedFunction.Eval(x))
				}
			}
		})
	}
}
