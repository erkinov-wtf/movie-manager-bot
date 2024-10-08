package image

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/config"
	"net/http"
)

func GetImage(imageId string) (*bytes.Buffer, error) {
	url := fmt.Sprintf("%s%s", config.Cfg.Endpoints.ImageUrl, imageId)
	log.Printf("making image retrieval request: %s", url)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching image data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var imageBuffer bytes.Buffer

	_, err = io.Copy(&imageBuffer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error writing image to buffer: %w", err)
	}

	log.Printf("Image successfully retrieved in memory")

	return &imageBuffer, nil
}
