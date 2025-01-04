package search

import (
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/api"
	"github.com/erkinov-wtf/movie-manager-bot/config"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/utils"
	"io"
	"net/http"
)

func SearchMovie(movieTitle string, userId int64) (*MovieSearch, error) {
	params := map[string]string{
		"include_adult": "true",
		"query":         movieTitle,
	}
	url := utils.MakeUrl(config.Cfg.Endpoints.SearchMovie, params, userId)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching movie data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result MovieSearch
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}

func SearchTV(tvTitle string, userId int64) (*TVSearch, error) {
	params := map[string]string{
		"include_adult": "true",
		"query":         tvTitle,
	}
	url := utils.MakeUrl(config.Cfg.Endpoints.SearchTv, params, userId)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result TVSearch
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}
