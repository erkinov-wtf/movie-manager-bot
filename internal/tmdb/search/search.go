package search

import (
	"encoding/json"
	"fmt"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"io"
	"net/http"
)

func SearchMovie(app *appCfg.App, movieTitle string, userId int64) (*MovieSearch, error) {
	params := map[string]string{
		"query": movieTitle,
		// new queries should be written here
	}
	url := utils.MakeUrl(app, app.Cfg.Endpoints.SearchMovie, params, userId)

	resp, err := app.TMDBClient.HttpClient.Get(url)
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

func SearchTV(app *appCfg.App, tvTitle string, userId int64) (*TVSearch, error) {
	params := map[string]string{
		"query": tvTitle,
	}
	url := utils.MakeUrl(app, app.Cfg.Endpoints.SearchTv, params, userId)

	resp, err := app.TMDBClient.HttpClient.Get(url)
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
