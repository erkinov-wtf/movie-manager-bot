package app

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/encryption"
	"gorm.io/gorm"
)

type App struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
	Encryptor  *encryption.KeyEncryptor
}

func NewApp(cfg *config.Config, db *gorm.DB, client *tmdb.Client, cache *cache.Manager, encryptor *encryption.KeyEncryptor) *App {
	return &App{
		Cfg:        cfg,
		Database:   db,
		TMDBClient: client,
		Cache:      cache,
		Encryptor:  encryptor,
	}
}
