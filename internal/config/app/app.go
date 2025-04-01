package app

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database/repository"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/encryption"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils/logger"
)

type App struct {
	Cfg        *config.Config
	Repository *repository.Manager
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
	Encryptor  *encryption.KeyEncryptor
	Logger     *logger.Logger
}

func NewApp(cfg *config.Config, repos *repository.Manager, client *tmdb.Client, cache *cache.Manager, encryptor *encryption.KeyEncryptor, logger *logger.Logger) *App {
	return &App{
		Cfg:        cfg,
		Repository: repos,
		TMDBClient: client,
		Cache:      cache,
		Encryptor:  encryptor,
		Logger:     logger,
	}
}
