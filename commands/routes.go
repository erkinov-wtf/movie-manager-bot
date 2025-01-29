package commands

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/dependencyInjection"
	"github.com/erkinov-wtf/movie-manager-bot/handlers"
	"github.com/erkinov-wtf/movie-manager-bot/middleware"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"gopkg.in/telebot.v3"
	"log"
	"strings"
)

func SetupDefaultRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle(telebot.OnText, func(context telebot.Context) error {
		// Checking if bot waits for user's api key input
		_, active, token := cache.UserCache.Get(context.Sender().ID)
		if active && token.IsTokenWaiting {
			log.Print("redirecting to api text input")
			return container.DefaultHandler.HandleTextInput(context)
		}
		log.Printf("Unknown command: %s", context.Message().Text)
		return context.Send(fmt.Sprintf("Unknown %s command. Please use /help", context.Message().Text))
	})

	bot.Handle(telebot.OnCallback, handleCallback(container))
	bot.Handle("/start", container.DefaultHandler.Start)
	bot.Handle("/token", middleware.RequireRegistration(container.DefaultHandler.GetToken))

	bot.Handle("/debug", handlers.DebugMessage)

}

func SetupMovieRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/sm", middleware.RequireRegistration(container.MovieHandler.SearchMovie))
}

func SetupTVRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/stv", middleware.RequireRegistration(container.TVHandler.SearchTV))
}

func SetupInfoRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/info", middleware.RequireRegistration(container.InfoHandler.Info))
}

func SetupWatchlistRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/w", middleware.RequireRegistration(container.WatchlistHandler.WatchlistInfo))
}

func handleCallback(container *dependencyInjection.Container) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		trimmed := strings.TrimSpace(c.Callback().Data)
		log.Print(trimmed)
		switch {
		case strings.HasPrefix(trimmed, "movie|"):
			return container.MovieHandler.MovieCallback(c)

		case strings.HasPrefix(trimmed, "tv|"):
			return container.TVHandler.TVCallback(c)

		case strings.HasPrefix(trimmed, "info|"):
			return container.InfoHandler.InfoCallback(c)

		case strings.HasPrefix(trimmed, "default|"):
			return container.DefaultHandler.DefaultCallback(c)

		case strings.HasPrefix(trimmed, "watchlist|"):
			return container.WatchlistHandler.WatchlistCallback(c)

		default:
			return c.Respond(&telebot.CallbackResponse{Text: "Unknown callback type"})
		}
	}
}
