package tmdb

import (
	"crypto/tls"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"net/http"
	"time"
)

type Client struct {
	HttpClient *http.Client
	BaseUrl    string
	ImageUrl   string
}

func NewClient(config *config.Config) *Client {
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

	return &Client{
		HttpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second, // Overall request timeout
		},
		BaseUrl:  config.Endpoints.BaseUrl,
		ImageUrl: config.Endpoints.ImageUrl,
	}
}

func (c *Client) NewClientWithCustomTimeout(timeout time.Duration) *http.Client {
	transport := c.HttpClient.Transport.(*http.Transport).Clone()
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
