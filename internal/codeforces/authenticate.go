package codeforces

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuqzii/konkurransetilsynet/internal/utils"
)

type authService struct {
	db      Repository
	discord *discordgo.Session
	client  api

	timeout                 time.Duration
	maxProblemRating        uint16
	submissionCheckCount    uint16
	submissionCheckInterval time.Duration
}

// The authService uses functional options for easier configuration
type authOption func(*authService)

func newAuthService(db Repository, discord *discordgo.Session, client api, opts ...authOption) *authService {
	const (
		defaultTimeout                 time.Duration = 2 * time.Minute
		defaultMaxProblemRating        uint16        = 1500
		defaultSubmissionCheckCount    uint16        = 5
		defaultSubmissionCheckInterval time.Duration = 5 * time.Second
	)
	s := &authService{
		db:                      db,
		discord:                 discord,
		client:                  client,
		timeout:                 defaultTimeout,
		maxProblemRating:        defaultMaxProblemRating,
		submissionCheckCount:    defaultSubmissionCheckCount,
		submissionCheckInterval: defaultSubmissionCheckInterval,
	}

	// Apply each of the function options to the service
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithTimeout(timeout time.Duration) authOption {
	return func(s *authService) {
		s.timeout = timeout
	}
}

func WithMaxProblemRating(maxRating uint16) authOption {
	return func(s *authService) {
		s.maxProblemRating = maxRating
	}
}

func WithSubmissionCheckCount(cnt uint16) authOption {
	return func(s *authService) {
		s.submissionCheckCount = cnt
	}
}

func WithSubmissionCheckInterval(interval time.Duration) authOption {
	return func(s *authService) {
		s.submissionCheckInterval = interval
	}
}

func (s *authService) authCommand(args []string, m *discordgo.MessageCreate) error {
	// Ensure correct argument count
	if len(args) < 3 {
		err := utils.UnknownCommand(s.discord, m)
		return err
	}
	handle := args[2]

	log.Printf("Received Codeforces authenticate for user with handle '%s' from %s (%s).",
		handle, m.Author.ID, m.Author.Username)

	connectedHandle, err := s.db.GetConnectedCodeforces(context.TODO(), m.Author.ID)
	if !errors.Is(err, ErrUserNotConnected) {
		if err != nil {
			log.Println("Failed to check in database:", err)
		} else if connectedHandle != "" {
			err = s.onAlreadyConnected(connectedHandle, m)
			if err != nil {
				log.Println("Failed to send already connected message:", err)
			}
			return nil
		}
	}

	userExists, err := s.client.checkUserExistence(context.TODO(), handle)
	if err != nil {
		return fmt.Errorf("failed to check existence of Codeforces user '%s': %w", handle, err)
	}
	if !userExists {
		log.Printf("Codeforces user with handle '%s' does not exist.", handle)
		err = s.onUserNotExist(handle, m)
		return err
	}

	err = s.authenticate(handle, m)
	if err != nil {
		log.Println("Authentication failed:", err)
	}
	return nil
}

func (s *authService) onAlreadyConnected(handle string, m *discordgo.MessageCreate) error {
	log.Printf("Discord user %s (%s) is already connected to Codeforces user '%s'.",
		m.Author.ID, m.Author.Username, handle)
	msgStr := fmt.Sprintf("<@%s> is already connected to the Codeforces user '%s'.", m.Author.ID, handle)
	_, err := s.discord.ChannelMessageSend(m.ChannelID, msgStr)
	return err
}

func (s *authService) onUserNotExist(handle string, m *discordgo.MessageCreate) error {
	_, err := s.discord.ChannelMessageSend(m.ChannelID,
		fmt.Sprintf("Could not find a Codeforces user with the name '%s', are you sure you spelled it correctly?",
			handle))
	return err
}

func (s *authService) authenticate(handle string, m *discordgo.MessageCreate) error {
	// Get random problem with a rating <= 1500
	problems, err := s.client.getProblems(context.TODO())
	if err != nil {
		return fmt.Errorf("getting problems from Codeforces API: %w", err)
	}
	problems = filterProblems(problems, func(prob *problem) bool {
		return prob.Rating <= s.maxProblemRating
	})
	prob, err := getRandomProblem(problems)
	if err != nil {
		return err
	}

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		var testProblem = &problem{
			ContestID: 1627,
			Index:     "C",
			Name:      "Not Assigning",
			Rating:    1400,
		}
		prob = testProblem
	}

	err = s.sendAuthInstructions(prob, m)
	if err != nil {
		return fmt.Errorf("failed to send auth instructions: %w", err)
	}

	authChan := make(chan bool)
	s.startAuthCheck(handle, prob.ContestID, prob.Index, authChan)
	success := <-authChan
	if success {
		err = s.onAuthSuccess(handle, m)
		if err != nil {
			return err
		}
	} else {
		err = s.onAuthFail(handle, prob, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *authService) sendAuthInstructions(prob *problem, m *discordgo.MessageCreate) error {
	probLink := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", prob.ContestID, prob.Index)
	msgStr := fmt.Sprintf("Submit a compilation error to [%s - %d%s](%s) within 2 minutes to authenticate. <@%s>",
		prob.Name, prob.ContestID, prob.Index, probLink, m.Author.ID)
	_, err := s.discord.ChannelMessageSend(m.ChannelID, msgStr)
	return err
}

func (s *authService) onAuthSuccess(handle string, m *discordgo.MessageCreate) error {
	inDB, err := s.db.DiscordIDExists(context.TODO(), m.Author.ID)
	if err != nil {
		msgErr := s.authSuccessFailMessage(m)
		err = errors.Join(err, msgErr)
		return fmt.Errorf("failed to check if Discord ID %s exists in database: %w", m.Author.ID, err)
	}
	if inDB {
		// Cannot insert new value, need to update
		log.Printf("Discord user %s (%s) already exists in database, updating Codeforces handle to '%s',",
			m.Author.ID, m.Author.Username, handle)
		err = s.db.UpdateCodeforcesUser(context.TODO(), m.Author.ID, handle)
		if err != nil {
			msgErr := s.authSuccessFailMessage(m)
			err = errors.Join(err, msgErr)
			return err
		}
	} else {
		// Insert new column
		log.Printf("Discord user %s (%s) does not exist in database, inserting new row with Codeforces handle '%s'.",
			m.Author.ID, m.Author.Username, handle)
		err = s.db.AddCodeforcesUser(context.TODO(), m.Author.ID, handle)
		if err != nil {
			msgErr := s.authSuccessFailMessage(m)
			err = errors.Join(err, msgErr)
			return err
		}
	}

	log.Printf("Successfully authenticated discord user %s (%s) with Codeforces handle '%s'",
		m.Author.ID, m.Author.Username, handle)
	// Tell user that the authentication succeeded
	msgStr := fmt.Sprintf("Successfully authenticated discord user <@%s> with Codeforces handle '%s'.",
		m.Author.ID, handle)
	_, err = s.discord.ChannelMessageSend(m.ChannelID, msgStr)
	if err != nil {
		return fmt.Errorf("failed to send authentication success message: %w", err)
	}

	return nil
}

// Send discord message to let user know that the authentication 'succeeded', but something went wrong on our end
func (s *authService) authSuccessFailMessage(m *discordgo.MessageCreate) error {
	msgStr := fmt.Sprintf("Successfully detected a compilation error submission, "+
		"but an error occurred when storing your information. "+
		"If the problem persists please contact one of the devs or open an issue on the "+
		"[Github page](https://github.com/yuqzii/konkurransetilsynet). <@%s>", m.Author.ID)
	_, err := s.discord.ChannelMessageSend(m.ChannelID, msgStr)
	return err
}

func (s *authService) onAuthFail(handle string, prob *problem, m *discordgo.MessageCreate) error {
	// Send message explaining that the authentication failed
	probLink := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", prob.ContestID, prob.Index)
	msgStr := fmt.Sprintf("Authentication for Codeforces user with handle '%s' failed. "+
		"Did not find a compilation error submitted to [%s - %d%s](%s). <@%s>",
		handle, prob.Name, prob.ContestID, prob.Index, probLink, m.Author.ID)
	_, err := s.discord.ChannelMessageSend(m.ChannelID, msgStr)
	if err != nil {
		return fmt.Errorf("failed to send authentication failed message: %w", err)
	}
	return nil
}

func (s *authService) startAuthCheck(handle string, contID int, problemIdx string, resultChan chan<- bool) {
	startTime := time.Now().Unix()
	log.Printf("Starting Codeforces authentication check for user with handle '%s'.", handle)
	go func() {
		for {
			// Stop if elapsed time has exceeded the timeout limit
			if time.Now().Unix()-startTime > int64(s.timeout.Seconds()) {
				resultChan <- false
				close(resultChan)
				return
			}
			time.Sleep(s.submissionCheckInterval)
			// Get submissions and check if any of them match the criteria
			subs, err := s.client.getSubmissions(context.TODO(), handle, s.submissionCheckCount)
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

// Filters problems based on the f function parameter
func filterProblems(problems []problem, f func(*problem) bool) (result []problem) {
	for _, problem := range problems {
		if f(&problem) {
			result = append(result, problem)
		}
	}
	return result
}

// Returns a random problem of the problem slice provided
func getRandomProblem(problems []problem) (*problem, error) {
	if len(problems) == 0 {
		return nil, errors.New("cannot get random problem from empty slice")
	}
	prob := &problems[rand.Intn(len(problems))]
	return prob, nil
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
