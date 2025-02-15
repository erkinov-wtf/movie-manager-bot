package movie

import (
	"errors"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/search"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"

	"strconv"
	"strings"
)

var (
	moviesCache = make(map[int64]*cache.Cache)
	pagePointer = make(map[int64]*int)
	maxPage     = make(map[int64]int)
	movieCount  = make(map[int64]int)
)

func (*MovieHandler) SearchMovie(context telebot.Context) error {
	log.Print(messages.MovieCommand)
	userId := context.Sender().ID

	searchQuery := context.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(context.Message().Text, "/sm") {
		searchQuery = context.Message().Text
	}

	if searchQuery == "" {
		return context.Send(messages.MovieEmptyPayload)
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("Looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	// Fetch search results
	movieData, err := search.SearchMovie(searchQuery, userId)
	if err != nil || movieData.TotalResults == 0 {
		_, err = context.Bot().Edit(msg, fmt.Sprintf("No movies found for *%s*", context.Message().Payload), telebot.ModeMarkdown)
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

	_, err = context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		return err
	}

	return nil
}

func (h *MovieHandler) handleMovieDetails(context telebot.Context, data string) error {
	parsedId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return err
	}

	movieData, err := movie.GetMovie(parsedId, context.Sender().ID)
	if err != nil {
		log.Print(err)
		return err
	}

	err = movie.ShowMovie(context, movieData, true)
	if err != nil {
		log.Print(err)
		return err
	}

	return context.Respond(&telebot.CallbackResponse{Text: messages.MovieSelected})
}

func (h *MovieHandler) handleWatchedDetails(context telebot.Context, movieIdStr string) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	var existingMovie models.Movie
	result := tx.Where("api_id = ? AND user_id = ?", movieId, context.Sender().ID).First(&existingMovie)
	if result.Error == nil {
		// Movie already exists in watched list
		log.Printf("user has watched the movie: %v", existingMovie)
		return context.Send(messages.WatchedMovie)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Database error: %v", result.Error)
		return fmt.Errorf("database error: %v", result.Error)
	}

	movieData, err := movie.GetMovie(movieId, context.Sender().ID)
	if err != nil {
		log.Printf("couldnt retrive movie from api: %v", err.Error())
		return context.Send(messages.InternalError)
	}

	newMovie := models.Movie{
		UserID:  context.Sender().ID,
		ApiID:   movieData.ID,
		Title:   movieData.Title,
		Runtime: movieData.Runtime,
	}

	if err = tx.Create(&newMovie).Error; err != nil {
		log.Printf("cant create new movie: %v", err.Error())
		return context.Send(messages.WatchedMovie)
	}

	if err = tx.Where("show_api_id = ? AND user_id = ?", movieId, context.Sender().ID).Delete(&models.Watchlist{}).Error; err != nil {
		log.Print(err)
		return context.Send(messages.WatchedMovie)
	}

	tx.Commit()

	_, err = context.Bot().Send(context.Chat(), fmt.Sprintf("The Movie as watched with below data:\nDuration: *%d minutes*", movieData.Runtime), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *MovieHandler) handleWatchlist(context telebot.Context, data string) error {
	movieId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return context.Send(messages.WatchedMovie)
	}

	movieData, err := movie.GetMovie(movieId, context.Sender().ID)
	if err != nil {
		log.Print(err)
		return context.Send(messages.WatchedMovie)
	}

	newWatchlist := models.Watchlist{
		UserID:    context.Sender().ID,
		ShowApiId: movieData.ID,
		Type:      models.MovieType,
		Title:     movieData.Title,
		Image:     movieData.PosterPath,
	}

	if err = database.DB.Create(&newWatchlist).Error; err != nil {
		log.Print(err)
		return context.Send(messages.WatchedMovie)
	}

	_, err = context.Bot().Send(context.Chat(), "Movie added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.WatchedMovie)
	}

	return nil
}

func (h *MovieHandler) handleBackToPagination(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Delete the current movie details message
	if err := context.Delete(); err != nil {
		log.Printf("Failed to delete movie details message: %v", err)
		return context.Send(messages.InternalError)
	}

	// Paginate and send updated movie list
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
	_, err := context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to return to paginated results: %v", err)
		return context.Send(messages.InternalError)
	}

	return context.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
}

func (h *MovieHandler) handleNextPage(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func (h *MovieHandler) handlePrevPage(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := moviesCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Update page pointer
	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	// Send updated page
	paginatedMovies := paginators.PaginateMovies(moviesCache[userId], *pagePointer[userId], movieCount[userId])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userId], maxPage[userId], movieCount[userId])
}

func updateMovieMessage(context telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage, movieCount int) error {
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Edit error: %v", err)
		if strings.Contains(err.Error(), "message is not modified") {
			return context.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
		}
		return context.Send(messages.InternalError)
	}

	return context.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})
}

func (h *MovieHandler) MovieCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "movie|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "movie":
		return h.handleMovieDetails(context, data)

	case "watched":
		return h.handleWatchedDetails(context, data)

	case "watchlist":
		return h.handleWatchlist(context, data)

	case "back_to_pagination":
		return h.handleBackToPagination(context)

	case "next":
		return h.handleNextPage(context)

	case "prev":
		return h.handlePrevPage(context)

	default:
		return context.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
