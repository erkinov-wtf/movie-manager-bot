package main

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/api"
	"github.com/erkinov-wtf/movie-manager-bot/commands"
	"github.com/erkinov-wtf/movie-manager-bot/config"
	"github.com/erkinov-wtf/movie-manager-bot/dependencyInjection"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/workers"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"github.com/erkinov-wtf/movie-manager-bot/storage/database"
	"gopkg.in/telebot.v3"
	"log"
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

	menu := &telebot.ReplyMarkup{}
	btnSearch := menu.Text("Search")
	btnHelp := menu.Text("Help")

	// Add buttons to the reply keyboard
	menu.Reply(
		menu.Row(btnSearch, btnHelp), // Buttons in a single row
	)

	// Handle the "Search" button
	bot.Handle(&btnSearch, func(c telebot.Context) error {
		return c.Send("You pressed the Search button!")
	})

	// Handle the "Help" button
	bot.Handle(&btnHelp, func(c telebot.Context) error {
		return c.Send("Here’s some help info.")
	})

	bot.Handle("/starts", func(c telebot.Context) error {
		return c.Send("Welcome! Choose an option:", menu)
	})

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
