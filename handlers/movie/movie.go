package movie

import (
	"errors"
	"fmt"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/api/media/movie"
	"movie-manager-bot/api/media/search"
	"movie-manager-bot/helpers"
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
	userID := context.Sender().ID

	if context.Message().Payload == "" {
		return context.Send("After /sm, a movie title must be provided")
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("Looking for *%v*...", context.Message().Payload), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	// Fetch search results
	movieData, err := search.SearchMovie(context.Message().Payload)
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

	paginatedMovies := helpers.PaginateMovies(moviesCache[userID], 1, movieCount[userID])
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])

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

	movieData, err := movie.GetMovie(parsedId)
	if err != nil {
		log.Print(err)
		return err
	}

	err = movie.ShowMovie(context, movieData)
	if err != nil {
		log.Print(err)
		return err
	}

	return context.Respond(&telebot.CallbackResponse{Text: "You found the movie!"})
}

func (h *movieHandler) handleWatchedDetails(context telebot.Context, movieIdStr string) error {
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		log.Print(err)
		return context.Send("Invalid movie id")
	}

	var existingMovie models.Movie
	result := database.DB.Where("api_id = ? AND user_id = ?", movieId, context.Sender().ID).First(&existingMovie)
	if result.Error == nil {
		// Movie already exists in watched list
		log.Printf("user has watched the movie: %v", existingMovie)
		context.Send("You have already watched this movie")
		return nil
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Printf("Database error: %v", result.Error)
		return fmt.Errorf("database error: %v", result.Error)
	}

	movieData, err := movie.GetMovie(movieId)
	if err != nil {
		log.Printf("couldnt retrive movie from api: %v", err.Error())
		return err
	}

	newMovie := models.Movie{
		UserID:  context.Sender().ID,
		ApiID:   movieData.ID,
		Title:   movieData.Title,
		Runtime: movieData.Runtime,
	}

	if err = database.DB.Create(&newMovie).Error; err != nil {
		log.Printf("cant create new movie: %v", err.Error())
		return err
	}

	_, err = context.Bot().Send(context.Chat(), fmt.Sprintf("The Movie as watched with below data:\nDuration: *%d minutes*", movieData.Runtime), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *movieHandler) handleBackToPagination(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: "No search results to return to"})
	}

	// Delete the current movie details message
	if err := context.Delete(); err != nil {
		log.Printf("Failed to delete movie details message: %v", err)
	}

	// Paginate and send updated movie list
	paginatedMovies := helpers.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
	_, err := context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to return to paginated results: %v", err)
		return err
	}

	return context.Respond(&telebot.CallbackResponse{Text: "Returning to search results"})
}

func (h *movieHandler) handleNextPage(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: "No search results found"})
	}

	*pagePointer[userID]++
	if *pagePointer[userID] > maxPage[userID] {
		*pagePointer[userID] = maxPage[userID]
	}

	paginatedMovies := helpers.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
}

func (h *movieHandler) handlePrevPage(context telebot.Context) error {
	userID := context.Sender().ID

	if _, ok := moviesCache[userID]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: "No search results found"})
	}

	// Update page pointer
	*pagePointer[userID]--
	if *pagePointer[userID] < 1 {
		*pagePointer[userID] = 1
	}

	// Send updated page
	paginatedMovies := helpers.PaginateMovies(moviesCache[userID], *pagePointer[userID], movieCount[userID])
	return updateMovieMessage(context, paginatedMovies, *pagePointer[userID], maxPage[userID], movieCount[userID])
}

func updateMovieMessage(context telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage, movieCount int) error {
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to update movie message: %v", err)
		return err
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
		return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "movie":
		return h.handleMovieDetails(context, data)
	case "watched":
		return h.handleWatchedDetails(context, data)
	case "back_to_pagination":
		return h.handleBackToPagination(context)
	case "next":
		return h.handleNextPage(context)
	case "prev":
		return h.handlePrevPage(context)
	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}
