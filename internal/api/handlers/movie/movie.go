package movie

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/search"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"strconv"
	"strings"
	"time"
)

var (
	moviesCache = make(map[int64]*cache.Item)
	pagePointer = make(map[int64]*int)
	maxPage     = make(map[int64]int)
	movieCount  = make(map[int64]int)
)

func (h *MovieHandler) SearchMovie(ctx telebot.Context) error {
	const op = "movie.SearchMovie"
	h.app.Logger.Info(op, ctx, "Movie search command received")

	userId := ctx.Sender().ID

	searchQuery := ctx.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(ctx.Message().Text, "/sm") {
		searchQuery = ctx.Message().Text
	}

	if searchQuery == "" {
		h.app.Logger.Warning(op, ctx, "Empty search query provided")
		return ctx.Send(messages.MovieEmptyPayload)
	}

	h.app.Logger.Debug(op, ctx, "Sending loading message", "search_query", searchQuery)
	msg, err := ctx.Bot().Send(ctx.Chat(), fmt.Sprintf("Looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send loading message", "error", err.Error())
		return err
	}

	// Fetch search results
	movieData, err := search.SearchMovie(h.app, searchQuery, userId)
	if err != nil || movieData.TotalResults == 0 {
		h.app.Logger.Info(op, ctx, "No movies found for query", "query", searchQuery)
		_, err = ctx.Bot().Edit(msg, fmt.Sprintf("No movies found for *%s*", searchQuery), telebot.ModeMarkdown)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to edit message with no results", "error", err.Error())
			return err
		}
		return nil
	}

	// Initialize user-specific cache and data
	if oldCache, exists := moviesCache[userId]; exists {
		oldCache.Clear()
	}
	moviesCache[userId] = cache.NewCache()

	pagePointer[userId] = new(int)
	*pagePointer[userId] = 1
	movieCount[userId] = len(movieData.Results)
	maxPage[userId] = (movieCount[userId] + 2) / 3 // Rounded max page

	for i, result := range movieData.Results {
		moviesCache[userId].Set(i+1, result)
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], 1, movieCount[userId])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])

	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to edit message with search results", "error", err.Error())
		return err
	}

	h.app.Logger.Info(op, ctx, "Movie search results displayed successfully",
		"result_count", movieCount[userId])
	return nil
}

func (h *MovieHandler) handleMovieDetails(ctx telebot.Context, data string) error {
	const op = "movie.handleMovieDetails"
	h.app.Logger.Info(op, ctx, "Fetching movie details", "movie_id", data)

	parsedId, err := strconv.Atoi(data)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse movie ID", "movie_id", data, "error", err.Error())
		return err
	}

	movieData, err := movie.GetMovie(h.app, parsedId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to get movie data from TMDB", "movie_id", parsedId, "error", err.Error())
		return err
	}

	err = movie.ShowMovie(h.app, ctx, movieData, true)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to show movie details", "movie_id", parsedId, "error", err.Error())
		return err
	}

	h.app.Logger.Info(op, ctx, "Movie details displayed successfully", "movie_id", parsedId, "title", movieData.Title)
	return ctx.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
}

func (h *MovieHandler) handleWatchedDetails(ctx telebot.Context, movieIdStr string) error {
	const op = "movie.handleWatchedDetails"
	h.app.Logger.Info(op, ctx, "Marking movie as watched", "movie_id", movieIdStr)

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Starting database transaction")
	tx, err := h.app.Repository.BeginTx(ctxDb)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to begin transaction", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}
	defer tx.Rollback(ctxDb)

	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse movie ID", "movie_id", movieIdStr, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Checking if movie exists in watched list", "movie_id", movieId)
	movieExists, err := h.app.Repository.Movies.MovieExists(ctxDb, int64(movieId), ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Database error when checking movie existence", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if movieExists {
		// Movie already exists in watched list (movies table)
		h.app.Logger.Info(op, ctx, "Movie already marked as watched", "movie_id", movieId)
		return ctx.Send(messages.WatchedMovie)
	}

	h.app.Logger.Debug(op, ctx, "Retrieving movie data from TMDB", "movie_id", movieId)
	movieData, err := movie.GetMovie(h.app, movieId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve movie from API", "movie_id", movieId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	newMovie := database.CreateMovieParams{
		UserID:  ctx.Sender().ID,
		ApiID:   movieData.ID,
		Title:   movieData.Title,
		Runtime: movieData.Runtime,
	}

	h.app.Logger.Debug(op, ctx, "Adding movie to watched list", "movie_title", movieData.Title)
	err = h.app.Repository.Movies.CreateMovie(ctxDb, newMovie)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to create new movie record", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Removing movie from watchlist", "movie_id", movieId)
	err = h.app.Repository.Watchlists.DeleteWatchlist(ctxDb, int64(movieId), ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Warning(op, ctx, "Failed to delete movie from watchlist, may not exist", "error", err.Error())
		// Continue execution as this is not critical
	}

	h.app.Logger.Debug(op, ctx, "Committing transaction")
	if err = tx.Commit(context.Background()); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to commit transaction", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	_, err = ctx.Bot().Send(ctx.Chat(),
		fmt.Sprintf("The Movie has been marked as watched:\nDuration: *%d minutes*", movieData.Runtime),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send confirmation message", "error", err.Error())
		return err
	}

	h.app.Logger.Info(op, ctx, "Movie successfully marked as watched",
		"movie_id", movieId, "title", movieData.Title, "runtime", movieData.Runtime)
	return nil
}

func (h *MovieHandler) handleWatchlist(ctx telebot.Context, data string) error {
	const op = "movie.handleWatchlist"
	h.app.Logger.Info(op, ctx, "Adding movie to watchlist", "movie_id", data)

	movieId, err := strconv.Atoi(data)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse movie ID", "movie_id", data, "error", err.Error())
		return ctx.Send(messages.WatchedMovie)
	}

	movieData, err := movie.GetMovie(h.app, movieId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to retrieve movie data", "movie_id", movieId, "error", err.Error())
		return ctx.Send(messages.WatchedMovie)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	newWatchlist := database.CreateWatchlistParams{
		UserID:    ctx.Sender().ID,
		ShowApiID: movieData.ID,
		Type:      constants.MovieType,
		Title:     movieData.Title,
		Image:     &movieData.PosterPath,
	}

	h.app.Logger.Debug(op, ctx, "Adding movie to watchlist in database", "movie_title", movieData.Title)
	err = h.app.Repository.Watchlists.CreateWatchlist(ctxDb, newWatchlist)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to add movie to watchlist", "error", err.Error())
		return ctx.Send(messages.WatchedMovie)
	}

	_, err = ctx.Bot().Send(ctx.Chat(), "Movie added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send confirmation message", "error", err.Error())
		return ctx.Send(messages.WatchedMovie)
	}

	h.app.Logger.Info(op, ctx, "Movie successfully added to watchlist",
		"movie_id", movieId, "title", movieData.Title)
	return nil
}

func (h *MovieHandler) handleBackToPagination(ctx telebot.Context) error {
	const op = "movie.handleBackToPagination"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Returning to paginated search results")

	if _, ok := moviesCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Delete the current movie details message
	if err := ctx.Delete(); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to delete movie details message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	// Paginate and send updated movie list
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
	_, err := ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send paginated results", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Successfully returned to search results")
	return ctx.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
}

func (h *MovieHandler) handleNextPage(ctx telebot.Context) error {
	const op = "movie.handleNextPage"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Moving to next page of search results")

	if _, ok := moviesCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	h.app.Logger.Debug(op, ctx, "Updating message with next page", "new_page", *pagePointer[userId])
	return updateMovieMessage(h, ctx, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func (h *MovieHandler) handlePrevPage(ctx telebot.Context) error {
	const op = "movie.handlePrevPage"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Moving to previous page of search results")

	if _, ok := moviesCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Update page pointer
	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	// Send updated page
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	h.app.Logger.Debug(op, ctx, "Updating message with previous page", "new_page", *pagePointer[userId])
	return updateMovieMessage(h, ctx, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func updateMovieMessage(h *MovieHandler, ctx telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage, movieCount int) error {
	const op = "movie.updateMovieMessage"

	response, btn := paginators.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		if strings.Contains(err.Error(), "message is not modified") {
			h.app.Logger.Debug(op, ctx, "No changes detected in message")
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
		}
		h.app.Logger.Error(op, ctx, "Failed to edit message with updated page", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Page updated successfully", "current_page", currentPage)
	return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})
}

func (h *MovieHandler) MovieCallback(ctx telebot.Context) error {
	const op = "movie.MovieCallback"
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	h.app.Logger.Info(op, ctx, "Processing movie callback", "callback_data", trimmed)

	if !strings.HasPrefix(trimmed, "movie|") {
		h.app.Logger.Warning(op, ctx, "Invalid callback prefix", "callback_data", trimmed)
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "callback_data", callback.Data,
			"parts_count", len(dataParts))
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]
	h.app.Logger.Debug(op, ctx, "Processing callback action", "action", action, "data", data)

	switch action {
	case "movie":
		return h.handleMovieDetails(ctx, data)

	case "watched":
		return h.handleWatchedDetails(ctx, data)

	case "watchlist":
		return h.handleWatchlist(ctx, data)

	case "back_to_pagination":
		return h.handleBackToPagination(ctx)

	case "next":
		return h.handleNextPage(ctx)

	case "prev":
		return h.handlePrevPage(ctx)

	default:
		h.app.Logger.Warning(op, ctx, "Unknown callback action", "action", action)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
