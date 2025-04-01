package utils

import (
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"log"
	"net/url"
)

func MakeUrl(app *appCfg.App, endpoint string, queryParams map[string]string, userId int64) string {
	base, err := url.Parse(app.TMDBClient.BaseUrl)
	if err != nil {
		log.Fatalf("Invalid base URL: %v", err)
	}

	base.Path += endpoint

	_, userCache := app.Cache.UserCache.Fetch(userId)

	params := url.Values{}
	params.Add("api_key", userCache.ApiToken.Token)

	for key, value := range queryParams {
		params.Add(key, value)
	}

	base.RawQuery = params.Encode()

	finalUrl := base.String()
	return finalUrl
}
