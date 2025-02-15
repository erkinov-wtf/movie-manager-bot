package interfaces

import "gopkg.in/telebot.v3"

type TVInterface interface {
	SearchTV(context telebot.Context) error
	TVCallback(context telebot.Context) error
}
