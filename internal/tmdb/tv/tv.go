package tv

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

func GetTV(app *appCfg.App, tvId int, userId int64) (*TV, error) {
	const op = "tv.GetTV"
	app.Logger.Debug(op, nil, "Fetching TV show details", "tv_id", tvId, "user_id", userId)

	url := utils.MakeUrl(app, fmt.Sprintf("%s/%v", app.Cfg.Endpoints.Resources.GetTV, tvId), nil, userId)
	app.Logger.Debug(op, nil, "Making API request", "url", url)

	resp, err := app.TMDBClient.HttpClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch TV show data", "tv_id", tvId, "error", err.Error())
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Received non-200 response from API",
			"tv_id", tvId, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result TV
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		app.Logger.Error(op, nil, "Failed to parse JSON response", "tv_id", tvId, "error", err.Error())
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	app.Logger.Info(op, nil, "TV show details fetched successfully",
		"tv_id", tvId, "name", result.Name, "seasons", result.Seasons)
	return &result, nil
}

func GetSeason(app *appCfg.App, tvId, seasonNumber int, userId int64) (*Season, error) {
	const op = "tv.GetSeason"
	app.Logger.Debug(op, nil, "Fetching TV season details",
		"tv_id", tvId, "season", seasonNumber, "user_id", userId)

	url := utils.MakeUrl(app, fmt.Sprintf("%s/%v/season/%v", app.Cfg.Endpoints.Resources.GetTV, tvId, seasonNumber), nil, userId)
	app.Logger.Debug(op, nil, "Making API request", "url", url)

	resp, err := app.TMDBClient.HttpClient.Get(url)
	if err != nil {
		app.Logger.Error(op, nil, "Failed to fetch season data",
			"tv_id", tvId, "season", seasonNumber, "error", err.Error())
		return nil, fmt.Errorf("error fetching tv data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error(op, nil, "Received non-200 response from API",
			"tv_id", tvId, "season", seasonNumber, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var result Season
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		app.Logger.Error(op, nil, "Failed to parse JSON response",
			"tv_id", tvId, "season", seasonNumber, "error", err.Error())
		return nil, fmt.Errorf("error parsing json response: %w", err)
	}

	app.Logger.Info(op, nil, "TV season details fetched successfully",
		"tv_id", tvId, "season", seasonNumber, "episode_count", len(result.Episodes))
	return &result, nil
}

func ShowTV(app *appCfg.App, ctx telebot.Context, tvData *TV, isTVShow bool) error {
	const op = "tv.ShowTV"
	app.Logger.Info(op, ctx, "Showing TV show details to user",
		"tv_id", tvData.Id, "name", tvData.Name)

	// Retrieve TV poster image
	app.Logger.Debug(op, ctx, "Retrieving TV poster image", "poster_path", tvData.PosterPath)
	imgBuffer, err := image.GetImage(app, tvData.PosterPath)
	if err != nil {
		app.Logger.Error(op, ctx, "Error retrieving image", "poster_path", tvData.PosterPath, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	// Prepare TV details caption
	app.Logger.Debug(op, ctx, "Preparing TV details caption")
	caption := fmt.Sprintf(
		"📺 *Name*: %v\n\n"+
			"📝 *Overview*: %v\n\n"+
			"📜 *Status*: %v\n\n"+
			"🔞 *Is Adult*: %v\n\n"+
			"🔥 *Popularity*: %.2f\n\n"+
			"🎥 *Seasons*: %v\n\n"+
			"#️⃣ *Episodes*: %v\n",
		tvData.Name,
		tvData.Overview,
		tvData.Status,
		tvData.Adult,
		tvData.Popularity,
		tvData.Seasons,
		tvData.Episodes,
	)

	// Delete the original ctx message
	if err = ctx.Delete(); err != nil {
		app.Logger.Error(op, ctx, "Failed to delete original message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Check if the tv is already in the user's watchlist
	app.Logger.Debug(op, ctx, "Checking if TV show is in user's watchlist",
		"tv_id", tvData.Id, "user_id", ctx.Sender().ID)
	tvShowExists, err := app.Repository.Watchlists.WatchlistExists(ctxDb, tvData.Id, ctx.Sender().ID, constants.TVShowType)
	if err != nil {
		app.Logger.Error(op, ctx, "Failed to check watchlist status", "error", err.Error())
		return err
	}

	replyMarkup := generateReplyMarkup(tvData.Id, tvShowExists, isTVShow)

	// Send the TV details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = ctx.Bot().Send(ctx.Chat(), imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		app.Logger.Error(op, ctx, "Failed to send TV details", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	app.Logger.Info(op, ctx, "TV show details sent successfully",
		"tv_id", tvData.Id, "name", tvData.Name, "is_watchlisted", tvShowExists)
	return nil
}

// generateReplyMarkup generates inline keyboard buttons for the TV show.
func generateReplyMarkup(TvId int64, isWatchlisted bool, isTVShow bool) *telebot.ReplyMarkup {
	btn := &telebot.ReplyMarkup{}

	var backButton telebot.Btn

	if isTVShow {
		backButton = btn.Data("🔙 Back to list", "tv|back_to_pagination|")
	} else {
		backButton = btn.Data("🔙 Back to list", fmt.Sprintf("watchlist|back_to_pagination|%s", constants.TVShowType))
	}
	watchlistButton := btn.Data(
		"🌟 Watchlist", fmt.Sprintf("tv|watchlist|%v", TvId),
	)
	watchlistedButton := btn.Data(
		"📌 Watchlisted", fmt.Sprintf("", TvId),
	)
	watchedButton := btn.Data(
		"👀 Watched", fmt.Sprintf("tv|select_seasons|%v", TvId),
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
