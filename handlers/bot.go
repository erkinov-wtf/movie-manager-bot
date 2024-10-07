package handlers

import (
	"gopkg.in/telebot.v3"
	"log"
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
