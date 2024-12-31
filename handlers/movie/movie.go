package movie

import (
	"errors"
	"fmt"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/api/media/movie"
	"movie-manager-bot/api/media/search"
	"movie-manager-bot/helpers/messages"
	"movie-manager-bot/helpers/paginators"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/cache"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

var (
	moviesCache = make(map[int64]*cache.Cache)
	pagePointer = make(map[int64]*int)
	maxPage     = make(map[int64]int)
	movieCount  = make(map[int64]int)
)

func (*movieHandler) SearchMovie(context telebot.Context) error {
	log.Print(messages.MovieCommand)
	userID := context.Sender().ID

	if context.Message().Payload == "" {
		return context.Send(messages.MovieEmptyPayload)
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("Looking for *%v*...", context.Message().Payload), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	// Fetch search results
	movieData, err := search.SearchMovie(context.Message().Payload, userID)
	if err != nil || movieData.TotalResults == 0 {
		_, err = context.Bot().Edit(msg, fmt.Sprintf("No movies found for *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			return err
		}
		return nil
	}

	// Initialize user-specific cache and data
	moviesCache[userID] = cache.NewCache()
	pagePointer[userID] = new(int)
	*pagePointer[userID] = 1
	movieCount[userID] = len(movieData.Results)
	maxPage[userID] = (movieCount[userID] + 2) / 3 // Rounded max page

	for i, result := range movieData.Results {
		moviesCache[userID].Set(i+1, result)
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userID], 1, movieCount[userID])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])

	_, err = context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		return err
	}

	return nil
}

func (h *movieHandler) handleMovieDetails(context telebot.Context, data string) error {
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

func (h *movieHandler) handleWatchedDetails(context telebot.Context, movieIdStr string) error {
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

func (h *movieHandler) handleWatchlist(context telebot.Context, data string) error {
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

func (h *movieHandler) handleBackToPagination(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Delete the current movie details message
	if err := context.Delete(); err != nil {
		log.Printf("Failed to delete movie details message: %v", err)
		return context.Send(messages.InternalError)
	}

	// Paginate and send updated movie list
	paginatedMovies := paginators.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
	_, err := context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to return to paginated results: %v", err)
		return context.Send(messages.InternalError)
	}

	return context.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
}

func (h *movieHandler) handleNextPage(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userID]++
	if *pagePointer[userID] > maxPage[userID] {
		*pagePointer[userID] = maxPage[userID]
	}

	paginatedMovies := paginators.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
}

func (h *movieHandler) handlePrevPage(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	// Update page pointer
	*pagePointer[userID]--
	if *pagePointer[userID] < 1 {
		*pagePointer[userID] = 1
	}

	// Send updated page
	paginatedMovies := paginators.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
}

func updateMovieMessage(context telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage, movieCount int) error {
	response, btn := paginators.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to update movie message: %v", err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *movieHandler) MovieCallback(context telebot.Context) error {
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
