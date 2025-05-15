package guessTheFunction

import (
	"fmt"
	"testing"
)

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
			AssertFunctionsApproxEqual(parsedFunc, correctFunc, t)
		})
	}
}
