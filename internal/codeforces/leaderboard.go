package codeforces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ratingChange struct {
	OldRating int `json:"oldRating"`
	NewRating int `json:"newRating"`
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

	type apiReturn struct {
		Status  string         `json:"status"`
		Comment string         `json:"comment"`
		Result  []ratingChange `json:"result"`
	}
	var api apiReturn
	err = json.Unmarshal(body, &api)
	if api.Status == "FAILED" {
		return nil, errors.New(api.Comment)
	}

	return &api.Result[0], err
}
