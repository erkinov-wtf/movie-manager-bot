package interfaces

import "gopkg.in/telebot.v3"

type DefaultInterface interface {
	Start(context telebot.Context) error
	DefaultCallback(context telebot.Context) error
}
