package watchlist

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"log"
	"time"

	"strconv"
	"strings"
)

func (h *WatchlistHandler) WatchlistInfo(ctx telebot.Context) error {
	log.Print(messages.WatchlistCommand)

	msg, err := ctx.Bot().Send(ctx.Chat(), messages.Loading)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	btn := &telebot.ReplyMarkup{}
	btnRows := []telebot.Row{
		btn.Row(btn.Data("📺 TV Shows Watchlist", "", fmt.Sprintf("watchlist|tv|%d", msg.ID))),
		btn.Row(btn.Data("🎥 Movies Watchlist", "", fmt.Sprintf("watchlist|movie|%d", msg.ID))),
		btn.Row(btn.Data("🍿 Whole Watchlist", "", fmt.Sprintf("watchlist|full|%d", msg.ID))),
	}

	btn.Inline(btnRows...)

	_, err = ctx.Bot().Edit(msg, messages.WatchlistSelectType, btn)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleTVWatchlist(ctx telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, string(models.TVShowType))
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		log.Print("No records found")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.TVShowType))

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleMovieWatchlist(ctx telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, string(models.MovieType))
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		log.Print("No records found")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.MovieType))

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleFullWatchlist(ctx telebot.Context, msgId string) error {
	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		log.Print("No records found")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlist(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, string(models.AllType))

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) handleWatchlistInfo(ctx telebot.Context, data string) error {
	dataParts := strings.Split(data, "-")
	if len(dataParts) < 2 {
		log.Printf("Received malformed callback data for waitlist: %s", data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	movieType := dataParts[0]
	movieId := dataParts[1]

	parsedId, err := strconv.Atoi(movieId)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	if movieType == string(models.MovieType) {
		movieData, err := movie.GetMovie(h.app, parsedId, ctx.Sender().ID)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}

		err = movie.ShowMovie(h.app, ctx, movieData, false)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}

		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
	} else {
		tvData, err := tv.GetTV(h.app, parsedId, ctx.Sender().ID)
		if err != nil {
			log.Print(err)
			return err
		}

		err = tv.ShowTV(h.app, ctx, tvData, false)
		if err != nil {
			log.Print(err)
			return err
		}

		return ctx.Respond(&telebot.CallbackResponse{Text: messages.TVShowSelected})
	}
}

func (h *WatchlistHandler) handleBackToPagination(ctx telebot.Context, showType string) error {
	const (
		timeout     = 2 * time.Second
		currentPage = 1
	)

	ctxDb, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Delete the movie/show details message before any database operations
	if err := ctx.Delete(); err != nil {
		log.Printf("Failed to delete message: %v", err)
		return ctx.Send(messages.InternalError)
	}

	var response string
	var btn *telebot.ReplyMarkup

	if showType == string(models.AllType) {
		// Handle all type watchlists
		watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}

		totalItems := len(watchlists)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		// Generate pagination and response
		paginatedWatchlist := paginators.PaginateWatchlist(watchlists, currentPage)
		response, btn = paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)
	} else {
		// Handle type-specific watchlists
		watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, showType)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}

		totalItems := len(watchlists)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		// Generate pagination and response
		paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
		response, btn = paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)
	}

	// Send new message with watchlist
	if _, err := ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown); err != nil {
		log.Printf("Failed to send watchlist: %v", err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *WatchlistHandler) WatchlistCallback(ctx telebot.Context) error {
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	log.Printf("Callback data: %s", trimmed)

	if !strings.HasPrefix(trimmed, "watchlist|") {
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) < 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "tv":
		return h.handleTVWatchlist(ctx, data)

	case "movie":
		return h.handleMovieWatchlist(ctx, data)

	case "full":
		return h.handleFullWatchlist(ctx, data)

	case "info":
		return h.handleWatchlistInfo(ctx, data)

	case "next", "prev":
		paginationData := strings.Split(data, "-")
		if len(paginationData) != 2 {
			log.Printf("Received malformed callback data for watchlist pagination: %s", data)
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
		}

		watchlistType := paginationData[0]
		currentPage, err := strconv.Atoi(paginationData[1])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.InvalidPageNumber})
		}

		// Create context with timeout for database operations
		ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var totalItems int
		var totalPages int
		var response string
		var btn *telebot.ReplyMarkup

		// Handle pagination based on watchlist type
		if watchlistType == string(models.AllType) {
			// Fetch all watchlists using repository
			watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
			if err != nil {
				log.Printf("Failed to fetch watchlists: %v", err)
				return ctx.Send(messages.InternalError)
			}

			// Calculate pagination metrics
			totalItems = len(watchlists)
			totalPages = (totalItems + itemsPerPage - 1) / itemsPerPage

			// Update current page based on action
			if action == "next" && currentPage < totalPages {
				currentPage++
			} else if action == "prev" && currentPage > 1 {
				currentPage--
			}

			// Generate paginated response
			paginatedWatchlist := paginators.PaginateWatchlist(watchlists, currentPage)
			response, btn = paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, watchlistType)
		} else {
			// Fetch type-specific watchlists using repository
			watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, watchlistType)
			if err != nil {
				log.Printf("Failed to fetch watchlists with type %s: %v", watchlistType, err)
				return ctx.Send(messages.InternalError)
			}

			// Calculate pagination metrics
			totalItems = len(watchlists)
			totalPages = (totalItems + itemsPerPage - 1) / itemsPerPage

			// Update current page based on action
			if action == "next" && currentPage < totalPages {
				currentPage++
			} else if action == "prev" && currentPage > 1 {
				currentPage--
			}

			// Generate paginated response
			paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
			response, btn = paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, watchlistType)
		}

		// Update the message with new pagination
		_, err = ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
		if err != nil {
			log.Printf("Edit error: %v", err)
			if strings.Contains(err.Error(), "message is not modified") {
				return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
			}
			return ctx.Send(messages.InternalError)
		}

		return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})

	case "back_to_pagination":
		return h.handleBackToPagination(ctx, data)

	default:
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
