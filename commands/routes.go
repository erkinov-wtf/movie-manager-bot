package commands

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/dependencyInjection"
	"movie-manager-bot/middleware"
	"strings"
)

func SetupDefaultRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		log.Printf("Unknown command: %s", c.Message().Text)
		return c.Send(fmt.Sprintf("Unknown %s command. Please use /help", c.Message().Text))
	})

	bot.Handle(telebot.OnCallback, handleCallback(container))
	bot.Handle("/start", container.DefaultHandler.Start)
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

		default:
			return c.Respond(&telebot.CallbackResponse{Text: "Unknown callback type"})
		}
	}
}
