package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/helpers/messages"
	"github.com/erkinov-wtf/movie-manager-bot/middleware"
	"github.com/erkinov-wtf/movie-manager-bot/storage/cache"
	"gopkg.in/telebot.v3"
)

func handleSearchTV(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, false)
	return c.Send(messages.MenuSearchTVResponse)
}

func handleSearchMovie(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, true)
	return c.Send(messages.MenuSearchMovieResponse)
}

func handleWatchlist(c telebot.Context) error {
	w := &watchlist.WatchlistHandler{}
	return w.WatchlistInfo(c)
}

func handleInfo(c telebot.Context) error {
	i := &info.InfoHandler{}
	return i.Info(c)
}

func LoadMenuKeyboards(bot *telebot.Bot) *telebot.ReplyMarkup {
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

	bot.Handle(&btnSearchTV, middleware.RequireRegistration(handleSearchTV))
	bot.Handle(&btnSearchMovie, middleware.RequireRegistration(handleSearchMovie))
	bot.Handle(&btnWatchlist, middleware.RequireRegistration(handleWatchlist))
	bot.Handle(&btnInfo, middleware.RequireRegistration(handleInfo))

	return menu
}
