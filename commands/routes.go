package commands

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/dependencyInjection"
	"strings"
)

func SetupDefaultRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		log.Printf("Unknown command: %s", c.Message().Text)
		return c.Send(fmt.Sprintf("Unknown %s command. Please use /help", c.Message().Text))
	})

	bot.Handle(telebot.OnCallback, handleCallback(container))
}

func SetupMovieRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/sm", container.MovieHandler.SearchMovie)
}

func SetupTVRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/stv", container.TVHandler.SearchTV)
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
		default:
			return c.Respond(&telebot.CallbackResponse{Text: "Unknown callback type"})
		}
	}
}
