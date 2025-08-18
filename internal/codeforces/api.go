package codeforces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/time/rate"
)

type api interface {
	getContests(ctx context.Context) ([]*contest, error)
	getProblems(ctx context.Context) ([]problem, error)
	getSubmissions(ctx context.Context, handle string, count uint16) ([]submission, error)
	getRating(ctx context.Context, handle string) (*ratingChange, error)
	hasUpdatedRating(ctx context.Context, c *contest) (bool, error)
	checkUserExistence(ctx context.Context, handle string) (bool, error)
}

type client struct {
	client  *http.Client
	limiter *rate.Limiter
	url     string
}

func NewClient(httpClient *http.Client, requestsPerSecond float64, burst int, url string) *client {
	return &client{
		client:  httpClient,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
		url:     url,
	}
}

func (c *client) makeRequest(ctx context.Context, method, endpoint string) (*http.Response, error) {
	// Wait for rate limiter permission
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.client.Do(req)
}

var ErrNoRating = errors.New("the user does not have a rating")
var ErrCodeforcesIssue = errors.New("issue with the Codeforces server")
var ErrClientIssue = errors.New("(skill) issue with our client")

type contest struct {
	ID                    uint32 `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Phase                 string `json:"phase"`
	Frozen                bool   `json:"frozen"`
	DurationSeconds       uint32 `json:"durationSeconds"`
	Description           string `json:"description,omitempty"`
	Difficulty            uint8  `json:"difficulty,omitempty"`
	Kind                  string `json:"kind,omitempty"`
	Season                string `json:"season,omitempty"`
	StartTimeSeconds      uint32 `json:"startTimeSeconds,omitempty"`
	RelativeTimeSeconds   int32  `json:"relativeTimeSeconds,omitempty"`
	PreparedBy            string `json:"preparedBy,omitempty"`
	Country               string `json:"country,omitempty"`
	City                  string `json:"city,omitempty"`
	IcpcRegion            string `json:"icpcRegion,omitempty"`
	WebsiteURL            string `json:"websiteUrl,omitempty"`
	FreezeDurationSeconds uint32 `json:"freezeDurationSeconds,omitempty"`
}

type problem struct {
	ContestID int    `json:"contestId"`
	Index     string `json:"index"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Rating    uint16 `json:"rating,omitempty"`
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

type ratingChange struct {
	Handle    string `json:"handle"`
	OldRating int    `json:"oldRating"`
	NewRating int    `json:"newRating"`
	discordID string
}

type ratingChangeAPIReturn struct {
	Status  string         `json:"status"`
	Result  []ratingChange `json:"result"`
	Comment string         `json:"comment"`
}

// Gets all contests from the Codeforces API
func (c *client) getContests(ctx context.Context) (contests []*contest, err error) {
	endpoint := "contest.list"
	res, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = res.Body.Close()
		}
	}()

	if err = responseCodeCheck(res); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var contestList struct {
		Status   string     `json:"status"`
		Contests []*contest `json:"result"`
		Comment  string     `json:"comment,omitempty"`
	}
	err = json.Unmarshal(body, &contestList)

	if contestList.Status == "FAILED" {
		return nil, errors.New(contestList.Comment)
	}

	return contestList.Contests, err
}

// Returns all problems from the Codeforces API
func (c *client) getProblems(ctx context.Context) (problems []problem, err error) {
	endpoint := "problemset.problems"
	res, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = res.Body.Close()
		}
	}()

	if err = responseCodeCheck(res); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiStruct struct {
		Status string `json:"status"`
		Result struct {
			Problems []problem `json:"problems"`
		} `json:"result"`
		Comment string `json:"comment,omitempty"`
	}
	err = json.Unmarshal(body, &apiStruct)

	if apiStruct.Status == "FAILED" {
		return nil, errors.New(apiStruct.Comment)
	}

	return apiStruct.Result.Problems, err
}

func (c *client) getSubmissions(ctx context.Context, handle string,
	count uint16) (submissions []submission, err error) {

	endpoint := "user.status?"
	params := url.Values{}
	params.Set("handle", handle)
	params.Set("from", "1")
	params.Set("count", strconv.FormatUint(uint64(count), 10))
	res, err := c.makeRequest(ctx, "GET", endpoint+params.Encode())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if err = responseCodeCheck(res); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiStruct struct {
		Status      string       `json:"status"`
		Submissions []submission `json:"result"`
		Comment     string       `json:"comment,omitempty"`
	}
	err = json.Unmarshal(body, &apiStruct)

	if apiStruct.Status == "FAILED" {
		return nil, errors.New(apiStruct.Comment)
	}

	return apiStruct.Submissions, err
}

func (c *client) getRating(ctx context.Context, handle string) (rating *ratingChange, err error) {
	endpoint := "user.rating?"
	params := url.Values{}
	params.Set("handle", handle)
	res, err := c.makeRequest(ctx, "GET", endpoint+params.Encode())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if err = responseCodeCheck(res); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var apiReturn ratingChangeAPIReturn
	err = json.Unmarshal(body, &apiReturn)
	if apiReturn.Status == "FAILED" {
		return nil, errors.New(apiReturn.Comment)
	}

	if len(apiReturn.Result) == 0 {
		return nil, ErrNoRating
	}

	return &apiReturn.Result[len(apiReturn.Result)-1], err
}

func (c *client) hasUpdatedRating(ctx context.Context, contest *contest) (updated bool, err error) {
	endpoint := "contest.ratingChanges?"
	params := url.Values{}
	params.Set("contestId", strconv.FormatUint(uint64(contest.ID), 10))
	res, err := c.makeRequest(ctx, "GET", endpoint+params.Encode())
	if err != nil {
		return false, fmt.Errorf("getting rating change from Codeforces api: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if err = responseCodeCheck(res); err != nil {
		return false, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var api ratingChangeAPIReturn
	err = json.Unmarshal(body, &api)
	if api.Status == "FAILED" {
		return false, errors.New(api.Comment)
	}

	// Codeforces returns an empty result before the ratings have updated
	return len(api.Result) != 0, err
}

func (c *client) checkUserExistence(ctx context.Context, handle string) (exists bool, err error) {
	type userInfo struct {
		Status  string `json:"status"`
		Comment string `json:"comment,omitempty"`
	}

	endpoint := "user.info?"
	params := url.Values{}
	params.Set("handles", handle)
	params.Set("checkHistoricHandles", "false")
	res, err := c.makeRequest(ctx, "GET", endpoint+params.Encode())
	if err != nil {
		return false, fmt.Errorf("failed to call Codeforces user.info api: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if err = responseCodeCheck(res); err != nil {
		return false, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var apiStruct userInfo
	err = json.Unmarshal(body, &apiStruct)
	if err != nil {
		return false, err
	}

	exists = apiStruct.Status == "OK"
	return exists, err
}

func responseCodeCheck(res *http.Response) error {
	switch res.StatusCode / 100 {
	case 4:
		return fmt.Errorf("%w: %s", ErrClientIssue, res.Status)
	case 5:
		return fmt.Errorf("%w: %s", ErrCodeforcesIssue, res.Status)
	default:
		return nil
	}
}
