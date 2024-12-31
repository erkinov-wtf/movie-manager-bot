package main

import (
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/commands"
	"movie-manager-bot/config"
	"movie-manager-bot/dependencyInjection"
	"movie-manager-bot/helpers/workers"
	"movie-manager-bot/storage/cache"
	"movie-manager-bot/storage/database"
	"time"
)

func main() {
	log.Print("starting bot...")
	config.MustLoad()
	api.NewClient()
	log.Print("api client initialized")
	database.DBConnect()
	cache.NewUserCache()

	settings := telebot.Settings{
		Token:  config.Cfg.General.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatal(fmt.Sprintf("cant create new bot: %v", err.Error()))
		return
	}

	container := dependencyInjection.NewContainer()

	commands.SetupDefaultRoutes(bot, container)
	commands.SetupMovieRoutes(bot, container)
	commands.SetupTVRoutes(bot, container)
	commands.SetupInfoRoutes(bot, container)
	commands.SetupWatchlistRoutes(bot, container)
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
