package commands

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/dependencyInjection"
	"movie-manager-bot/middleware"
	"movie-manager-bot/storage/cache"
	"strings"
	"time"
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

	// for dev debugging only
	bot.Handle("/debug", func(context telebot.Context) error {
		// Collect user info
		user := context.Sender()
		message := context.Message()

		// Log detailed user and message info
		log.Printf("Debug Info - Timestamp: %v", time.Now())
		log.Printf("User Info: ID=%d, Username=%s, FirstName=%s, LastName=%s",
			user.ID, user.Username, user.FirstName, user.LastName)
		log.Printf("Message Info: Text=%s, Payload=%s, Date=%s",
			message.Text, message.Payload, message.Time().Format("2006-01-02 15:04:05"))

		// Send debug response to user
		debugMessage := fmt.Sprintf("Hello %s! Here is your debug info:\n\n", user.FirstName)
		debugMessage += fmt.Sprintf("User ID: %d\nUsername: %s\nFirst Name: %s\nLast Name: %s\n",
			user.ID, user.Username, user.FirstName, user.LastName)
		debugMessage += fmt.Sprintf("Message Text: %s\nMessage Payload: %s\nMessage Date: %s\n",
			message.Text, message.Payload, message.Time().Format("2006-01-02 15:04:05"))
		debugMessage += fmt.Sprint("============\n")

		//cache data retrieval
		cacheValue, cacheExpired, token := cache.UserCache.Get(context.Sender().ID)
		debugMessage += fmt.Sprintf("Current User Cache \nCache value: %v\nIs Cache Active: %v\nIs Token Waiting: %v\nToken value: %v\n",
			cacheValue, cacheExpired, token.IsTokenWaiting, token.Token)

		return context.Send(debugMessage)
	})

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
