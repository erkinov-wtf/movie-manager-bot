package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	"gopkg.in/telebot.v3"
)

func LoadTokenRegistrationKeyboard(bot *telebot.Bot, handlers interfaces.DefaultInterface) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnSearch := menu.Text("Get Token")
	menu.Reply(
		menu.Row(btnSearch),
	)

	bot.Handle(&btnSearch, middleware.RequireRegistration(handlers.GetToken))

	return menu
}
