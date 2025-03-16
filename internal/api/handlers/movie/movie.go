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
	"log"
	"time"

	"strconv"
	"strings"
)

var (
	moviesCache = make(map[int64]*cache.Item)
	pagePointer = make(map[int64]*int)
	maxPage     = make(map[int64]int)
	movieCount  = make(map[int64]int)
)

func (h *MovieHandler) SearchMovie(ctx telebot.Context) error {
	log.Print(messages.MovieCommand)
	userId := ctx.Sender().ID

	searchQuery := ctx.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(ctx.Message().Text, "/sm") {
		searchQuery = ctx.Message().Text
	}

	if searchQuery == "" {
		return ctx.Send(messages.MovieEmptyPayload)
	}

	msg, err := ctx.Bot().Send(ctx.Chat(), fmt.Sprintf("Looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	// Fetch search results
	movieData, err := search.SearchMovie(h.app, searchQuery, userId)
	if err != nil || movieData.TotalResults == 0 {
		_, err = ctx.Bot().Edit(msg, fmt.Sprintf("No movies found for *%s*", ctx.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
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
		return err
	}

	return nil
}

func (h *MovieHandler) handleMovieDetails(ctx telebot.Context, data string) error {
	parsedId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return err
	}

	movieData, err := movie.GetMovie(h.app, parsedId, ctx.Sender().ID)
	if err != nil {
		log.Print(err)
		return err
	}

	err = movie.ShowMovie(h.app, ctx, movieData, true)
	if err != nil {
		log.Print(err)
		return err
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
}

func (h *MovieHandler) handleWatchedDetails(ctx telebot.Context, movieIdStr string) error {
	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := h.app.Repository.BeginTx(ctxDb)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return ctx.Send(messages.InternalError)
	}
	defer tx.Rollback(ctxDb)

	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	movieExists, err := h.app.Repository.Movies.MovieExists(ctxDb, int64(movieId), ctx.Sender().ID)
	if err != nil {
		log.Printf("Database error: %v", err)
		return ctx.Send(messages.InternalError)
	}

	if movieExists {
		// Movie already exists in watched list (movies table)
		log.Printf("User %d has already watched movie %d", ctx.Sender().ID, movieId)
		return ctx.Send(messages.WatchedMovie)
	}

	movieData, err := movie.GetMovie(h.app, movieId, ctx.Sender().ID)
	if err != nil {
		log.Printf("couldnt retrive movie from api: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	newMovie := database.CreateMovieParams{
		UserID:  ctx.Sender().ID,
		ApiID:   movieData.ID,
		Title:   movieData.Title,
		Runtime: movieData.Runtime,
	}

	err = h.app.Repository.Movies.CreateMovie(ctxDb, newMovie)
	if err != nil {
		log.Printf("cant create new movie: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	err = h.app.Repository.Watchlists.DeleteWatchlist(ctxDb, int64(movieId), ctx.Sender().ID)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.WatchedMovie)
	}

	if err = tx.Commit(context.Background()); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return ctx.Send(messages.InternalError)
	}

	_, err = ctx.Bot().Send(ctx.Chat(),
		fmt.Sprintf("The Movie has been marked as watched:\nDuration: *%d minutes*", movieData.Runtime),
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown},
	)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return err
	}

	return nil
}

func (h *MovieHandler) handleWatchlist(ctx telebot.Context, data string) error {
	movieId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.WatchedMovie)
	}

	movieData, err := movie.GetMovie(h.app, movieId, ctx.Sender().ID)
	if err != nil {
		log.Print(err)
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

	err = h.app.Repository.Watchlists.CreateWatchlist(ctxDb, newWatchlist)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.WatchedMovie)
	}

	_, err = ctx.Bot().Send(ctx.Chat(), "Movie added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.WatchedMovie)
	}

	return nil
}

func (h *MovieHandler) handleBackToPagination(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Delete the current movie details message
	if err := ctx.Delete(); err != nil {
		log.Printf("Failed to delete movie details message: %v", err)
		return ctx.Send(messages.InternalError)
	}

	// Paginate and send updated movie list
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
	_, err := ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to return to paginated results: %v", err)
		return ctx.Send(messages.InternalError)
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
}

func (h *MovieHandler) handleNextPage(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	return updateMovieMessage(ctx, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func (h *MovieHandler) handlePrevPage(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Update page pointer
	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	// Send updated page
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	return updateMovieMessage(ctx, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func updateMovieMessage(ctx telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage, movieCount int) error {
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Edit error: %v", err)
		if strings.Contains(err.Error(), "message is not modified") {
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
		}
		return ctx.Send(messages.InternalError)
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})
}

func (h *MovieHandler) MovieCallback(ctx telebot.Context) error {
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "movie|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

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
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
