package watchlist

import "github.com/erkinov-wtf/movie-manager-bot/interfaces"

type watchlistHandler struct{}

func NewWatchlistHandler() interfaces.WatchlistInterface {
	return &watchlistHandler{}
}

const itemsPerPage = 3
