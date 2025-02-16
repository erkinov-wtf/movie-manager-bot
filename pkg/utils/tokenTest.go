package utils

import (
	"context"
	"encoding/json"
	"fmt"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"io"
	"log"
	"net/http"
	"time"
)

type loginResponse struct {
	IsSuccess bool   `json:"success"`
	Error     string `json:"status_message,omitempty"`
}

func TestApiToken(app *appCfg.App, token string) bool {
	if token == "" {
		log.Print("TestApiToken: empty token provided")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s?api_key=%s", app.Cfg.Endpoints.LoginUrl, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("TestApiToken: failed to create request: %v", err)
		return false
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("TestApiToken: request failed: %v", err)
		return false
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("TestApiToken: failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("TestApiToken: failed to read response body: %v", err)
		return false
	}

	var login loginResponse
	if err = json.Unmarshal(body, &login); err != nil {
		log.Printf("TestApiToken: failed to decode response: %v, body: %s", err, body)
		return false
	}

	if !login.IsSuccess {
		log.Printf("TestApiToken: validation failed: %s", login.Error)
		return false
	}

	log.Printf("TestApiToken: successfully validated token")
	return true
}
