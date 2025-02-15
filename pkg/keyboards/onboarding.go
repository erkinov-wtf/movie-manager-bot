package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"gopkg.in/telebot.v3"
)

func LoadTokenRegistrationKeyboard(bot *telebot.Bot, handlers interfaces.DefaultInterface, app *appCfg.App) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnSearch := menu.Text("Get Token")
	menu.Reply(
		menu.Row(btnSearch),
	)

	bot.Handle(&btnSearch, middleware.RequireRegistration(handlers.GetToken, app))

	return menu
}
