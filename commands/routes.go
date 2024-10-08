package commands

import (
	"gopkg.in/telebot.v3"
	"movie-manager-bot/dependencyInjection"
)

func SetupBotRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/hello", container.BotHandler.Hello)
	bot.Handle("/search", container.BotHandler.Search)
	bot.Handle(telebot.OnCallback, container.BotHandler.OnCallback)
}
