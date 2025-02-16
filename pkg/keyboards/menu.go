package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
)

func (f *KeyboardFactory) handleSearchTVWrapper() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		f.app.Cache.UserCache.SetSearchStartTrue(c.Sender().ID, false)
		return c.Send(messages.MenuSearchTVResponse)
	}
}

func (f *KeyboardFactory) handleSearchMovieWrapper() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		f.app.Cache.UserCache.SetSearchStartTrue(c.Sender().ID, true)
		return c.Send(messages.MenuSearchMovieResponse)
	}
}

func (f *KeyboardFactory) handleWatchlist(c telebot.Context) error {
	return f.watchlistHandler.WatchlistInfo(c)
}

func (f *KeyboardFactory) handleInfo(c telebot.Context) error {
	return f.infoHandler.Info(c)
}

func (f *KeyboardFactory) loadMenuKeyboards(bot *telebot.Bot) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnSearchTV := menu.Text(messages.SearchTVLabel)
	btnSearchMovie := menu.Text(messages.SearchMovieLabel)
	btnWatchlist := menu.Text(messages.WatchlistLabel)
	btnInfo := menu.Text(messages.InfoLabel)

	menu.Reply(
		menu.Row(btnSearchTV, btnSearchMovie),
		menu.Row(btnWatchlist),
		menu.Row(btnInfo),
	)

	bot.Handle(&btnSearchTV, middleware.RequireRegistration(f.handleSearchTVWrapper(), f.app))
	bot.Handle(&btnSearchMovie, middleware.RequireRegistration(f.handleSearchMovieWrapper(), f.app))
	bot.Handle(&btnWatchlist, middleware.RequireRegistration(f.handleWatchlist, f.app))
	bot.Handle(&btnInfo, middleware.RequireRegistration(f.handleInfo, f.app))

	return menu
}
