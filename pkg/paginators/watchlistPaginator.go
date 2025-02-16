package paginators

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
)

func PaginateWatchlist(watchlist []models.Watchlist, page int) []models.Watchlist {
	const itemsPerPage = 3

	if page < 1 {
		page = 1
	}

	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage

	if len(watchlist) == 0 {
		return []models.Watchlist{}
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
