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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yuqzii/konkurransetilsynet/internal/database"
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
	Rating    int    `json:"rating,omitempty"`
}

func authCommand(args []string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	// Ensure correct argument count
	if len(args) < 3 {
		err := utilCommands.UnknownCommand(s, m)
		return err
	}

	log.Printf("Received codeforces authenticate for user with handle '%s'.", args[2])

	userExists, err := checkUserExistence(args[2])
	if err != nil {
		return fmt.Errorf("failed to check existance of user: %w", err)
	}
	if !userExists {
		log.Printf("Codeforces user with handle '%s' does not exist.", args[2])
		err = onUserNotExist(args[2], s, m)
		return err
	}

	// Start authentication goroutine
	go func() {
		err := authenticate(args[2], s, m)
		if err != nil {
			log.Println("Authentication failed:", err)
		}
	}()
	return nil
}

func checkUserExistence(handle string) (exists bool, err error) {
	type userInfo struct {
		Status  string `json:"status"`
		Comment string `json:"comment,omitempty"`
	}
	apiStr := fmt.Sprintf("https://codeforces.com/api/user.info?handles=%s&checkHistoricHandles=false", handle)
	res, err := http.Get(apiStr)
	if err != nil {
		return false, fmt.Errorf("failed to call codeforces user.info api: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var apiStruct userInfo
	err = json.Unmarshal(body, &apiStruct)

	exists = apiStruct.Status == "OK"
	return exists, err
}

func onUserNotExist(handle string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Could not find a codeforces user with the name '%s', are you sure you spelled it correctly?",
			handle))
	return err
}

//var testProblem *problem = &problem{
//	ContestID: 1627,
//	Index:     "C",
//	Name:      "Not Assigning",
//	Rating:    1400,
//}

func authenticate(handle string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	prob, err := getRandomProblem()
	if err != nil {
		return err
	}
	//prob = testProblem

	sendAuthInstructions(prob, s, m)

	authChan := make(chan bool)
	startAuthCheck(handle, prob.ContestID, prob.Index, 120, authChan)
	success := <-authChan
	if success {
		err = onAuthSuccess(handle, s, m)
		if err != nil {
			return err
		}
	} else {
		err = onAuthFail(handle, prob, s, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendAuthInstructions(prob *problem, s *discordgo.Session, m *discordgo.MessageCreate) error {
	probLink := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", prob.ContestID, prob.Index)
	msgStr := fmt.Sprintf("Submit a compilation error to [%s - %d%s](%s) within 2 minutes to authenticate. <@%s>",
		prob.Name, prob.ContestID, prob.Index, probLink, m.Author.ID)
	_, err := s.ChannelMessageSend(m.ChannelID, msgStr)
	return err
}

func onAuthSuccess(handle string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	// Add to database
	err := database.AddCodeforcesUser(m.Author.ID, handle)
	if err != nil {
		// Did it fail because the user already exists in the database?
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique violation error code
			log.Printf("Discord user %s (%s) trying to authenticate for codeforces handle "+
				"'%s' already exists in the database.",
				m.Author.ID, m.Author.Username, handle)
			msgStr := fmt.Sprintf("<@%s> is already connected to a codeforces user.", m.Author.ID)
			_, err = s.ChannelMessageSend(m.ChannelID, msgStr)
			return err
		}

		// Send discord message to let user know that the authentication succeeded but did not complete
		msgStr := "Authentication succeeded, but failed to store the information. " +
			"If the problem persists please contact one of the devs or open an issue on the GitHub page."
		_, msgErr := s.ChannelMessageSend(m.ChannelID, msgStr)
		if msgErr != nil {
			err = errors.Join(err, msgErr)
		}
		return fmt.Errorf("failed to add to database: %w", err)
	}

	// Tell user that the authentication succeeded
	msgStr := fmt.Sprintf("Successfully authenticated discord user <@%s> with codeforces handle '%s'", m.Author.ID, handle)
	_, err = s.ChannelMessageSend(m.ChannelID, msgStr)
	if err != nil {
		return fmt.Errorf("failed to send authentication success message: %w", err)
	}

	log.Printf("Successfully authenticated discord user %s (%s) with codeforces handle '%s'",
		m.Author.ID, m.Author.Username, handle)
	return nil
}

func onAuthFail(handle string, prob *problem, s *discordgo.Session, m *discordgo.MessageCreate) error {
	// Send message explaining that the authentication failed
	probLink := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", prob.ContestID, prob.Index)
	msgStr := fmt.Sprintf("Authentication for codeforces user with handle '%s' failed. "+
		"Did not find a compilation error submitted to [%s - %d%s](%s). <@%s>",
		handle, prob.Name, prob.ContestID, prob.Index, probLink, m.Author.ID)
	_, err := s.ChannelMessageSend(m.ChannelID, msgStr)
	if err != nil {
		return fmt.Errorf("failed to send authentication failed message: %w", err)
	}
	return nil
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
	log.Printf("Starting codeforces authentication check for user with handle '%s'.", handle)
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
		err = errors.Join(err, res.Body.Close())
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
