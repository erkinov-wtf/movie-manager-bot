package main

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/routes"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database/repository"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/encryption"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/workers"
	"gopkg.in/telebot.v3"
	"log"
	"time"
)

func main() {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Print("starting bot...")
	cfg := config.MustLoad()
	tmdbClient := tmdb.NewClient(cfg)
	log.Print("api client initialized")
	//db := repository.MustLoadDb(cfg)
	repoManager := repository.MustConnectDB(cfg, ctx)
	encryptor := encryption.NewKeyEncryptor(cfg.General.SecretKey)
	cacheManager := cache.NewCacheManager(repoManager, encryptor)

	appCfg := app.NewApp(cfg, repoManager, tmdbClient, cacheManager, encryptor)

	settings := telebot.Settings{
		Token:  cfg.General.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatal(fmt.Sprintf("cant create new bot: %v", err.Error()))
		return
	}

	resolver := api.NewResolver(appCfg)
	resolver.KeyboardFactory.LoadAllKeyboards(bot, resolver.DefaultHandler)

	routes.SetupDefaultRoutes(bot, resolver, appCfg)
	routes.SetupMovieRoutes(bot, resolver, appCfg)
	routes.SetupTVRoutes(bot, resolver, appCfg)
	routes.SetupInfoRoutes(bot, resolver, appCfg)
	routes.SetupWatchlistRoutes(bot, resolver, appCfg)
	log.Print("bot handlers setup")

	// Start the checker in a separate goroutine
	apiClient := workers.NewWorkerApiClient(50)
	checker := workers.NewTVShowChecker(appCfg, bot, apiClient)
	go checker.StartChecking(ctx, 7*24*time.Hour)

	log.Print("bot started")
	bot.Start()
}
