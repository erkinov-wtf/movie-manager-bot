package tv

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api/media/search"
	"movie-manager-bot/api/media/tv"
	"movie-manager-bot/helpers"
	"movie-manager-bot/storage"
	"strconv"
	"strings"
)

var (
	tvCache          = storage.NewCache()
	pagePointer      *int
	maxPage, tvCount int
)

func (*tvHandler) SearchTV(context telebot.Context) error {
	log.Print("/stv command received")

	if context.Message().Payload == "" {
		err := context.Send("after /stv title must be provided")
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

	tvData, err := search.SearchTV(context.Message().Payload)
	if err != nil {
		log.Print(err)
		return err
	}

	if tvData.TotalResults == 0 {
		log.Print("no tv found")
		_, err = context.Bot().Edit(msg, fmt.Sprintf("no tv found for search *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	tvCache.Clear()

	for i, result := range tvData.Results {
		tvCache.Set(i+1, result)
	}

	tvCount = len(tvData.Results)
	maxPage = tvCount / 3
	currentPage := 1
	pagePointer = &currentPage

	paginatedTV := helpers.PaginateTV(tvCache, 1, tvCount)
	response, btn := helpers.GenerateTVResponse(paginatedTV, currentPage, maxPage, tvCount)
	_, err = context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
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
	case "tv":
		return h.handleTVDetails(context, data)

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

func (h *tvHandler) handleTVDetails(context telebot.Context, data string) error {
	parsedId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return err
	}

	tvData, err := tv.GetTV(parsedId)
	if err != nil {
		log.Print(err)
		return err
	}

	err = tv.ShowTV(context, tvData)
	if err != nil {
		log.Print(err)
		return err
	}

	return context.Respond(&telebot.CallbackResponse{Text: "You found the tv!"})
}

func (h *tvHandler) handleBackToPagination(context telebot.Context) error {
	log.Print("returning to paginated results")

	err := context.Delete()
	if err != nil {
		log.Printf("Failed to delete tv details message: %v", err)
	}

	paginatedTV := helpers.PaginateTV(tvCache, *pagePointer, tvCount)
	response, btn := helpers.GenerateTVResponse(paginatedTV, *pagePointer, maxPage, tvCount)
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

func (h *tvHandler) handleNextPage(context telebot.Context) error {
	*pagePointer++
	if *pagePointer > maxPage {
		*pagePointer = maxPage
	}

	paginatedTV := helpers.PaginateTV(tvCache, *pagePointer, tvCount)
	return updateTVMessage(context, paginatedTV, *pagePointer, maxPage)
}

func (h *tvHandler) handlePrevPage(context telebot.Context) error {
	*pagePointer--
	if (*pagePointer) < 1 {
		*pagePointer = 1
	}

	paginatedTV := helpers.PaginateTV(tvCache, *pagePointer, tvCount)
	return updateTVMessage(context, paginatedTV, *pagePointer, maxPage)
}

func updateTVMessage(context telebot.Context, paginatedTV []tv.TV, currentPage, maxPage int) error {
	response, btn := helpers.GenerateTVResponse(paginatedTV, currentPage, maxPage, tvCount)
	_, err := context.Bot().Edit(context.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Failed to update tv message: %v", err)
		return err
	}

	return nil
}
