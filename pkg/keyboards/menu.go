package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/cache"
	messages2 "github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
)

func handleSearchTV(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, false)
	return c.Send(messages2.MenuSearchTVResponse)
}

func handleSearchMovie(c telebot.Context) error {
	cache.UserCache.SetSearchStartTrue(c.Sender().ID, true)
	return c.Send(messages2.MenuSearchMovieResponse)
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

	btnSearchTV := menu.Text(messages2.SearchTVLabel)
	btnSearchMovie := menu.Text(messages2.SearchMovieLabel)
	btnWatchlist := menu.Text(messages2.WatchlistLabel)
	btnInfo := menu.Text(messages2.InfoLabel)

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
