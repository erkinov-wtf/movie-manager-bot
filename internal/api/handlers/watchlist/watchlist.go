package watchlist

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"strconv"
	"strings"
	"time"
)

func (h *WatchlistHandler) WatchlistInfo(ctx telebot.Context) error {
	const op = "watchlist.WatchlistInfo"
	h.app.Logger.Info(op, ctx, "Watchlist command received")

	msg, err := ctx.Bot().Send(ctx.Chat(), messages.Loading)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send loading message", "error", err.Error())
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
		h.app.Logger.Error(op, ctx, "Failed to edit message with watchlist type options", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Watchlist type selection displayed successfully")
	return nil
}

func (h *WatchlistHandler) handleTVWatchlist(ctx telebot.Context, msgId string) error {
	const op = "watchlist.handleTVWatchlist"
	h.app.Logger.Info(op, ctx, "Fetching TV show watchlist", "message_id", msgId)

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving TV watchlist from database")
	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, constants.TVShowType)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve TV watchlists", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		h.app.Logger.Info(op, ctx, "No TV shows found in watchlist")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to edit message for empty watchlist", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	h.app.Logger.Debug(op, ctx, "Generating TV watchlist pagination",
		"items_count", totalItems, "total_pages", totalPages)
	paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, constants.TVShowType)

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to edit message with TV watchlist", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "TV watchlist displayed successfully", "items_count", totalItems)
	return nil
}

func (h *WatchlistHandler) handleMovieWatchlist(ctx telebot.Context, msgId string) error {
	const op = "watchlist.handleMovieWatchlist"
	h.app.Logger.Info(op, ctx, "Fetching movie watchlist", "message_id", msgId)

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving movie watchlist from database")
	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, constants.MovieType)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve movie watchlists", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		h.app.Logger.Info(op, ctx, "No movies found in watchlist")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to edit message for empty watchlist", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, constants.MovieType)

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to edit message with movie watchlist", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Movie watchlist displayed successfully", "items_count", totalItems)
	return nil
}

func (h *WatchlistHandler) handleFullWatchlist(ctx telebot.Context, msgId string) error {
	const op = "watchlist.handleFullWatchlist"
	h.app.Logger.Info(op, ctx, "Fetching full watchlist", "message_id", msgId)

	msgID, _ := strconv.Atoi(msgId)
	msg := &telebot.Message{ID: msgID, Chat: ctx.Chat()}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Retrieving complete watchlist from database")
	watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve watchlists", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if len(watchlists) == 0 {
		h.app.Logger.Info(op, ctx, "No items found in watchlist")
		_, err = ctx.Bot().Edit(msg, messages.NoWatchlistData, telebot.ModeMarkdown)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to edit message for empty watchlist", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	currentPage := 1
	totalItems := len(watchlists)
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

	h.app.Logger.Debug(op, ctx, "Generating full watchlist pagination",
		"items_count", totalItems, "total_pages", totalPages)
	paginatedWatchlist := paginators.PaginateWatchlist(watchlists, currentPage)
	response, btn := paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, constants.AllType)

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to edit message with full watchlist", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Full watchlist displayed successfully", "items_count", totalItems)
	return nil
}

func (h *WatchlistHandler) handleWatchlistInfo(ctx telebot.Context, data string) error {
	const op = "watchlist.handleWatchlistInfo"
	h.app.Logger.Info(op, ctx, "Fetching watchlist item details", "data", data)

	dataParts := strings.Split(data, "-")
	if len(dataParts) < 2 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "data", data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	movieType := dataParts[0]
	movieId := dataParts[1]
	h.app.Logger.Debug(op, ctx, "Processing watchlist item", "type", movieType, "id", movieId)

	parsedId, err := strconv.Atoi(movieId)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse item ID", "id", movieId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if movieType == constants.MovieType {
		h.app.Logger.Debug(op, ctx, "Retrieving movie details", "movie_id", parsedId)
		movieData, err := movie.GetMovie(h.app, parsedId, ctx.Sender().ID)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to get movie data", "movie_id", parsedId, "error", err.Error())
			return ctx.Send(messages.InternalError)
		}

		h.app.Logger.Debug(op, ctx, "Displaying movie details", "title", movieData.Title)
		err = movie.ShowMovie(h.app, ctx, movieData, false)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to show movie details", "movie_id", parsedId, "error", err.Error())
			return ctx.Send(messages.InternalError)
		}

		h.app.Logger.Info(op, ctx, "Movie details displayed successfully", "movie_id", parsedId, "title", movieData.Title)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
	} else {
		h.app.Logger.Debug(op, ctx, "Retrieving TV show details", "tv_id", parsedId)
		tvData, err := tv.GetTV(h.app, parsedId, ctx.Sender().ID)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to get TV show data", "tv_id", parsedId, "error", err.Error())
			return err
		}

		h.app.Logger.Debug(op, ctx, "Displaying TV show details", "name", tvData.Name)
		err = tv.ShowTV(h.app, ctx, tvData, false)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to show TV show details", "tv_id", parsedId, "error", err.Error())
			return err
		}

		h.app.Logger.Info(op, ctx, "TV show details displayed successfully", "tv_id", parsedId, "name", tvData.Name)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.TVShowSelected})
	}
}

func (h *WatchlistHandler) handleBackToPagination(ctx telebot.Context, showType string) error {
	const op = "watchlist.handleBackToPagination"
	h.app.Logger.Info(op, ctx, "Returning to paginated watchlist", "type", showType)

	const (
		timeout     = 2 * time.Second
		currentPage = 1
	)

	ctxDb, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Delete the movie/show details message before any database operations
	h.app.Logger.Debug(op, ctx, "Deleting current message")
	if err := ctx.Delete(); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to delete message", "error", err.Error())
		return ctx.Send(messages.WatchlistCheckError)
	}

	var response string
	var btn *telebot.ReplyMarkup

	if showType == constants.AllType {
		// Handle all type watchlists
		h.app.Logger.Debug(op, ctx, "Retrieving complete watchlist")
		watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to get watchlists", "error", err.Error())
			return ctx.Send(messages.WatchlistCheckError)
		}

		totalItems := len(watchlists)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		// Generate pagination and response
		h.app.Logger.Debug(op, ctx, "Generating full watchlist response",
			"items_count", totalItems, "pages", totalPages)
		paginatedWatchlist := paginators.PaginateWatchlist(watchlists, currentPage)
		response, btn = paginators.GenerateWatchlistResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)
	} else {
		// Handle type-specific watchlists
		h.app.Logger.Debug(op, ctx, "Retrieving type-specific watchlist", "type", showType)
		watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, showType)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to get watchlists by type", "type", showType, "error", err.Error())
			return ctx.Send(messages.WatchlistCheckError)
		}

		totalItems := len(watchlists)
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage

		// Generate pagination and response
		h.app.Logger.Debug(op, ctx, "Generating type-specific watchlist response",
			"items_count", totalItems, "pages", totalPages, "type", showType)
		paginatedWatchlist := paginators.PaginateWatchlistWithType(watchlists, currentPage)
		response, btn = paginators.GenerateWatchlistWithTypeResponse(&paginatedWatchlist, currentPage, totalPages, totalItems, showType)
	}

	// Send new message with watchlist
	h.app.Logger.Debug(op, ctx, "Sending watchlist message")
	if _, err := ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send watchlist", "error", err.Error())
		return ctx.Send(messages.WatchlistCheckError)
	}

	h.app.Logger.Info(op, ctx, "Successfully returned to watchlist")
	return nil
}

func (h *WatchlistHandler) WatchlistCallback(ctx telebot.Context) error {
	const op = "watchlist.WatchlistCallback"
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	h.app.Logger.Info(op, ctx, "Processing watchlist callback", "callback_data", trimmed)

	if !strings.HasPrefix(trimmed, "watchlist|") {
		h.app.Logger.Warning(op, ctx, "Invalid callback prefix", "callback_data", trimmed)
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) < 3 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "callback_data", callback.Data,
			"parts_count", len(dataParts))
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]
	h.app.Logger.Debug(op, ctx, "Processing callback action", "action", action, "data", data)

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
			h.app.Logger.Warning(op, ctx, "Malformed pagination data", "data", data)
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
		}

		watchlistType := paginationData[0]
		currentPage, err := strconv.Atoi(paginationData[1])
		if err != nil {
			h.app.Logger.Error(op, ctx, "Invalid page number", "page", paginationData[1], "error", err.Error())
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
		if watchlistType == constants.AllType {
			// Fetch all watchlists using repository
			h.app.Logger.Debug(op, ctx, "Fetching complete watchlist for pagination")
			watchlists, err := h.app.Repository.Watchlists.GetUserWatchlists(ctxDb, ctx.Sender().ID)
			if err != nil {
				h.app.Logger.Error(op, ctx, "Failed to fetch watchlists", "error", err.Error())
				return ctx.Send(messages.WatchlistCheckError)
			}

			// Calculate pagination metrics
			totalItems = len(watchlists)
			totalPages = (totalItems + itemsPerPage - 1) / itemsPerPage

			// Update current page based on action
			h.app.Logger.Debug(op, ctx, "Updating page pointer",
				"action", action, "current_page", currentPage, "total_pages", totalPages)
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
			h.app.Logger.Debug(op, ctx, "Fetching type-specific watchlist for pagination", "type", watchlistType)
			watchlists, err := h.app.Repository.Watchlists.GetUserWatchlistsWithType(ctxDb, ctx.Sender().ID, watchlistType)
			if err != nil {
				h.app.Logger.Error(op, ctx, "Failed to fetch watchlists with type", "type", watchlistType, "error", err.Error())
				return ctx.Send(messages.WatchlistCheckError)
			}

			// Calculate pagination metrics
			totalItems = len(watchlists)
			totalPages = (totalItems + itemsPerPage - 1) / itemsPerPage

			// Update current page based on action
			h.app.Logger.Debug(op, ctx, "Updating page pointer",
				"action", action, "current_page", currentPage, "total_pages", totalPages)
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
		h.app.Logger.Debug(op, ctx, "Updating message with new page", "new_page", currentPage)
		_, err = ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
		if err != nil {
			if strings.Contains(err.Error(), "message is not modified") {
				h.app.Logger.Debug(op, ctx, "No changes detected in message")
				return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
			}
			h.app.Logger.Error(op, ctx, "Failed to edit message with updated page", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}

		h.app.Logger.Info(op, ctx, "Page updated successfully", "current_page", currentPage, "type", watchlistType)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})

	case "back_to_pagination":
		return h.handleBackToPagination(ctx, data)

	default:
		h.app.Logger.Warning(op, ctx, "Unknown callback action", "action", action)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
