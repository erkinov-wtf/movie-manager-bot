package watchlist

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/helpers"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

func (*watchlistHandler) WatchlistInfo(context telebot.Context) error {
	log.Print("/w command received")

	msg, err := context.Bot().Send(context.Chat(), "Loading...")
	if err != nil {
		log.Print(err)
		return err
	}

	btn := &telebot.ReplyMarkup{}
	btnRows := []telebot.Row{
		btn.Row(btn.Data("📺 TV Shows Watchlist", "", fmt.Sprintf("watchlist|tv|%d", msg.ID))),
		btn.Row(btn.Data("🎥 Movies Watchlist", "", fmt.Sprintf("watchlist|movie|%d", msg.ID))),
		btn.Row(btn.Data("🍿 Whole Watchlist", "", fmt.Sprintf("watchlist|whole|%d", msg.ID))),
	}

	btn.Inline(btnRows...)

	_, err = context.Bot().Edit(msg, "Which type of watchlist do you want?", btn)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *watchlistHandler) handleTVWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, models.TVShowType).Find(&watchlist).Error; err != nil {
		log.Print(err)
		return context.Send("Something went wrong")
	}

	if len(watchlist) == 0 {
		log.Print("No records found")
		_, err := context.Bot().Edit(msg, "No records found", telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	const itemsPerPage = 3
	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems)

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *watchlistHandler) handleMovieWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, models.MovieType).Find(&watchlist).Error; err != nil {
		log.Print(err)
		return context.Send("Something went wrong")
	}

	if len(watchlist) == 0 {
		log.Print("No records found")
		_, err := context.Bot().Edit(msg, "No records found", telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	const itemsPerPage = 3
	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems)

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *watchlistHandler) handleWatchlistInfo(context telebot.Context, data string) error {
	return context.Send("You found this part cool " + data) //todo implement fully
}

func (h *watchlistHandler) handlePagination(context telebot.Context, watchlist []models.Watchlist, currentPage, totalPages int, action string) error {
	if action == "next" && currentPage < totalPages {
		currentPage++
	} else if action == "prev" && currentPage > 1 {
		currentPage--
	}

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)

	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, len(watchlist))

	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return context.Respond(&telebot.CallbackResponse{Text: "Page updated!"})
}

func (h *watchlistHandler) WatchlistCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "watchlist|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "tv":
		return h.handleTVWatchlist(context, data)

	case "movie":
		return h.handleMovieWatchlist(context, data)

	case "info":
		return h.handleWatchlistInfo(context, data)

	case "next", "prev":
		watchlistType := dataParts[2]
		var watchlist []models.Watchlist

		if watchlistType == "tv" {
			_ = database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, models.TVShowType).Find(&watchlist).Error
		} else if watchlistType == "movie" {
			_ = database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, models.MovieType).Find(&watchlist).Error
		}

		totalItems := len(watchlist)
		itemsPerPage := 3
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
		currentPage, _ := strconv.Atoi(callback.Message.Payload)

		return h.handlePagination(context, watchlist, currentPage, totalPages, action)

	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}
