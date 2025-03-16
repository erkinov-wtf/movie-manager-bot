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
	"log"
	"time"

	"net/http"
)

func GetTV(app *appCfg.App, tvId int, userId int64) (*TV, error) {
	url := utils.MakeUrl(app, fmt.Sprintf("%s/%v", app.Cfg.Endpoints.GetTv, tvId), nil, userId)

	resp, err := app.TMDBClient.HttpClient.Get(url)
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

func GetSeason(app *appCfg.App, tvId, seasonNumber int, userId int64) (*Season, error) {
	url := utils.MakeUrl(app, fmt.Sprintf("%s/%v/season/%v", app.Cfg.Endpoints.GetTv, tvId, seasonNumber), nil, userId)

	resp, err := app.TMDBClient.HttpClient.Get(url)
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

func ShowTV(app *appCfg.App, ctx telebot.Context, tvData *TV, isTVShow bool) error {
	// Retrieve TV poster image
	imgBuffer, err := image.GetImage(app, tvData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return ctx.Send(messages.InternalError)
	}

	// Prepare TV details caption
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
		log.Printf("Failed to delete original message: %v", err)
		return ctx.Send(messages.InternalError)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Check if the tv is already in the user's watchlist
	tvShowExists, err := app.Repository.Watchlists.WatchlistExists(ctxDb, tvData.Id, ctx.Sender().ID, constants.TVShowType)
	if err != nil {
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
		log.Printf("Failed to send TV details: %v", err)
		return ctx.Send(messages.InternalError)
	}

	log.Printf("TV details sent successfully for TV Id: %d", tvData.Id)
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
