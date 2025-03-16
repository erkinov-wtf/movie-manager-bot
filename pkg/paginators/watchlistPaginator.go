package paginators

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
)

const itemsPerPage = 3

func PaginateWatchlist(watchlist []database.GetUserWatchlistsRow, page int) []database.GetUserWatchlistsRow {

	if page < 1 {
		page = 1
	}

	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage

	if len(watchlist) == 0 {
		return []database.GetUserWatchlistsRow{}
	}

	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(watchlist) {
		startIndex = (len(watchlist) - 1) / itemsPerPage * itemsPerPage
	}

	if endIndex > len(watchlist) {
		endIndex = len(watchlist)
	}

	return watchlist[startIndex:endIndex]
}

func PaginateWatchlistWithType(watchlist []database.GetUserWatchlistsWithTypeRow, page int) []database.GetUserWatchlistsWithTypeRow {

	if page < 1 {
		page = 1
	}

	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage

	if len(watchlist) == 0 {
		return []database.GetUserWatchlistsWithTypeRow{}
	}

	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(watchlist) {
		startIndex = (len(watchlist) - 1) / itemsPerPage * itemsPerPage
	}

	if endIndex > len(watchlist) {
		endIndex = len(watchlist)
	}

	return watchlist[startIndex:endIndex]
}
