package handlers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"movie-manager-bot/api/auth"
	"movie-manager-bot/api/media/search"
)

func (*botHandler) Hello(context telebot.Context) error {
	log.Print("/hello command received")
	err := context.Send("Hello mathafuck")
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (*botHandler) Search(context telebot.Context) error {
	log.Print("/search command received")

	if context.Message().Payload == "" {
		err := context.Send("after /search title must be provided")
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	msg, err := context.Bot().Send(context.Chat(), fmt.Sprintf("looking for *%v*...", context.Message().Payload), telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	err = auth.Login()
	if err != nil {
		log.Print(err)
		return err
	}

	movie, err := search.SearchMovie(context.Message().Payload)
	if err != nil {
		log.Print(err)
		return err
	}

	if movie.TotalResults == 0 {
		log.Print("no movies found")
		_, err = context.Bot().Edit(msg, fmt.Sprintf("no movies found for search *%s*", context.Message().Payload), telebot.ModeMarkdown)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}

	var response string
	for _, result := range movie.Results {
		response += fmt.Sprintf(
			"*Title*: %v\n"+
				"*Overview*: %v\n"+
				"*Release* Date: %s\n"+
				"*Runtime*: %v\n"+
				"*Is Adult*: %v\n"+
				"*Popularity*: %v\n\n",
			result.Title,
			result.Overview,
			result.ReleaseDate,
			result.Runtime,
			result.Adult,
			result.Popularity,
		)
	}

	_, err = context.Bot().Edit(msg, response, telebot.ModeMarkdown)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
