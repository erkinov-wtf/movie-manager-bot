package app

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"gorm.io/gorm"
)

type App struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
}

func NewApp(cfg *config.Config, db *gorm.DB, client *tmdb.Client, ca *cache.Manager) *App {
	return &App{
		Cfg:        cfg,
		Database:   db,
		TMDBClient: client,
		Cache:      ca,
	}
}
