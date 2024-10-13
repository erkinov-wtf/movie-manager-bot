package tv

import (
	"gopkg.in/telebot.v3"
	"log"
	movieType "movie-manager-bot/api/media/movie"
	"movie-manager-bot/helpers"
	"movie-manager-bot/storage"
	"strings"
)

var (
	tvCache             = storage.NewCache()
	pagePointer         *int
	maxPage, movieCount int
)

func (*tvHandler) SearchTV(context telebot.Context) error {
	err := context.Send("fuck")
	if err != nil {
		return err
	}
	return nil
}

func (h *tvHandler) TVCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "tv|") {
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
	case "show":
		return h.handleTVShowDetails(context, data)
	case "next":
		return h.handleNextPage(context)
	case "prev":
		return h.handlePrevPage(context)
	default:
		return context.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
}

func (h *tvHandler) handleTVShowDetails(context telebot.Context, data string) error {
	// Implement TV show details logic here
	return nil
}

func (h *tvHandler) handleNextPage(context telebot.Context) error {
	// Implement next page logic for TV shows here
	return nil
}

func (h *tvHandler) handlePrevPage(context telebot.Context) error {
	// Implement previous page logic for TV shows here
	return nil
}

func updateTVMessage(context telebot.Context, paginatedMovies []movieType.Movie, currentPage, maxPage int) error {
	response, btn := helpers.GenerateMovieResponse(paginatedMovies, currentPage, maxPage, movieCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to update movie message: %v", err)
		return err
	}

	return nil
}
