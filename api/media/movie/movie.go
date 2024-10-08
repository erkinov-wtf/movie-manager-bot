package movie

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
)

func GetMovie(movieId int) (*Movie, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v", config.Cfg.Endpoints.GetMovie, movieId), nil)

	resp, err := api.Client.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching movie data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result Movie
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	return &result, nil
}

func ShowMovie(context telebot.Context, movieData *Movie) error {
	imgBuffer, err := image.GetImage(movieData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return fmt.Errorf("could not retrieve image: %w", err)
	}

	caption := fmt.Sprintf(
		"🎬 *Title*: %v\n\n"+
			"📝 *Overview*: %v\n\n"+
			"📅 *Release Date*: %s\n\n"+
			"⏳ *Runtime*: %v minutes\n\n"+
			"🔞 *Is Adult*: %v\n\n"+
			"🔥 *Popularity*: %.2f\n\n"+
			"🌐 *Language*: %v\n\n"+
			"🎥 *Status*: %v\n",
		movieData.Title,
		movieData.Overview,
		movieData.ReleaseDate,
		movieData.Runtime,
		movieData.Adult,
		movieData.Popularity,
		movieData.OriginalLanguage,
		movieData.Status,
	)

	backBtn := &telebot.ReplyMarkup{}
	backBtn.Inline(
		backBtn.Row(backBtn.Data("🔙 Back to list", "back_to_pagination|")),
	)

	err = context.Delete()
	if err != nil {
		log.Printf("Failed to delete original message: %v", err)
	}

	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, backBtn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send movie details: %v", err)
		return fmt.Errorf("could not send movie details: %w", err)
	}

	log.Printf("Movie details successfully sent for movie ID: %d", movieData.ID)
	return nil
}
