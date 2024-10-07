package helpers

import (
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/config"
	"net/url"
)

func MakeUrl(endpoint string, queryParams map[string]string) string {
	base, err := url.Parse(api.Client.BaseUrl)
	if err != nil {
		log.Fatalf("Invalid base URL: %v", err)
	}

	base.Path += endpoint

	params := url.Values{}
	params.Add("api_key", config.Cfg.General.ApiKey)

	for key, value := range queryParams {
		params.Add(key, value)
	}

	base.RawQuery = params.Encode()

	finalUrl := base.String()
	log.Printf("Making request: %s", finalUrl)
	return finalUrl
}
