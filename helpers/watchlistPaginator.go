package helpers

import "movie-manager-bot/models"

func PaginateWatchlist(watchlist []models.Watchlist, page int) []models.Watchlist {
	const itemsPerPage = 3

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if start >= len(watchlist) {
		return []models.Watchlist{}
	}
	if end > len(watchlist) {
		end = len(watchlist)
	}

	return watchlist[start:end]
}
