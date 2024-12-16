package tv

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api/media/search"
	"movie-manager-bot/api/media/tv"
	"movie-manager-bot/helpers"
	"movie-manager-bot/models"
	"movie-manager-bot/storage/cache"
	"movie-manager-bot/storage/database"
	"strconv"
	"strings"
)

var (
	tvCache          = cache.NewCache()
	pagePointer      *int
	maxPage, tvCount int
	selectedTvShow   *tv.TV
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

func (h *tvHandler) handleSelectSeasons(context telebot.Context, tvId string) error {
	TVId, _ := strconv.Atoi(tvId)

	var watchedSeasons int = 0
	if err := database.DB.Model(&models.TVShows{}).
		Select("seasons").
		Where("api_id = ? AND user_id = ?", TVId, context.Sender().ID).
		Scan(&watchedSeasons).Error; err != nil {
		if err.Error() != "record not found" {
			log.Printf("Database error: %v", err)
			return fmt.Errorf("database error: %v", err)
		}
	}

	tvShow, err := tv.GetTV(TVId)
	if err != nil {
		log.Printf("Error fetching TV show: %v", err)
		return fmt.Errorf("error fetching TV show: %v", err)
	}

	if watchedSeasons > 0 {
		log.Printf("userId %v already watched tv show name %s, id %v", context.Sender().ID, tvShow.Name, tvShow.ID)
	}

	selectedTvShow = tvShow

	btn := &telebot.ReplyMarkup{}
	var btnRows []telebot.Row

	emojis := map[int]string{
		0: "Season 0️⃣", 1: "Season 1️⃣", 2: "Season 2️⃣", 3: "Season 3️⃣", 4: "Season 4️⃣",
		5: "Season 5️⃣", 6: "Season 6️⃣", 7: "Season 7️⃣", 8: "Season 8️⃣", 9: "Season 9️⃣",
	}

	for i := 1; i <= int(selectedTvShow.Seasons); i++ {
		emoji := emojis[i]
		if emoji == "" {
			emoji = fmt.Sprintf("Season %d", i)
		}
		if i <= watchedSeasons {
			emoji = fmt.Sprintf("✅ %v", emoji)
		}
		btnRows = append(btnRows, btn.Row(btn.Data(fmt.Sprintf("%s", emoji), "", fmt.Sprintf("tv|watched|%v", i))))
	}

	btn.Inline(btnRows...)

	_, err = context.Bot().Send(context.Chat(), "How many seasons have you watched?", btn)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *tvHandler) handleWatched(context telebot.Context, data string) error {
	var watchedSeasons int = 0
	if err := database.DB.Model(&models.TVShows{}).
		Select("seasons").
		Where("api_id = ? AND user_id = ?", selectedTvShow.ID, context.Sender().ID).
		Scan(&watchedSeasons).Error; err != nil {
		if err.Error() != "record not found" {
			log.Printf("Database error: %v", err)
			return fmt.Errorf("database error: %v", err)
		}
	}

	seasonNum, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return context.Send("Invalid season number.")
	}

	if seasonNum <= watchedSeasons {
		context.Send("You already watched this season, please select later seasons")
		return nil
	}

	var episodes, runtime int64
	for i := 1; i <= seasonNum; i++ {
		tvSeason, err := tv.GetSeason(int(selectedTvShow.ID), i)
		if err != nil {
			log.Print(err.Error())
			return err
		}

		for _, episode := range tvSeason.Episodes {
			episodes++
			runtime += episode.Runtime
			log.Printf("TV Show: %v, Season: %v, Episode: %v, Runtime: %v", selectedTvShow.ID, i, episode.EpisodeNumber, episode.Runtime)
		}
	}

	newTv := models.TVShows{
		UserID:   context.Sender().ID,
		ApiID:    selectedTvShow.ID,
		Name:     selectedTvShow.Name,
		Seasons:  int64(seasonNum),
		Episodes: episodes,
		Runtime:  runtime,
	}

	if watchedSeasons > 0 {
		if err = database.DB.Model(&models.TVShows{}).
			Where("api_id = ? AND user_id = ?", selectedTvShow.ID, context.Sender().ID).
			Updates(models.TVShows{Seasons: newTv.Seasons, Episodes: newTv.Episodes, Runtime: newTv.Runtime}).Error; err != nil {
			log.Printf("cant update existing tv show data: %v", err.Error())
			return err
		}
	} else {
		if err = database.DB.Create(&newTv).Error; err != nil {
			log.Printf("cant create new tv show: %v", err.Error())
			return err
		}
	}

	_, err = context.Bot().Send(context.Chat(), fmt.Sprintf("The TV Show added as watched with below data:\nSeasons: %v\nEpisodes: %v\nRuntime: %v", seasonNum, episodes, runtime), telebot.ModeMarkdown)
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

	case "select_seasons": //callback for Watched button
		return h.handleSelectSeasons(context, data)

	case "watched":
		return h.handleWatched(context, data)

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
