package info

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"gorm.io/gorm"
)

type InfoHandler struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
}

func NewInfoHandler(app *app.App) interfaces.InfoInterface {
	return &InfoHandler{
		Cfg:        app.Cfg,
		Database:   app.Database,
		TMDBClient: app.TMDBClient,
		Cache:      app.Cache,
	}
}

type tvStats struct {
	amount    int
	totalTime int64
}

type movieStats struct {
	amount    int
	totalTime int64
}
