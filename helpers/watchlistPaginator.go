package helpers

import "movie-manager-bot/models"

func PaginateWatchlist(watchlist []models.Watchlist, page int) []models.Watchlist {
	const itemsPerPage = 3

	// Ensure page is at least 1
	if page < 1 {
		page = 1
	}

	// Calculate start and end indices for the current page
	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage

	// Handle empty watchlist
	if len(watchlist) == 0 {
		return []models.Watchlist{}
	}

	// Ensure startIndex is not negative and within bounds
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(watchlist) {
		startIndex = (len(watchlist) - 1) / itemsPerPage * itemsPerPage
	}

	// Ensure endIndex doesn't exceed slice length
	if endIndex > len(watchlist) {
		endIndex = len(watchlist)
	}

	return watchlist[startIndex:endIndex]
}
