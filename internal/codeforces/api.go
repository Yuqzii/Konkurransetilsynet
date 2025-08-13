package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

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
	Rating    int    `json:"rating,omitempty"`
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
func getContests() (contests []contest, err error) {
	res, err := http.Get("https://codeforces.com/api/contest.list")
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

	var contestList struct {
		Status   string    `json:"status"`
		Contests []contest `json:"result"`
		Comment  string    `json:"comment,omitempty"`
	}
	err = json.Unmarshal(body, &contestList)

	if contestList.Status == "FAILED" {
		return nil, errors.New(contestList.Comment)
	}

	return contestList.Contests, err
}

// Returns all problems from the Codeforces API
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

func getRating(handle string) (rating *ratingChange, err error) {
	apiStr := fmt.Sprintf("https://codeforces.com/api/user.rating?handle=%s", handle)
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

	var apiReturn ratingChangeAPIReturn
	err = json.Unmarshal(body, &apiReturn)
	if apiReturn.Status == "FAILED" {
		return nil, errors.New(apiReturn.Comment)
	}

	return &apiReturn.Result[len(apiReturn.Result)-1], err
}

func hasUpdatedRating(c *contest) (updated bool, err error) {
	apiStr := fmt.Sprintf("https://codeforces.com/api/contest.ratingChanges?contestId=%d", c.ID)
	res, err := http.Get(apiStr)
	if err != nil {
		return false, fmt.Errorf("getting rating change from Codeforces api: %w", err)
	}
	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

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
