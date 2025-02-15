package tv

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"gorm.io/gorm"
)

type TVHandler struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
}

func NewTVHandler(app *app.App) interfaces.TVInterface {
	return &TVHandler{
		Cfg:        app.Cfg,
		Database:   app.Database,
		TMDBClient: app.TMDBClient,
		Cache:      app.Cache,
	}
}
