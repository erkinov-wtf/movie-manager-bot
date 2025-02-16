package interfaces

import (
	"gopkg.in/telebot.v3"
)

type DefaultInterface interface {
	Start(context telebot.Context) error
	GetToken(context telebot.Context) error
	HandleReplySearch(context telebot.Context) error
	HandleTextInput(context telebot.Context) error
	DefaultCallback(context telebot.Context) error
}
