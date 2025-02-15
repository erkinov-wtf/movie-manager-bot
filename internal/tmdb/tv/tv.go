package tv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/image"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"io"
	"log"

	"net/http"
)

func GetTV(tvId int, userId int64) (*TV, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v", config.Cfg.Endpoints.GetTv, tvId), nil, userId)

	resp, err := tmdb.Client.HttpClient.Get(url)
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

func GetSeason(tvId, seasonNumber int, userId int64) (*Season, error) {
	url := utils.MakeUrl(fmt.Sprintf("%s/%v/season/%v", config.Cfg.Endpoints.GetTv, tvId, seasonNumber), nil, userId)

	resp, err := tmdb.Client.HttpClient.Get(url)
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

func ShowTV(context telebot.Context, tvData *TV, isTVShow bool) error {
	// Retrieve TV poster image
	imgBuffer, err := image.GetImage(tvData.PosterPath)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return context.Send(messages.InternalError)
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

	// Check if the movie is already in the user's watchlist
	var watchlist []models.Watchlist
	if err = database.DB.Where("show_api_id = ? AND user_id = ?", tvData.ID, context.Sender().ID).Find(&watchlist).Error; err != nil {
		log.Printf("Database error: %v", err)
		return context.Send(messages.WatchlistCheckError)
	}

	replyMarkup := generateReplyMarkup(tvData.ID, len(watchlist) > 0, isTVShow)

	// Delete the original context message
	if err = context.Delete(); err != nil {
		log.Printf("Failed to delete original message: %v", err)
		return context.Send(messages.InternalError)
	}

	// Send the TV details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = context.Bot().Send(context.Chat(), imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send TV details: %v", err)
		return context.Send(messages.InternalError)
	}

	log.Printf("TV details sent successfully for TV ID: %d", tvData.ID)
	return nil
}

// generateReplyMarkup generates inline keyboard buttons for the TV show.
func generateReplyMarkup(TVID int64, isWatchlisted bool, isTVShow bool) *telebot.ReplyMarkup {
	btn := &telebot.ReplyMarkup{}

	var backButton telebot.Btn

	if isTVShow {
		backButton = btn.Data("🔙 Back to list", "tv|back_to_pagination|")
	} else {
		backButton = btn.Data("🔙 Back to list", fmt.Sprintf("watchlist|back_to_pagination|%s", models.TVShowType))
	}
	watchlistButton := btn.Data(
		"🌟 Watchlist", fmt.Sprintf("tv|watchlist|%v", TVID),
	)
	watchlistedButton := btn.Data(
		"📌 Watchlisted", fmt.Sprintf("", TVID),
	)
	watchedButton := btn.Data(
		"👀 Watched", fmt.Sprintf("tv|select_seasons|%v", TVID),
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
