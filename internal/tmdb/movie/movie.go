package movie

import (
	"bytes"
	"encoding/json"
	"fmt"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/image"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"io"
	"log"

	"net/http"
)

// GetMovie fetches movie details by ID from the API.
func GetMovie(app *appCfg.App, movieId int, userId int64) (*Movie, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v", app.Cfg.Endpoints.GetMovie, movieId), nil, userId)

	resp, err := app.TMDBClient.HttpClient.Get(url)
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
func ShowMovie(app *appCfg.App, context telebot.Context, movieData *Movie, isMovie bool) error {
	// Retrieve movie poster image
	imgBuffer, err := image.GetImage(app, movieData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return context.Send(messages.InternalError)
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
	if err = app.Database.Where("show_api_id = ? AND user_id = ?", movieData.ID, context.Sender().ID).Find(&watchlist).Error; err != nil {
		log.Printf("Database error: %v", err)
		return context.Send(messages.WatchlistCheckError)
	}

	replyMarkup := generateReplyMarkup(movieData.ID, len(watchlist) > 0, isMovie)

	// Delete the original context message
	if err = context.Delete(); err != nil {
		log.Printf("Failed to delete original message: %v", err)
		return context.Send(messages.InternalError)
	}

	// Send the movie details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send movie details: %v", err)
		return context.Send(messages.InternalError)
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
