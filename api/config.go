package api

import (
	"log"
	"movie-manager-bot/config"
	"net/http"
	"time"
)

var Client TMDBClient

type TMDBClient struct {
	ApiKey     string
	HttpClient *http.Client
	BaseUrl    string
	ImageUrl   string
}

func NewClient() {
	apiKey := config.Cfg.General.ApiKey
	if apiKey == "" {
		log.Fatal("Client API Key is not found")
	}

	Client = TMDBClient{
		ApiKey: apiKey,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseUrl:  config.Cfg.Endpoints.BaseUrl,
		ImageUrl: config.Cfg.Endpoints.ImageUrl,
	}
}
