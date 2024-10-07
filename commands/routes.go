package commands

import (
	"gopkg.in/telebot.v3"
	"movie-manager-bot/dependencyInjection"
)

func SetupBotRoutes(bot *telebot.Bot, container *dependencyInjection.Container) {
	bot.Handle("/hello", container.BotHandler.Hello)
}
