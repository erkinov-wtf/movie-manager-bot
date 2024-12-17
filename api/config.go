package api

import (
	"crypto/tls"
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

	// Creating custom Transport with connection pooling and timeouts
	transport := &http.Transport{
		MaxIdleConns:    100,
		IdleConnTimeout: 20 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	Client = TMDBClient{
		ApiKey: apiKey,
		HttpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second, // Overall request timeout
		},
		BaseUrl:  config.Cfg.Endpoints.BaseUrl,
		ImageUrl: config.Cfg.Endpoints.ImageUrl,
	}
}

func (c *TMDBClient) NewClientWithCustomTimeout(timeout time.Duration) *http.Client {
	transport := c.HttpClient.Transport.(*http.Transport).Clone()
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
