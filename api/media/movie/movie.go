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
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"net/http"
)

// GetMovie fetches movie details by ID from the API.
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
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var result Movie
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	return &result, nil
}

// ShowMovie displays movie details along with an image and interactive buttons.
func ShowMovie(context telebot.Context, movieData *Movie, isMovie bool) error {
	// Retrieve movie poster image
	imgBuffer, err := image.GetImage(movieData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return fmt.Errorf("could not retrieve image: %w", err)
	}

	// Prepare movie details caption
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

	// Check if the movie is already in the user's watchlist
	var watchlist []models.Watchlist
	if err = database.DB.Where("show_api_id = ? AND user_id = ?", movieData.ID, context.Sender().ID).Find(&watchlist).Error; err != nil {
		log.Printf("Database error: %v", err)
		return context.Send("Something went wrong while checking your watchlist.")
	}

	replyMarkup := generateReplyMarkup(movieData.ID, len(watchlist) > 0, isMovie)

	// Delete the original context message
	if err = context.Delete(); err != nil {
		log.Printf("Failed to delete original message: %v", err)
	}

	// Send the movie details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send movie details: %v", err)
		return fmt.Errorf("could not send movie details: %w", err)
	}

	log.Printf("Movie details sent successfully for movie ID: %d", movieData.ID)
	return nil
}

// generateReplyMarkup generates inline keyboard buttons for the movie.
func generateReplyMarkup(movieID int64, isWatchlisted bool, isMovie bool) *telebot.ReplyMarkup {
	btn := &telebot.ReplyMarkup{}

	var backButton telebot.Btn
	if isMovie {
		backButton = btn.Data("🔙 Back to list", "movie|back_to_pagination|")
	} else {
		backButton = btn.Data("🔙 Back to list", fmt.Sprintf("watchlist|back_to_pagination|%s", models.MovieType))
	}

	watchlistButton := btn.Data(
		"🌟 Watchlist", fmt.Sprintf("movie|watchlist|%v", movieID),
	)
	watchlistedButton := btn.Data(
		"📌 Watchlisted", fmt.Sprintf("", movieID),
	)
	watchedButton := btn.Data(
		"👀 Watched", fmt.Sprintf("movie|watched|%v", movieID),
	)

	if isWatchlisted {
		btn.Inline(
			btn.Row(backButton),
			btn.Row(watchlistedButton, watchedButton),
		)
	} else {
		btn.Inline(
			btn.Row(backButton),
			btn.Row(watchlistButton, watchedButton),
		)
	}

	return btn
}
