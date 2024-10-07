package interfaces

import "gopkg.in/telebot.v3"

type BotInterface interface {
	Hello(context telebot.Context) error
	Search(context telebot.Context) error
}
