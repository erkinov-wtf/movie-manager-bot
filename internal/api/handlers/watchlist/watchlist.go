package watchlist

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"log"

	"strconv"
	"strings"
)

func (h *WatchlistHandler) WatchlistInfo(context telebot.Context) error {
	log.Print(messages.WatchlistCommand)

	msg, err := context.Bot().Send(context.Chat(), messages.Loading)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	btn := &telebot.ReplyMarkup{}
	btnRows := []telebot.Row{
		btn.Row(btn.Data("📺 TV Shows Watchlist", "", fmt.Sprintf("watchlist|tv|%d", msg.ID))),
		btn.Row(btn.Data("🎥 Movies Watchlist", "", fmt.Sprintf("watchlist|movie|%d", msg.ID))),
		btn.Row(btn.Data("🍿 Whole Watchlist", "", fmt.Sprintf("watchlist|full|%d", msg.ID))),
	}

	btn.Inline(btnRows...)

	_, err = context.Bot().Edit(msg, messages.WatchlistSelectType, btn)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleTVWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := h.Database.Where("user_id = ? AND type = ?", context.Sender().ID, models.TVShowType).Find(&watchlist).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	if len(watchlist) == 0 {
		log.Print("No records found")
		_, err := context.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlist(watchlist, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.TVShowType))

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleMovieWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := h.Database.Where("user_id = ? AND type = ?", context.Sender().ID, models.MovieType).Find(&watchlist).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	if len(watchlist) == 0 {
		log.Print("No records found")
		_, err := context.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlist(watchlist, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.MovieType))

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleFullWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := h.Database.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	if len(watchlist) == 0 {
		log.Print("No records found")
		_, err := context.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlist(watchlist, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.AllType))

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleWatchlistInfo(context telebot.Context, data string) error {
	dataParts := strings.Split(data, "-")
	if len(dataParts) < 2 {
		log.Printf("Received malformed callback data for waitlist: %s", data)
		return context.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	movieType := dataParts[0]
	movieId := dataParts[1]

	parsedId, err := strconv.Atoi(movieId)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	if movieType == string(models.MovieType) {
		movieData, err := movie.GetMovie(parsedId, context.Sender().ID)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}

		err = movie.ShowMovie(context, movieData, false)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}

		return context.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
	} else {
		tvData, err := tv.GetTV(parsedId, context.Sender().ID)
		if err != nil {
			log.Print(err)
			return err
		}

		err = tv.ShowTV(context, tvData, false)
		if err != nil {
			log.Print(err)
			return err
		}

		return context.Respond(&telebot.CallbackResponse{Text: messages.TVShowSelected})
	}
}

func (h *WatchlistHandler) handleBackToPagination(context telebot.Context, showType string) error {
	currentPage := 1

	var watchlist []models.Watchlist
	if showType == string(models.AllType) {
		if err := h.Database.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
	} else {
		if err := h.Database.Where("user_id = ? AND type = ?", context.Sender().ID, showType).Find(&watchlist).Error; err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
	}

	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlist(watchlist, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)

	// Delete the movie/show details message
	if err := context.Delete(); err != nil {
		log.Printf("Failed to delete message: %v", err)
		return context.Send(messages.InternalError)
	}

	// Send new message with watchlist
	_, err := context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send watchlist: %v", err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) WatchlistCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	log.Printf("Callback data: %s", trimmed)

	if !strings.HasPrefix(trimmed, "watchlist|") {
		return context.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) < 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "tv":
		return h.handleTVWatchlist(context, data)

	case "movie":
		return h.handleMovieWatchlist(context, data)

	case "full":
		return h.handleFullWatchlist(context, data)

	case "info":
		return h.handleWatchlistInfo(context, data)

	case "next", "prev":
		paginationData := strings.Split(data, "-")
		if len(paginationData) != 2 {
			log.Printf("Received malformed callback data for watchlist pagination: %s", data)
			return context.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
		}

		watchlistType := paginationData[0]
		currentPage, err := strconv.Atoi(paginationData[1])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return context.Respond(&telebot.CallbackResponse{Text: messages.InvalidPageNumber})
		}

		var watchlist []models.Watchlist
		// Determine watchlist type and fetch data
		if watchlistType == string(models.AllType) {
			if err = h.Database.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
				log.Print(err)
				return context.Send(messages.InternalError)
			}
		} else {
			if err = h.Database.Where("user_id = ? AND type = ?", context.Sender().ID, watchlistType).Find(&watchlist).Error; err != nil {
				log.Print(err)
				return context.Send(messages.InternalError)
			}
		}

		totalItems := len(watchlist)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		if action == "next" && currentPage < totalPages {
			currentPage++
		} else if action == "prev" && currentPage > 1 {
			currentPage--
		}

		paginatedWatchlist := paginators.PaginateWatchlist(watchlist, currentPage)
		response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, watchlistType)

		_, err = context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
		if err != nil {
			log.Printf("Edit error: %v", err)
			if strings.Contains(err.Error(), "message is not modified") {
				return context.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
			}
			return context.Send(messages.InternalError)
		}

		return context.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})

	case "back_to_pagination":
		return h.handleBackToPagination(context, data)

	default:
		return context.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
