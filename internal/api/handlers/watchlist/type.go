package watchlist

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type WatchlistHandler struct {
	App *app.App
}

func NewWatchlistHandler(app *app.App) interfaces.WatchlistInterface {
	return &WatchlistHandler{
		App: app,
	}
}

const itemsPerPage = 3
