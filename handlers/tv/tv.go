package tv

import (
	"errors"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/api/media/search"
	"github.com/erkinov-wtf/movie-manager-bot/api/media/tv"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/messages"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/paginators"
	"github.com/erkinov-wtf/movie-manager-bot/models"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/storage/database"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"

	"strconv"
	"strings"
)

var (
	tvCache        = make(map[int64]*cache.Cache)
	pagePointer    = make(map[int64]*int)
	maxPage        = make(map[int64]int)
	tvCount        = make(map[int64]int)
	selectedTvShow = make(map[int64]*tv.TV)
)

func (*TVHandler) SearchTV(context telebot.Context) error {
	log.Print(messages.TVShowCommand)
	userId := context.Sender().ID

	searchQuery := context.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(context.Message().Text, "/stv") {
		searchQuery = context.Message().Text
	}

	if searchQuery == "" {
		return context.Send(messages.TVShowEmptyPayload)
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	tvData, err := search.SearchTV(searchQuery, userId)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	if tvData.TotalResults == 0 {
		log.Print("no tv found")
		_, err = context.Bot().Edit(msg, fmt.Sprintf("no tv found for search *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return context.Send(messages.InternalError)
		}
		return nil
	}

	// Initialize user-specific cache and data
	tvCache[userId] = cache.NewCache()
	pagePointer[userId] = new(int)
	*pagePointer[userId] = 1
	tvCount[userId] = len(tvData.Results)
	maxPage[userId] = (tvCount[userId] + 2) / 3 // Rounded max page

	for i, result := range tvData.Results {
		tvCache[userId].Set(i+1, result)
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], 1, tvCount[userId])
	response, btn := paginators.GenerateTVResponse(paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
	_, err = context.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleTVDetails(context telebot.Context, data string) error {
	parsedId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	tvData, err := tv.GetTV(parsedId, context.Sender().ID)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	err = tv.ShowTV(context, tvData, true)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return context.Respond(&telebot.CallbackResponse{Text: messages.TVShowSelected})
}

func (h *TVHandler) handleSelectSeasons(context telebot.Context, tvId string) error {
	userId := context.Sender().ID
	TVId, _ := strconv.Atoi(tvId)

	var watchedSeasons int = 0
	if err := database.DB.Model(&models.TVShows{}).
		Select("seasons").
		Where("api_id = ? AND user_id = ?", TVId, userId).
		Scan(&watchedSeasons).Error; err != nil {
		if err.Error() != "record not found" {
			log.Printf("Database error: %v", err)
			return context.Send(messages.InternalError)
		}
	}

	tvShow, err := tv.GetTV(TVId, context.Sender().ID)
	if err != nil {
		log.Printf("Error fetching TV show: %v", err)
		return context.Send(messages.InternalError)
	}

	if watchedSeasons > 0 {
		log.Printf("userId %v already watched tv show name %s, id %v", userId, tvShow.Name, tvShow.ID)
	}

	selectedTvShow[userId] = tvShow

	btn := &telebot.ReplyMarkup{}
	var btnRows []telebot.Row

	emojis := map[int]string{
		0: "Season 0️⃣", 1: "Season 1️⃣", 2: "Season 2️⃣", 3: "Season 3️⃣", 4: "Season 4️⃣",
		5: "Season 5️⃣", 6: "Season 6️⃣", 7: "Season 7️⃣", 8: "Season 8️⃣", 9: "Season 9️⃣",
	}

	for i := 1; i <= int(selectedTvShow[userId].Seasons); i++ {
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
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleWatched(context telebot.Context, data string) error {
	// begin transaction
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

	userId := context.Sender().ID
	var watchedSeasons int = 0
	if err := tx.Model(&models.TVShows{}).
		Select("seasons").
		Where("api_id = ? AND user_id = ?", selectedTvShow[userId].ID, userId).
		Scan(&watchedSeasons).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Database error: %v", err)
			return context.Send(messages.InternalError)
		}
	}

	seasonNum, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InvalidSeason)
	}

	if seasonNum <= watchedSeasons {
		return context.Send(messages.WatchedSeason)
	}

	var episodes, runtime int64
	for i := 1; i <= seasonNum; i++ {
		tvSeason, err := tv.GetSeason(int(selectedTvShow[userId].ID), i, userId)
		if err != nil {
			log.Print(err.Error())
			return context.Send(messages.InternalError)
		}

		for _, episode := range tvSeason.Episodes {
			episodes++
			runtime += episode.Runtime
			log.Printf("TV Show: %v, Season: %v, Episode: %v, Runtime: %v", selectedTvShow[userId].ID, i, episode.EpisodeNumber, episode.Runtime)
		}
	}

	newTv := models.TVShows{
		UserID:   userId,
		ApiID:    selectedTvShow[userId].ID,
		Name:     selectedTvShow[userId].Name,
		Seasons:  int64(seasonNum),
		Episodes: episodes,
		Runtime:  runtime,
		Status:   selectedTvShow[userId].Status,
	}

	if watchedSeasons > 0 {
		if err = tx.Model(&models.TVShows{}).
			Where("api_id = ? AND user_id = ?", selectedTvShow[userId].ID, userId).
			Updates(models.TVShows{Seasons: newTv.Seasons, Episodes: newTv.Episodes, Runtime: newTv.Runtime}).Error; err != nil {
			log.Printf("cant update existing tv show data: %v", err.Error())
			return context.Send(messages.InternalError)
		}
	} else {
		if err = tx.Create(&newTv).Error; err != nil {
			log.Printf("cant create new tv show: %v", err.Error())
			return context.Send(messages.InternalError)
		}
	}

	if err = tx.Where("show_api_id = ? AND user_id = ?", selectedTvShow[userId].ID, userId).Delete(&models.Watchlist{}).Error; err != nil {
		log.Print(err)
		return context.Send("Something went wrong")
	}

	tx.Commit()

	_, err = context.Bot().Send(context.Chat(), fmt.Sprintf("The TV Show added as watched with below data:\nName: %v\nSeasons: %v\nEpisodes: %v\nRuntime: %v minutes", selectedTvShow[userId].Name, seasonNum, episodes, runtime), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleWatchlist(context telebot.Context, tvId string) error {
	tvShowId, err := strconv.Atoi(tvId)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	tvShow, err := tv.GetTV(tvShowId, context.Sender().ID)
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	newWatchlist := models.Watchlist{
		UserID:    context.Sender().ID,
		ShowApiId: tvShow.ID,
		Type:      models.TVShowType,
		Title:     tvShow.Name,
		Image:     tvShow.PosterPath,
	}

	if err = database.DB.Create(&newWatchlist).Error; err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	_, err = context.Bot().Send(context.Chat(), "Tv Show added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *TVHandler) handleBackToPagination(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	err := context.Delete()
	if err != nil {
		log.Printf("Failed to delete tv details message: %v", err)
		return context.Send(messages.InternalError)
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	response, btn := paginators.GenerateTVResponse(paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
	_, err = context.Bot().Send(context.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	err = context.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
	if err != nil {
		log.Print(err)
		return context.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleNextPage(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	return updateTVMessage(context, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func (h *TVHandler) handlePrevPage(context telebot.Context) error {
	userId := context.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return context.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	return updateTVMessage(context, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func updateTVMessage(context telebot.Context, paginatedTV []tv.TV, currentPage, maxPage, tvCount int) error {
	response, btn := paginators.GenerateTVResponse(paginatedTV, currentPage, maxPage, tvCount)
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

func (h *TVHandler) TVCallback(context telebot.Context) error {
	callback := context.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "tv|") {
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
	case "tv":
		return h.handleTVDetails(context, data)

	case "select_seasons": //callback for Watched button
		return h.handleSelectSeasons(context, data)

	case "watched":
		return h.handleWatched(context, data)

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
