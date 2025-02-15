package main

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/routes"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/workers"
	"gopkg.in/telebot.v3"
	"log"
	"time"
)

func main() {
	log.Print("starting bot...")
	cfg := config.MustLoad()
	tmdbClient := tmdb.NewClient(cfg)
	log.Print("api client initialized")
	db := database.MustLoadDb(cfg)
	cacheManager := cache.NewCacheManager(db)

	appCfg := app.NewApp(cfg, db, tmdbClient, cacheManager)

	settings := telebot.Settings{
		Token:  cfg.General.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatal(fmt.Sprintf("cant create new bot: %v", err.Error()))
		return
	}

	container := api.NewResolver(appCfg)

	routes.SetupDefaultRoutes(bot, container, appCfg)
	routes.SetupMovieRoutes(bot, container, appCfg)
	routes.SetupTVRoutes(bot, container, appCfg)
	routes.SetupInfoRoutes(bot, container, appCfg)
	routes.SetupWatchlistRoutes(bot, container, appCfg)
	log.Print("bot handlers setup")

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the checker in a separate goroutine
	apiClient := workers.NewWorkerApiClient(50)
	checker := workers.NewTVShowChecker(appCfg, bot, apiClient)
	go checker.StartChecking(ctx, 336*time.Hour)

	log.Print("bot started")
	bot.Start()
}
