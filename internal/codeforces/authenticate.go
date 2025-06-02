package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utilCommands"
)

type submissionList struct {
	Submissions []submission `json:"result"`
	Status      string       `json:"status"`
	Comment     string       `json:"comment,omitempty"`
}

type submission struct {
	ID                  int   `json:"id"`
	ContestID           int   `json:"contestId"`
	CreationTimeSeconds int64 `json:"creationTimeSeconds"`
	RelativeTimeSeconds int   `json:"relativeTimeSeconds"`
	Problem             struct {
		ContestID int    `json:"contestId"`
		Index     string `json:"index"`
		Name      string `json:"name"`
	}
	Verdict string `json:"verdict"`
}

type problemList struct {
	Result struct {
		Problems []problem `json:"problems"`
	} `json:"result"`
	Status  string `json:"status"`
	Comment string `json:"comment,omitempty"`
}

type problem struct {
	ContestID int    `json:"contestId"`
	Index     string `json:"index"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

func authCommand(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	// Ensure correct argument count
	if len(args) < 3 {
		err := utilCommands.UnknownCommand(s, m)
		return err
	}
	// Start authentication goroutine
	go func() {
		err := authenticate(args[2], s, m)
		if err != nil {
			log.Println("Authentication failed: ", err)
		}
	}()
	return nil
}

func authenticate(handle string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	log.Printf("Received codeforces authenticate for handle '%s'.", handle)
	prob, err := getRandomProblem()
	if err != nil {
		return fmt.Errorf("failed to get a random problem: %w", err)
	}
	sendAuthInstructions(prob, s, m)

	authChan := make(chan bool)
	startAuthCheck(handle, prob.ContestID, prob.Index, 120, authChan)
	success := <-authChan
	if success {
		// Add to database and let user know it succeeded
	} else {
		// Tell user that the authentication failed
	}
	return nil
}

func sendAuthInstructions(prob *problem, s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := discordgo.MessageEmbed{
		Title: fmt.Sprintf("Submit a compilation error to %d%s (%s) to authenticate",
			prob.ContestID, prob.Index, prob.Name),
		URL: fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", prob.ContestID, prob.Index),
	}
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}

func getRandomProblem() (*problem, error) {
	problems, err := getProblems()
	if err != nil {
		return nil, fmt.Errorf("failed to get problems: %w", err)
	}
	if len(problems) == 0 {
		return nil, errors.New("cannot get random problem from empty slice.")
	}
	return &problems[rand.Intn(len(problems))], nil
}

func getProblems() (problems []problem, err error) {
	res, err := http.Get("https://codeforces.com/api/problemset.problems")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = res.Body.Close()
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiStruct problemList
	err = json.Unmarshal(body, &apiStruct)

	if apiStruct.Status == "FAILED" {
		return nil, errors.New(apiStruct.Comment)
	}

	return apiStruct.Result.Problems, err
}

func startAuthCheck(handle string, contID int, problemIdx string, timeoutSeconds int, resultChan chan<- bool) {
	startTime := time.Now().Unix()
	log.Printf("Starting codeforces authentication for user '%s'.", handle)
	go func() {
		for {
			// Stop if elapsed time has exceeded the timeout limit
			if time.Now().Unix()-startTime > int64(timeoutSeconds) {
				resultChan <- false
				close(resultChan)
				return
			}
			// Check every 5 seconds
			time.Sleep(time.Second * 5)
			// Get submissions and check if any of them match the criteria
			subs, err := getSubmissions(handle, 5)
			if err != nil {
				log.Printf("Failed to get submissions from user '%s': %v, retrying...", handle, err)
				continue
			}
			if checkSubmissions(subs, startTime, contID, problemIdx) {
				resultChan <- true
				close(resultChan)
				return
			}
		}
	}()
}

func checkSubmissions(subs []submission, startTime int64, contID int, problemIdx string) bool {
	for _, sub := range subs {
		// Ensure that submission was made after command was initiated
		if sub.CreationTimeSeconds-startTime < 0 {
			return false
		}
		correctID := sub.Problem.ContestID == contID
		correctIdx := sub.Problem.Index == problemIdx
		compilationError := sub.Verdict == "COMPILATION_ERROR"
		if correctID && correctIdx && compilationError {
			return true
		}
	}
	return false
}

func getSubmissions(handle string, count int) (submissions []submission, err error) {
	apiStr := fmt.Sprintf("https://codeforces.com/api/user.status?handle=%s&from=1&count=%d", handle, count)
	res, err := http.Get(apiStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = res.Body.Close()
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiStruct submissionList
	err = json.Unmarshal(body, &apiStruct)

	if apiStruct.Status == "FAILED" {
		return nil, errors.New(apiStruct.Comment)
	}

	return apiStruct.Submissions, err
}
