package tv

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/search"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/constants"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/paginators"
	"gopkg.in/telebot.v3"
	"log"
	"time"

	"strconv"
	"strings"
)

var (
	tvCache        = make(map[int64]*cache.Item)
	pagePointer    = make(map[int64]*int)
	maxPage        = make(map[int64]int)
	tvCount        = make(map[int64]int)
	selectedTvShow = make(map[int64]*tv.TV)
)

func (h *TVHandler) SearchTV(ctx telebot.Context) error {
	log.Print(messages.TVShowCommand)
	userId := ctx.Sender().ID

	searchQuery := ctx.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(ctx.Message().Text, "/stv") {
		searchQuery = ctx.Message().Text
	}

	if searchQuery == "" {
		return ctx.Send(messages.TVShowEmptyPayload)
	}

	msg, err := ctx.Bot().Send(ctx.Chat(), fmt.Sprintf("Looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	tvData, err := search.SearchTV(h.app, searchQuery, userId)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	if tvData.TotalResults == 0 {
		log.Print("no tv found")
		_, err = ctx.Bot().Edit(msg, fmt.Sprintf("no tv found for search *%s*", ctx.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}
		return nil
	}

	// Initialize user-specific cache and data
	if oldCache, exists := tvCache[userId]; exists {
		oldCache.Clear()
	}
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
	_, err = ctx.Bot().Edit(msg, response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleTVDetails(ctx telebot.Context, data string) error {
	parsedId, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	tvData, err := tv.GetTV(h.app, parsedId, ctx.Sender().ID)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	err = tv.ShowTV(h.app, ctx, tvData, true)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.TVShowSelected})
}

func getSeasonEmoji(seasonNumber int32) string {
	numberEmojis := []string{"0️⃣", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣"}

	if seasonNumber >= 0 && seasonNumber <= 9 {
		return fmt.Sprintf("Season %s", numberEmojis[seasonNumber])
	}

	return fmt.Sprintf("Season %d", seasonNumber)
}

func (h *TVHandler) handleSelectSeasons(ctx telebot.Context, tvId string) error {
	userId := ctx.Sender().ID
	TVId, _ := strconv.Atoi(tvId)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	watchedSeasons, err := h.app.Repository.TVShows.GetWatchedSeasons(ctxDb, int64(TVId), userId)
	if err != nil {
		log.Printf("Error fetching TV show: %v", err)
		return ctx.Send(messages.InternalError)
	}

	tvShow, err := tv.GetTV(h.app, TVId, ctx.Sender().ID)
	if err != nil {
		log.Printf("Error fetching TV show: %v", err)
		return ctx.Send(messages.InternalError)
	}

	if watchedSeasons > 0 {
		log.Printf("userId %v already watched tv show name %s, id %v", userId, tvShow.Name, tvShow.Id)
	}

	selectedTvShow[userId] = tvShow

	btn := &telebot.ReplyMarkup{}
	var btnRows []telebot.Row

	for i := int32(1); i <= selectedTvShow[userId].Seasons; i++ {
		emoji := getSeasonEmoji(i)

		if i <= int32(watchedSeasons) {
			emoji = fmt.Sprintf("✅ %s", emoji)
		}

		btnRows = append(btnRows, btn.Row(btn.Data(emoji, "", fmt.Sprintf("tv|watched|%v", i))))
	}

	btn.Inline(btnRows...)

	_, err = ctx.Bot().Send(ctx.Chat(), "How many seasons have you watched?", btn)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleWatched(ctx telebot.Context, data string) error {
	// begin transaction
	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := h.app.Repository.BeginTx(ctxDb)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return ctx.Send(messages.InternalError)
	}
	defer tx.Rollback(ctxDb)

	userId := ctx.Sender().ID
	watchedSeasons, err := h.app.Repository.TVShows.GetWatchedSeasons(ctxDb, selectedTvShow[userId].Id, userId)
	if err != nil {
		log.Printf("Error fetching TV show: %v", err)
		return ctx.Send(messages.InternalError)
	}

	seasonNum, err := strconv.Atoi(data)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InvalidSeason)
	}

	if int32(seasonNum) <= watchedSeasons {
		return ctx.Send(messages.WatchedSeason)
	}

	var episodes, runtime int32
	for i := 1; i <= seasonNum; i++ {
		tvSeason, err := tv.GetSeason(h.app, int(selectedTvShow[userId].Id), i, userId)
		if err != nil {
			log.Print(err.Error())
			return ctx.Send(messages.InternalError)
		}

		for _, episode := range tvSeason.Episodes {
			episodes++
			runtime += episode.Runtime
			log.Printf("TV Show: %v, Season: %v, Episode: %v, Runtime: %v", selectedTvShow[userId].Id, i, episode.EpisodeNumber, episode.Runtime)
		}
	}

	if watchedSeasons > 0 {
		updateTv := database.UpdateTVShowParams{
			ApiID:    selectedTvShow[userId].Id,
			UserID:   userId,
			Seasons:  int32(seasonNum),
			Episodes: episodes,
			Runtime:  runtime,
		}
		err = tx.Repos.TVShows.UpdateTVShow(ctxDb, updateTv)
		if err != nil {
			log.Printf("cant update existing tv show data: %v", err.Error())
			return ctx.Send(messages.InternalError)
		}
	} else {
		newTv := database.CreateTVShowParams{
			UserID:   userId,
			ApiID:    selectedTvShow[userId].Id,
			Name:     selectedTvShow[userId].Name,
			Seasons:  int32(seasonNum),
			Episodes: episodes,
			Runtime:  runtime,
			Status:   selectedTvShow[userId].Status,
		}
		err = tx.Repos.TVShows.CreateTVShow(ctxDb, newTv)
		if err != nil {
			log.Printf("cant create new tv show: %v", err.Error())
			return ctx.Send(messages.InternalError)
		}
	}

	err = tx.Repos.Watchlists.DeleteWatchlist(ctxDb, selectedTvShow[userId].Id, userId)
	if err != nil {
		log.Print(err)
		return ctx.Send("Something went wrong")
	}

	if err = tx.Commit(context.Background()); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return ctx.Send(messages.InternalError)
	}

	_, err = ctx.Bot().Send(ctx.Chat(), fmt.Sprintf("The TV Show added as watched with below data:\nName: %v\nSeasons: %v\nEpisodes: %v\nRuntime: %v minutes", selectedTvShow[userId].Name, seasonNum, episodes, runtime), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleWatchlist(ctx telebot.Context, tvId string) error {
	tvShowId, err := strconv.Atoi(tvId)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	tvShow, err := tv.GetTV(h.app, tvShowId, ctx.Sender().ID)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	newWatchlist := database.CreateWatchlistParams{
		UserID:    ctx.Sender().ID,
		ShowApiID: tvShow.Id,
		Type:      constants.TVShowType,
		Title:     tvShow.Name,
		Image:     &tvShow.PosterPath,
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = h.app.Repository.Watchlists.CreateWatchlist(ctxDb, newWatchlist); err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	_, err = ctx.Bot().Send(ctx.Chat(), "Tv Show added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (h *TVHandler) handleBackToPagination(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	err := ctx.Delete()
	if err != nil {
		log.Printf("Failed to delete tv details message: %v", err)
		return ctx.Send(messages.InternalError)
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	response, btn := paginators.GenerateTVResponse(paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
	_, err = ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	err = ctx.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	return nil
}

func (h *TVHandler) handleNextPage(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	return updateTVMessage(ctx, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func (h *TVHandler) handlePrevPage(ctx telebot.Context) error {
	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	return updateTVMessage(ctx, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func updateTVMessage(ctx telebot.Context, paginatedTV []tv.TV, currentPage, maxPage, tvCount int) error {
	response, btn := paginators.GenerateTVResponse(paginatedTV, currentPage, maxPage, tvCount)
	_, err := ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("Edit error: %v", err)
		if strings.Contains(err.Error(), "message is not modified") {
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
		}
		return ctx.Send(messages.InternalError)
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})
}

func (h *TVHandler) TVCallback(ctx telebot.Context) error {
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "tv|") {
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]

	switch action {
	case "tv":
		return h.handleTVDetails(ctx, data)

	case "select_seasons": //callback for Watched button
		return h.handleSelectSeasons(ctx, data)

	case "watched":
		return h.handleWatched(ctx, data)

	case "watchlist":
		return h.handleWatchlist(ctx, data)

	case "back_to_pagination":
		return h.handleBackToPagination(ctx)

	case "next":
		return h.handleNextPage(ctx)

	case "prev":
		return h.handlePrevPage(ctx)

	default:
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
