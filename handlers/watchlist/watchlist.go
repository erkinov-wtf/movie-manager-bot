package watchlist

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api/media/movie"
	"movie-manager-bot/api/media/tv"
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
		btn.Row(btn.Data("🍿 Whole Watchlist", "", fmt.Sprintf("watchlist|full|%d", msg.ID))),
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

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.TVShowType))

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

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.MovieType))

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *watchlistHandler) handleFullWatchlist(context telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: context.Chat()}

	var watchlist []models.Watchlist

	if err := database.DB.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
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

	currentPage := 1
	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.AllType))

	_, err := context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *watchlistHandler) handleWatchlistInfo(context telebot.Context, data string) error {
	dataParts := strings.Split(data, "-")
	if len(dataParts) < 2 {
		log.Printf("Received malformed callback data for waitlist: %s", data)
		return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
	}

	movieType := dataParts[0]
	movieId := dataParts[1]

	parsedId, err := strconv.Atoi(movieId)
	if err != nil {
		log.Print(err)
		return err
	}

	if movieType == string(models.MovieType) {
		movieData, err := movie.GetMovie(parsedId)
		if err != nil {
			log.Print(err)
			return err
		}

		err = movie.ShowMovie(context, movieData, false)
		if err != nil {
			log.Print(err)
			return err
		}

		return context.Respond(&telebot.CallbackResponse{Text: "You found the movie!"})
	} else {
		tvData, err := tv.GetTV(parsedId)
		if err != nil {
			log.Print(err)
			return err
		}

		err = tv.ShowTV(context, tvData, false)
		if err != nil {
			log.Print(err)
			return err
		}

		return context.Respond(&telebot.CallbackResponse{Text: "You found the tv!"})
	}
}

func (h *watchlistHandler) handleBackToPagination(context telebot.Context, showType string) error {
	currentPage := 1

	var watchlist []models.Watchlist
	if showType == string(models.AllType) {
		if err := database.DB.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
			log.Print(err)
			return context.Send("Something went wrong")
		}
	} else {
		if err := database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, showType).Find(&watchlist).Error; err != nil {
			log.Print(err)
			return context.Send("Something went wrong")
		}
	}

	totalItems := len(watchlist)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
	response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)

	// Delete the movie/show details message
	if err := context.Delete(); err != nil {
		log.Printf("Failed to delete message: %v", err)
	}

	// Send new message with watchlist
	_, err := context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to send watchlist: %v", err)
		return err
	}

	return nil
}

func (h *watchlistHandler) WatchlistCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	log.Printf("Callback data: %s", trimmed)

	if !strings.HasPrefix(trimmed, "watchlist|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) < 3 {
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

	case "full":
		return h.handleFullWatchlist(context, data)

	case "info":
		return h.handleWatchlistInfo(context, data)

	case "next", "prev":
		paginationData := strings.Split(data, "-")
		if len(paginationData) != 2 {
			log.Printf("Received malformed callback data for watchlist pagination: %s", data)
			return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
		}

		watchlistType := paginationData[0]
		currentPage, err := strconv.Atoi(paginationData[1])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return context.Respond(&telebot.CallbackResponse{Text: "Invalid page number"})
		}

		var watchlist []models.Watchlist
		// Determine watchlist type and fetch data
		if watchlistType == string(models.AllType) {
			if err = database.DB.Where("user_id = ?", context.Sender().ID).Find(&watchlist).Error; err != nil {
				log.Print(err)
				return context.Send("Something went wrong")
			}
		} else {
			if err = database.DB.Where("user_id = ? AND type = ?", context.Sender().ID, watchlistType).Find(&watchlist).Error; err != nil {
				log.Print(err)
				return context.Send("Something went wrong")
			}
		}

		totalItems := len(watchlist)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		if action == "next" && currentPage < totalPages {
			currentPage++
		} else if action == "prev" && currentPage > 1 {
			currentPage--
		}

		paginatedWatchlist := helpers.PaginateWatchlist(watchlist, currentPage)
		response, btn := helpers.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, watchlistType)

		_, err = context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
		if err != nil {
			log.Printf("Edit error: %v", err)
			if strings.Contains(err.Error(), "message is not modified") {
				return context.Respond(&telebot.CallbackResponse{Text: "No changes to display"})
			}
			return err
		}

		return context.Respond(&telebot.CallbackResponse{Text: "Page updated!"})

	case "back_to_pagination":
		return h.handleBackToPagination(context, data)

	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}
