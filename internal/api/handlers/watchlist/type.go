package watchlist

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"gorm.io/gorm"
)

type WatchlistHandler struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
}

func NewWatchlistHandler(app *app.App) interfaces.WatchlistInterface {
	return &WatchlistHandler{
		Cfg:        app.Cfg,
		Database:   app.Database,
		TMDBClient: app.TMDBClient,
		Cache:      app.Cache,
	}
}

const itemsPerPage = 3
