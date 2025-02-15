package utils

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"log"
	"net/url"
)

func MakeUrl(endpoint string, queryParams map[string]string, userId int64) string {
	base, err := url.Parse(tmdb.Client.BaseUrl)
	if err != nil {
		log.Fatalf("Invalid base URL: %v", err)
	}

	base.Path += endpoint

	_, userCache := cache.UserCache.Get(userId)

	params := url.Values{}
	params.Add("api_key", userCache.ApiToken.Token)

	for key, value := range queryParams {
		params.Add(key, value)
	}

	base.RawQuery = params.Encode()

	finalUrl := base.String()
	log.Printf("making request: %s", finalUrl)
	return finalUrl
}
