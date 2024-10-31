package movie

import (
	"fmt"
	"gopkg.in/telebot.v3"
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
	moviesCache         = cache.NewCache()
	pagePointer         *int
	maxPage, movieCount int
)

func (*movieHandler) SearchMovie(context telebot.Context) error {
	log.Print("/search command received")

	if context.Message().Payload == "" {
		err := context.Send("after /search title must be provided")
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("looking for *%v*...", context.Message().Payload), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	movieData, err := search.SearchMovie(context.Message().Payload)
	if err != nil {
		log.Print(err)
		return err
	}

	if movieData.TotalResults == 0 {
		log.Print("no movies found")
		_, err = context.Bot().Edit(msg, fmt.Sprintf("no movies found for search *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	moviesCache.Clear()

	for i, result := range movieData.Results {
		moviesCache.Set(i+1, result)
	}

	movieCount = len(movieData.Results)
	maxPage = movieCount / 3
	currentPage := 1
	pagePointer = &currentPage

	paginatedMovies := helpers.PaginateMovies(moviesCache, 1, movieCount)
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err = context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
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

	movieData, err := movie.GetMovie(movieId)
	if err != nil {
		log.Printf("couldnt retrive movie from api: %v", err.Error())
		return err
	}

	newMovie := models.Movie{
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
	log.Print("returning to paginated results")

	err := context.Delete()
	if err != nil {
		log.Printf("Failed to delete movie details message: %v", err)
	}

	paginatedMovies := helpers.PaginateMovies(moviesCache, *pagePointer, movieCount)
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, *pagePointer, maxPage, movieCount)
	_, err = context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	err = context.Respond(&telebot.CallbackResponse{Text: "Returning to list"})
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *movieHandler) handleNextPage(context telebot.Context) error {
	*pagePointer++
	if *pagePointer > maxPage {
		*pagePointer = maxPage
	}

	paginatedMovies := helpers.PaginateMovies(moviesCache, *pagePointer, movieCount)
	return updateMovieMessage(context, paginatedMovies, *pagePointer, maxPage)
}

func (h *movieHandler) handlePrevPage(context telebot.Context) error {
	*pagePointer--
	if (*pagePointer) < 1 {
		*pagePointer = 1
	}

	paginatedMovies := helpers.PaginateMovies(moviesCache, *pagePointer, movieCount)
	return updateMovieMessage(context, paginatedMovies, *pagePointer, maxPage)
}

func updateMovieMessage(context telebot.Context, paginatedMovies []movie.Movie, currentPage, maxPage int) error {
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
