package tv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/telebot.v3"
	"io"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/api/media/image"
	"movie-manager-bot/config"
	"movie-manager-bot/helpers/utils"
	"net/http"
	"time"
)

func GetTV(tvId int) (*TV, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v", config.Cfg.Endpoints.GetTv, tvId), nil)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result TV
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}

func GetSeason(tvId, seasonNumber int) (*Season, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v/season/%v", config.Cfg.Endpoints.GetTv, tvId, seasonNumber), nil)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result Season
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}

func ShowTV(context telebot.Context, tvData *TV) error {
	imageChan := make(chan *bytes.Buffer, 1)
	errChan := make(chan error, 1)

	// Fetch image concurrently
	go func() {
		imgBuffer, err := image.GetImage(tvData.PosterPath)
		if err != nil {
			errChan <- fmt.Errorf("could not retrieve image: %w", err)
			return
		}
		imageChan <- imgBuffer
	}()

	caption := fmt.Sprintf(
		"📺 *Name*: %v\n"+
			"📝 *Overview*: %v\n"+
			"📜 *Status*: %v\n"+
			"🔞 *Is Adult*: %v\n"+
			"🔥 *Popularity*: %v\n"+
			"🎥 *Seasons*: %v\n"+
			"#️⃣ *Episodes*: %v\n\n",
		tvData.Name,
		tvData.Overview,
		tvData.Status,
		tvData.Adult,
		tvData.Popularity,
		tvData.Seasons,
		tvData.Episodes,
	)

	backBtn := &telebot.ReplyMarkup{}
	backBtn.Inline(
		backBtn.Row(backBtn.Data("🔙 Back to list", "tv|back_to_pagination|")),
		backBtn.Row(
			backBtn.Data("📋 Watchlist", fmt.Sprintf("tv|watchlist|%v", tvData.ID)),
			backBtn.Data("✅ Watched", fmt.Sprintf("tv|select_seasons|%v", tvData.ID))),
	)

	err := context.Delete()
	if err != nil {
		log.Printf("Failed to delete original message: %v", err)
	}

	// Wait for image or error
	var imgBuffer *bytes.Buffer
	select {
	case img := <-imageChan:
		imgBuffer = img
	case err := <-errChan:
		log.Printf("Image retrieval error: %v", err)
		// Send message without image if retrieval fails
		_, sendErr := context.Bot().Send(context.Chat(), caption, backBtn, telebot.ModeMarkdown)
		return sendErr
	case <-time.After(10 * time.Second):
		log.Printf("Image retrieval timeout for TV ID: %d", tvData.ID)
		// Send message without image if retrieval times out
		_, sendErr := context.Bot().Send(context.Chat(), caption, backBtn, telebot.ModeMarkdown)
		return sendErr
	}

	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, backBtn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send tv details: %v", err)
		return fmt.Errorf("could not send tv details: %w", err)
	}

	log.Printf("Tv details successfully sent for tv ID: %d", tvData.ID)
	return nil
}
