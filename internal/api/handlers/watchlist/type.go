package watchlist

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
)

type WatchlistHandler struct{}

func NewWatchlistHandler() interfaces.WatchlistInterface {
	return &WatchlistHandler{}
}

const itemsPerPage = 3
