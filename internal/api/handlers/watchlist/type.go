package watchlist

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
)

type WatchlistHandler struct {
	app *app.App
}

func NewWatchlistHandler(app *app.App) interfaces.WatchlistInterface {
	return &WatchlistHandler{
		app: app,
	}
}

const itemsPerPage = 3
