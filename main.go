package main

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/commands"
	"movie-manager-bot/dependencyInjection"
	"os"
	"time"
)

func main() {
	pref := telebot.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(fmt.Sprintf("cant create new bot: %v", err.Error()))
		return
	}

	container := dependencyInjection.NewContainer()

	commands.SetupBotRoutes(bot, container)

	log.Print("starting bot...")
	bot.Start()
}
