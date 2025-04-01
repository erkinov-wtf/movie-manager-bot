package movie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/image"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"io"
	"net/http"
	"time"
)

// GetMovie fetches movie details by Id from the API.
func GetMovie(app *appCfg.App, movieId int, userId int64) (*Movie, error) {
	const op = "movie.GetMovie"
	app.Logger.Debug(op, nil, "Fetching movie details", "movie_id", movieId, "user_id", userId)

	url := utils.MakeUrl(app, fmt.Sprintf("%s/%v", app.Cfg.Endpoints.Resources.GetMovie, movieId), nil, userId)

	app.Logger.Debug(op, nil, "Making API request", "url", url)
	resp, err := app.TMDBClient.HttpClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch movie data", "movie_id", movieId, "error", err.Error())
		return nil, fmt.Errorf("error fetching movie data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Non-200 response from API",
			"movie_id", movieId, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var result Movie
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		app.Logger.Error(op, nil, "Failed to parse JSON response", "movie_id", movieId, "error", err.Error())
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	app.Logger.Info(op, nil, "Movie details fetched successfully",
		"movie_id", movieId, "title", result.Title)
	return &result, nil
}

// ShowMovie displays movie details along with an image and interactive buttons.
func ShowMovie(app *appCfg.App, ctx telebot.Context, movieData *Movie, isMovie bool) error {
	const op = "movie.ShowMovie"
	app.Logger.Info(op, ctx, "Showing movie details to user",
		"movie_id", movieData.ID, "title", movieData.Title)

	// Retrieve movie poster image
	app.Logger.Debug(op, ctx, "Retrieving movie poster image", "poster_path", movieData.PosterPath)
	imgBuffer, err := image.GetImage(app, movieData.PosterPath)
	if err != nil {
		app.Logger.Error(op, ctx, "Error retrieving image", "poster_path", movieData.PosterPath, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	// Prepare movie details caption
	app.Logger.Debug(op, ctx, "Preparing movie details caption")
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

	// Delete the original ctx message
	if err = ctx.Delete(); err != nil {
		app.Logger.Error(op, ctx, "Failed to delete original message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Check if the movie is already in the user's watchlist
	app.Logger.Debug(op, ctx, "Checking if movie is in user's watchlist",
		"movie_id", movieData.ID, "user_id", ctx.Sender().ID)
	movieExists, err := app.Repository.Watchlists.WatchlistExists(ctxDb, movieData.ID, ctx.Sender().ID, constants.MovieType)
	if err != nil {
		app.Logger.Error(op, ctx, "Failed to check watchlist status", "error", err.Error())
		return err
	}

	replyMarkup := generateReplyMarkup(movieData.ID, movieExists, isMovie)

	// Send the movie details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = ctx.Bot().Send(ctx.Chat(), imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		app.Logger.Error(op, ctx, "Failed to send movie details", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	app.Logger.Info(op, ctx, "Movie details sent successfully",
		"movie_id", movieData.ID, "title", movieData.Title, "is_watchlisted", movieExists)
	return nil
}

// generateReplyMarkup generates inline keyboard buttons for the movie.
func generateReplyMarkup(movieID int64, isWatchlisted bool, isMovie bool) *telebot.ReplyMarkup {
	btn := &telebot.ReplyMarkup{}

	var backButton telebot.Btn
	if isMovie {
		backButton = btn.Data("🔙 Back to list", "movie|back_to_pagination|")
	} else {
		backButton = btn.Data("🔙 Back to list", fmt.Sprintf("watchlist|back_to_pagination|%s", constants.MovieType))
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
