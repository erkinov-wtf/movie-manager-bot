package movie

import (
	"gopkg.in/telebot.v3"
	"log"
	movieType "movie-manager-bot/api/media/movie"
	"movie-manager-bot/helpers"
	"movie-manager-bot/interfaces"
)

type botHandler struct{}

func NewBotHandler() interfaces.BotInterface {
	return &botHandler{}
}

func updateMovieMessage(context telebot.Context, paginatedMovies []movieType.Movie, currentPage, maxPage int) error {
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to update movie message: %v", err)
		return err
	}

	return nil
}
