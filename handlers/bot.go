package handlers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api/auth"
	"movie-manager-bot/api/media/movie"
	"movie-manager-bot/api/media/search"
	"movie-manager-bot/helpers"
	"movie-manager-bot/storage"
	"strconv"
	"strings"
)

var (
	moviesCache         = storage.NewCache()
	pagePointer         *int
	maxPage, movieCount int
)

func (*botHandler) Hello(context telebot.Context) error {
	log.Print("/hello command received")
	err := context.Send("Hello mathafuck")
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (*botHandler) Search(context telebot.Context) error {
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

	err = auth.Login()
	if err != nil {
		log.Print(err)
		return err
	}

	movie, err := search.SearchMovie(context.Message().Payload)
	if err != nil {
		log.Print(err)
		return err
	}

	if movie.TotalResults == 0 {
		log.Print("no movies found")
		_, err = context.Bot().Edit(msg, fmt.Sprintf("no movies found for search *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	moviesCache.Clear()

	for i, result := range movie.Results {
		moviesCache.Set(i+1, result)
	}

	movieCount = len(movie.Results)
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

func (*botHandler) OnCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 2 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return context.Respond(&telebot.CallbackResponse{Text: "Malformed data received"})
	}

	unique := dataParts[0]
	data := dataParts[1]

	log.Printf("Received callback: unique=%s, data=%s", unique, data)

	switch unique {
	case "movie":
		log.Printf("movie with id %s", data)
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

		err = context.Respond(&telebot.CallbackResponse{Text: "You found the movie!"})
		if err != nil {
			log.Print(err)
			return err
		}

	case "next":
		log.Print("next pagination result")
		*pagePointer++
		if *pagePointer > maxPage {
			*pagePointer = maxPage
		}

		paginatedMovies := helpers.PaginateMovies(moviesCache, *pagePointer, movieCount)
		err := updateMovieMessage(context, paginatedMovies, *pagePointer, maxPage)
		if err != nil {
			log.Print(err)
			return err
		}

	case "prev":
		log.Print("previous pagination result")
		*pagePointer--
		if *pagePointer < 1 {
			*pagePointer = 1
		}

		paginatedMovies := helpers.PaginateMovies(moviesCache, *pagePointer, movieCount)
		err := updateMovieMessage(context, paginatedMovies, *pagePointer, maxPage)
		if err != nil {
			log.Print(err)
			return err
		}
	default:
		err := context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
		if err != nil {
			log.Print(err)
			return err
		}
	}

	return nil
}
