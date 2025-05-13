package main

import (
	"encoding/json"
	"fmt"
	"github.com/yuqzii/konkurransetilsynet/internal/guessTheFunction"
	"os"
)

func saveTestCasesToFile(testCases []*guessTheFunction.TestCase, filename string) {
	jsonData, err := json.Marshal(testCases)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(fmt.Sprintf("./testdata/%s", filename), jsonData, 0644); err != nil {
		panic(err)
	}
}

func main() {
	saveTestCasesToFile(guessTheFunction.TestCasesLinear, "linear_test.json")
	saveTestCasesToFile(guessTheFunction.TestCasesPolynomial, "polynomial_test.json")
}
