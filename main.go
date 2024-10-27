package main

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api"
	"movie-manager-bot/commands"
	"movie-manager-bot/config"
	"movie-manager-bot/dependencyInjection"
	"movie-manager-bot/storage/firebase"
	"time"
)

func main() {
	config.MustLoad()
	api.NewClient()
	firebase.InitFirebase()

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

	log.Print("starting bot...")
	bot.Start()
}
