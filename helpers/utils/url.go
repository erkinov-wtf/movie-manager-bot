package utils

import (
	"github.com/erkinov-wtf/movie-manager-bot/api"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"log"
	"net/url"
)

func MakeUrl(endpoint string, queryParams map[string]string, userId int64) string {
	base, err := url.Parse(api.Client.BaseUrl)
	if err != nil {
		log.Fatalf("Invalid base URL: %v", err)
	}

	base.Path += endpoint

	_, _, token := cache.UserCache.Get(userId)

	params := url.Values{}
	params.Add("api_key", token.Token)

	for key, value := range queryParams {
		params.Add(key, value)
	}

	base.RawQuery = params.Encode()

	finalUrl := base.String()
	log.Printf("making request: %s", finalUrl)
	return finalUrl
}
