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
	const op = "search.SearchMovie"
	app.Logger.Info(op, nil, "Searching for movie", "query", movieTitle, "user_id", userId)

	params := map[string]string{
		"query": movieTitle,
		// new queries should be written here
	}

	url := utils.MakeUrl(app, fmt.Sprintf("%v%v", app.Cfg.Endpoints.Resources.Search.Prefix, app.Cfg.Endpoints.Resources.Search.Movie), params, userId)
	app.Logger.Debug(op, nil, "Making API request", "url", url)

	resp, err := app.TMDBClient.HttpClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch movie search results",
			"query", movieTitle, "error", err.Error())
		return nil, fmt.Errorf("error fetching movie data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Received non-200 response from API",
			"query", movieTitle, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	app.Logger.Debug(op, nil, "Parsing search results JSON")
	var result MovieSearch
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		app.Logger.Error(op, nil, "Failed to parse JSON response",
			"query", movieTitle, "error", err.Error())
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	app.Logger.Info(op, nil, "Movie search completed successfully",
		"query", movieTitle, "results_count", result.TotalResults)
	return &result, nil
}

func SearchTV(app *appCfg.App, tvTitle string, userId int64) (*TVSearch, error) {
	const op = "search.SearchTV"
	app.Logger.Info(op, nil, "Searching for TV show", "query", tvTitle, "user_id", userId)

	params := map[string]string{
		"query": tvTitle,
	}

	url := utils.MakeUrl(app, fmt.Sprintf("%v%v", app.Cfg.Endpoints.Resources.Search.Prefix, app.Cfg.Endpoints.Resources.Search.TV), params, userId)
	app.Logger.Debug(op, nil, "Making API request", "url", url)

	resp, err := app.TMDBClient.HttpClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch TV search results",
			"query", tvTitle, "error", err.Error())
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Received non-200 response from API",
			"query", tvTitle, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	app.Logger.Debug(op, nil, "Parsing search results JSON")
	var result TVSearch
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		app.Logger.Error(op, nil, "Failed to parse JSON response",
			"query", tvTitle, "error", err.Error())
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	app.Logger.Info(op, nil, "TV show search completed successfully",
		"query", tvTitle, "results_count", result.TotalResults)
	return &result, nil
}
