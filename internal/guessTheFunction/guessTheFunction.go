package guessTheFunction

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// Defined here so that it can be accessed outside the module
type TestCase struct {
	Input    string `json:"input"`
	Expected Expr   `json:"expected"`
}

func (tc *TestCase) MarshalJSON() ([]byte, error) {
	var jsonFormat struct {
		Input    string          `json:"input"`
		Expected json.RawMessage `json:"expected"`
	}
	jsonFormat.Input = tc.Input
	data, err := MarshalExpr(tc.Expected)
	if err != nil {
		return nil, err
	}
	jsonFormat.Expected = data

	return json.Marshal(jsonFormat)
}

func (tc *TestCase) UnmarshalJSON(data []byte) error {
	var jsonFormat struct {
		Input    string          `json:"input"`
		Expected json.RawMessage `json:"expected"`
	}
	if err := json.Unmarshal(data, &jsonFormat); err != nil {
		return err
	}

	tc.Input = jsonFormat.Input
	expr, err := UnmarshalExpr(jsonFormat.Expected)
	if err != nil {
		return err
	}
	tc.Expected = expr
	return nil
}

func HandleGuessTheFunctionCommands(args []string, session *discordgo.Session, message *discordgo.MessageCreate) {
	log.Println("received guessTheFunction command")
	// TODO: this does not support spaces in function definition
	function, parseError := MakeNewFunction(args[1])
	if parseError != nil {
		log.Println("error parsing function: ", parseError)
		return
	}

	// Testing purposes
	output := ""
	output += fmt.Sprintf("f(2) = %f", function.Eval(2)) + "\n"
	output += fmt.Sprintf("f(10) = %f", function.Eval(10))
	log.Println(output)

	_, messageError := session.ChannelMessageSend(message.ChannelID, output)
	if messageError != nil {
		log.Println("guessTheFunction command failed to execute, ", messageError)
		return
	}
}
