package auth

import (
	"fmt"
	"io"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/config"
	"movie-manager-bot/helpers/utils"
	"net/http"
)

func Login() error {
	url := utils.MakeUrl(config.Cfg.Endpoints.Login, nil)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	log.Print("logged in successfully")

	return nil
}
