package main

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/dependencyInjection"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/routes"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/workers"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"time"
)

type App struct {
	Cfg        *config.Config
	Database   *gorm.DB
	TMDBClient *tmdb.Client
	Cache      *cache.Manager
}

func main() {
	log.Print("starting bot...")
	cfg := config.MustLoad()
	tmdbClient := tmdb.NewClient(cfg)
	log.Print("api client initialized")
	db := database.MustLoadDb(cfg)
	cacheManager := cache.NewCacheManager()

	app := App{
		Cfg:        cfg,
		Database:   db,
		TMDBClient: tmdbClient,
		Cache:      cacheManager,
	}

	settings := telebot.Settings{
		Token:  cfg.General.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatal(fmt.Sprintf("cant create new bot: %v", err.Error()))
		return
	}

	container := dependencyInjection.NewContainer()

	routes.SetupDefaultRoutes(bot, container)
	routes.SetupMovieRoutes(bot, container)
	routes.SetupTVRoutes(bot, container)
	routes.SetupInfoRoutes(bot, container)
	routes.SetupWatchlistRoutes(bot, container)
	log.Print("bot handlers setup")

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the checker in a separate goroutine
	apiClient := workers.NewWorkerApiClient(50)
	checker := workers.NewTVShowChecker(database.DB, bot, apiClient)
	go checker.StartChecking(ctx, 336*time.Hour)

	log.Print("bot started")
	bot.Start()
}
