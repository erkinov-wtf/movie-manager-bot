package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/middleware"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"gopkg.in/telebot.v3"
)

func handleSearchTV(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, false)
	return c.Send("search tv series")
}

func handleSearchMovie(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, true)
	return c.Send("search movie")
}

func LoadMenuKeyboards(bot *telebot.Bot) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnSearch := menu.Text("Search TV")
	btnHelp := menu.Text("Search Movie")

	menu.Reply(
		menu.Row(btnSearch, btnHelp),
	)

	bot.Handle(&btnSearch, middleware.RequireRegistration(handleSearchTV))
	bot.Handle(&btnHelp, middleware.RequireRegistration(handleSearchMovie))

	return menu
}
