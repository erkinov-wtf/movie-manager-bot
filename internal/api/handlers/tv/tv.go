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
	"strconv"
	"strings"
	"time"
)

var (
	tvCache        = make(map[int64]*cache.Item)
	pagePointer    = make(map[int64]*int)
	maxPage        = make(map[int64]int)
	tvCount        = make(map[int64]int)
	selectedTvShow = make(map[int64]*tv.TV)
)

func (h *TVHandler) SearchTV(ctx telebot.Context) error {
	const op = "tv.SearchTV"
	h.app.Logger.Info(op, ctx, "TV show search command received")
	userId := ctx.Sender().ID

	searchQuery := ctx.Message().Payload
	if searchQuery == "" && !strings.HasPrefix(ctx.Message().Text, "/stv") {
		searchQuery = ctx.Message().Text
	}

	if searchQuery == "" {
		h.app.Logger.Warning(op, ctx, "Empty search query provided")
		return ctx.Send(messages.TVShowEmptyPayload)
	}

	h.app.Logger.Debug(op, ctx, "Sending loading message", "search_query", searchQuery)
	msg, err := ctx.Bot().Send(ctx.Chat(), fmt.Sprintf("Looking for *%v*...", searchQuery), telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send loading message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	tvData, err := search.SearchTV(h.app, searchQuery, userId)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to search TV shows", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if tvData.TotalResults == 0 {
		h.app.Logger.Info(op, ctx, "No TV shows found for query", "query", searchQuery)
		_, err = ctx.Bot().Edit(msg, fmt.Sprintf("no tv found for search *%s*", ctx.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to edit message with no results", "error", err.Error())
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
		h.app.Logger.Error(op, ctx, "Failed to edit message with search results", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "TV show search results displayed successfully",
		"result_count", tvCount[userId])
	return nil
}

func (h *TVHandler) handleTVDetails(ctx telebot.Context, data string) error {
	const op = "tv.handleTVDetails"
	h.app.Logger.Info(op, ctx, "Fetching TV show details", "tv_id", data)

	parsedId, err := strconv.Atoi(data)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse TV show ID", "tv_id", data, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	tvData, err := tv.GetTV(h.app, parsedId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to get TV show data from TMDB", "tv_id", parsedId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	err = tv.ShowTV(h.app, ctx, tvData, true)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to show TV show details", "tv_id", parsedId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "TV show details displayed successfully", "tv_id", parsedId, "name", tvData.Name)
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
	const op = "tv.handleSelectSeasons"
	h.app.Logger.Info(op, ctx, "Handling season selection for TV show", "tv_id", tvId)

	userId := ctx.Sender().ID
	TVId, _ := strconv.Atoi(tvId)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Fetching watched seasons from database", "tv_id", TVId)
	watchedSeasons, err := h.app.Repository.TVShows.GetWatchedSeasons(ctxDb, int64(TVId), userId)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Error fetching watched seasons", "tv_id", TVId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Retrieving TV show data from TMDB", "tv_id", TVId)
	tvShow, err := tv.GetTV(h.app, TVId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Error fetching TV show from TMDB", "tv_id", TVId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if watchedSeasons > 0 {
		h.app.Logger.Info(op, ctx, "User has already watched some seasons of this show",
			"tv_id", TVId, "name", tvShow.Name, "watched_seasons", watchedSeasons)
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
		h.app.Logger.Error(op, ctx, "Failed to send season selection keyboard", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Season selection keyboard displayed successfully", "tv_name", tvShow.Name)
	return nil
}

func (h *TVHandler) handleWatched(ctx telebot.Context, data string) error {
	const op = "tv.handleWatched"
	h.app.Logger.Info(op, ctx, "Processing TV show watch status update", "season_number", data)

	// Use a more appropriate timeout for the transaction
	ctxDb, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Starting database transaction")
	tx, err := h.app.Repository.BeginTx(ctxDb)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to begin transaction", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}
	defer tx.Rollback(ctxDb)

	userId := ctx.Sender().ID
	tvShow, ok := selectedTvShow[userId]
	if !ok {
		h.app.Logger.Error(op, ctx, "No selected TV show found for user")
		return ctx.Send(messages.InternalError)
	}

	seasonNum, err := strconv.Atoi(data)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse season number", "data", data, "error", err.Error())
		return ctx.Send(messages.InvalidSeason)
	}

	h.app.Logger.Debug(op, ctx, "Checking if user has already watched any seasons",
		"tv_id", tvShow.Id, "name", tvShow.Name)
	watchedSeasons, err := h.app.Repository.TVShows.GetWatchedSeasons(ctxDb, tvShow.Id, userId)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Error fetching watched seasons", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if int32(seasonNum) <= watchedSeasons {
		h.app.Logger.Info(op, ctx, "User has already watched this season",
			"season", seasonNum, "watched_up_to", watchedSeasons)
		return ctx.Send(messages.WatchedSeason)
	}

	// Get existing show data if any seasons were previously watched
	var existingEpisodes, existingRuntime int32
	if watchedSeasons > 0 {
		h.app.Logger.Debug(op, ctx, "Fetching existing TV show data", "tv_id", tvShow.Id)
		existingShow, err := h.app.Repository.TVShows.GetUserTVShow(ctxDb, tvShow.Id, userId)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Error fetching existing TV show data", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
		existingEpisodes = existingShow.Episodes
		existingRuntime = existingShow.Runtime
	}

	// Process seasons concurrently
	type seasonResult struct {
		Episodes int32
		Runtime  int32
		Error    error
	}

	numSeasonsToProcess := seasonNum - int(watchedSeasons)
	resultChan := make(chan seasonResult, numSeasonsToProcess)

	// Create a context with a reasonable timeout for all API calls
	fetchCtx, fetchCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer fetchCancel()

	h.app.Logger.Debug(op, ctx, "Starting concurrent fetch of season data",
		"seasons_to_process", numSeasonsToProcess)
	// Start goroutines for each season
	for i := int(watchedSeasons) + 1; i <= seasonNum; i++ {
		go func(seasonIndex int) {
			var result seasonResult

			tvSeason, err := tv.GetSeason(h.app, int(tvShow.Id), seasonIndex, userId)
			if err != nil {
				result.Error = fmt.Errorf("error fetching season %d: %v", seasonIndex, err)
				resultChan <- result
				return
			}

			for _, episode := range tvSeason.Episodes {
				result.Episodes++
				result.Runtime += episode.Runtime
			}

			resultChan <- result
		}(i)
	}

	// Collect results
	h.app.Logger.Debug(op, ctx, "Collecting season data results")
	var newEpisodes, newRuntime int32
	for i := 0; i < numSeasonsToProcess; i++ {
		select {
		case result := <-resultChan:
			if result.Error != nil {
				h.app.Logger.Error(op, ctx, "Error while fetching season data", "error", result.Error.Error())
				return ctx.Send(messages.InternalError)
			}
			newEpisodes += result.Episodes
			newRuntime += result.Runtime
		case <-fetchCtx.Done():
			h.app.Logger.Error(op, ctx, "Timed out while fetching seasons")
			return ctx.Send(messages.InternalError)
		}
	}

	// Total episodes and runtime (existing + new)
	totalEpisodes := existingEpisodes + newEpisodes
	totalRuntime := existingRuntime + newRuntime

	h.app.Logger.Debug(op, ctx, "Updating database with TV show data",
		"seasons", seasonNum, "episodes", totalEpisodes, "runtime", totalRuntime)
	// Use a single database operation - update or create
	var dbErr error
	if watchedSeasons > 0 {
		dbErr = tx.Repos.TVShows.UpdateTVShow(ctxDb, database.UpdateTVShowParams{
			ApiID:    tvShow.Id,
			UserID:   userId,
			Seasons:  int32(seasonNum),
			Episodes: totalEpisodes,
			Runtime:  totalRuntime,
		})
	} else {
		dbErr = tx.Repos.TVShows.CreateTVShow(ctxDb, database.CreateTVShowParams{
			UserID:   userId,
			ApiID:    tvShow.Id,
			Name:     tvShow.Name,
			Seasons:  int32(seasonNum),
			Episodes: newEpisodes,
			Runtime:  newRuntime,
			Status:   tvShow.Status,
		})
	}

	if dbErr != nil {
		h.app.Logger.Error(op, ctx, "Database operation failed", "error", dbErr.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Removing TV show from watchlist if present", "tv_id", tvShow.Id)
	if err = tx.Repos.Watchlists.DeleteWatchlist(ctxDb, tvShow.Id, userId); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to delete from watchlist", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Committing transaction")
	if err = tx.Commit(ctxDb); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to commit transaction", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	var episodesCount, runtimeCount int32
	if watchedSeasons > 0 {
		episodesCount = totalEpisodes
		runtimeCount = totalRuntime
	} else {
		episodesCount = newEpisodes
		runtimeCount = newRuntime
	}

	message := fmt.Sprintf(
		"The TV Show added as watched with below data:\nName: %v\nSeasons: %v\nEpisodes: %v\nRuntime: %v minutes",
		tvShow.Name, seasonNum, episodesCount, runtimeCount,
	)

	if _, err = ctx.Bot().Send(ctx.Chat(), message, telebot.ModeMarkdown); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send confirmation message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "TV show watch status updated successfully",
		"name", tvShow.Name, "seasons", seasonNum, "episodes", episodesCount, "runtime", runtimeCount)
	return nil
}

func (h *TVHandler) handleWatchlist(ctx telebot.Context, tvId string) error {
	const op = "tv.handleWatchlist"
	h.app.Logger.Info(op, ctx, "Adding TV show to watchlist", "tv_id", tvId)

	tvShowId, err := strconv.Atoi(tvId)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to parse TV show ID", "tv_id", tvId, "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Debug(op, ctx, "Retrieving TV show data from TMDB", "tv_id", tvShowId)
	tvShow, err := tv.GetTV(h.app, tvShowId, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to get TV show data", "error", err.Error())
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

	h.app.Logger.Debug(op, ctx, "Adding TV show to watchlist in database", "name", tvShow.Name)
	if err = h.app.Repository.Watchlists.CreateWatchlist(ctxDb, newWatchlist); err != nil {
		h.app.Logger.Error(op, ctx, "Failed to create watchlist entry", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	_, err = ctx.Bot().Send(ctx.Chat(), "Tv Show added to Watchlist", telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send confirmation message", "error", err.Error())
		return err
	}

	h.app.Logger.Info(op, ctx, "TV show successfully added to watchlist", "tv_id", tvShow.Id, "name", tvShow.Name)
	return nil
}

func (h *TVHandler) handleBackToPagination(ctx telebot.Context) error {
	const op = "tv.handleBackToPagination"
	h.app.Logger.Info(op, ctx, "Returning to paginated search results")

	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	err := ctx.Delete()
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to delete TV show details message", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	response, btn := paginators.GenerateTVResponse(paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
	_, err = ctx.Bot().Send(ctx.Chat(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to send paginated results", "error", err.Error())
		return err
	}

	err = ctx.Respond(&telebot.CallbackResponse{Text: messages.BackToSearchResults})
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to respond with callback", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Logger.Info(op, ctx, "Successfully returned to search results")
	return nil
}

func (h *TVHandler) handleNextPage(ctx telebot.Context) error {
	const op = "tv.handleNextPage"
	h.app.Logger.Info(op, ctx, "Moving to next page of search results")

	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]++
	if *pagePointer[userId] > maxPage[userId] {
		*pagePointer[userId] = maxPage[userId]
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	h.app.Logger.Debug(op, ctx, "Updating message with next page", "new_page", *pagePointer[userId])
	return updateTVMessage(ctx, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func (h *TVHandler) handlePrevPage(ctx telebot.Context) error {
	const op = "tv.handlePrevPage"
	h.app.Logger.Info(op, ctx, "Moving to previous page of search results")

	userId := ctx.Sender().ID

	if _, ok := tvCache[userId]; !ok {
		h.app.Logger.Warning(op, ctx, "No search results in cache for user")
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoSearchResult})
	}

	*pagePointer[userId]--
	if *pagePointer[userId] < 1 {
		*pagePointer[userId] = 1
	}

	paginatedTV := paginators.PaginateTV(tvCache[userId], *pagePointer[userId], tvCount[userId])
	h.app.Logger.Debug(op, ctx, "Updating message with previous page", "new_page", *pagePointer[userId])
	return updateTVMessage(ctx, paginatedTV, *pagePointer[userId], maxPage[userId], tvCount[userId])
}

func updateTVMessage(ctx telebot.Context, paginatedTV []tv.TV, currentPage, maxPage, tvCount int) error {
	response, btn := paginators.GenerateTVResponse(paginatedTV, currentPage, maxPage, tvCount)
	_, err := ctx.Bot().Edit(ctx.Message(), response, btn, telebot.ModeMarkdown)
	if err != nil {
		if strings.Contains(err.Error(), "message is not modified") {
			return ctx.Respond(&telebot.CallbackResponse{Text: messages.NoChanges})
		}
		return ctx.Send(messages.InternalError)
	}

	return ctx.Respond(&telebot.CallbackResponse{Text: messages.PageUpdated})
}

func (h *TVHandler) TVCallback(ctx telebot.Context) error {
	const op = "tv.TVCallback"
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)
	h.app.Logger.Info(op, ctx, "Processing TV callback", "callback_data", trimmed)

	if !strings.HasPrefix(trimmed, "tv|") {
		h.app.Logger.Warning(op, ctx, "Invalid callback prefix", "callback_data", trimmed)
		return nil
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "callback_data", callback.Data,
			"parts_count", len(dataParts))
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	data := dataParts[2]
	h.app.Logger.Debug(op, ctx, "Processing callback action", "action", action, "data", data)

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
		h.app.Logger.Warning(op, ctx, "Unknown callback action", "action", action)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
