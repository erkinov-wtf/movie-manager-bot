package search

import (
	"encoding/json"
	"fmt"
	"io"
	"movie-manager-bot/api"
	"movie-manager-bot/config"
	"movie-manager-bot/helpers/utils"
	"net/http"
)

func SearchMovie(movieTitle string) (*MovieSearch, error) {
	params := map[string]string{
		"include_adult": "true",
		"query":         movieTitle,
	}
	url := utils.MakeUrl(config.Cfg.Endpoints.SearchMovie, params)

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
