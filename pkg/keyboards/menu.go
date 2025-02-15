package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/info"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers/watchlist"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
)

func handleSearchTVWrapper(app *appCfg.App) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		app.Cache.UserCache.SetSearchStartTrue(c.Sender().ID, false)
		return c.Send(messages.MenuSearchTVResponse)
	}
}

func handleSearchMovieWrapper(app *appCfg.App) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		app.Cache.UserCache.SetSearchStartTrue(c.Sender().ID, true)
		return c.Send(messages.MenuSearchMovieResponse)
	}
}

func handleWatchlist(c telebot.Context) error {
	w := &watchlist.WatchlistHandler{}
	return w.WatchlistInfo(c)
}

func handleInfo(c telebot.Context) error {
	i := &info.InfoHandler{}
	return i.Info(c)
}

func LoadMenuKeyboards(bot *telebot.Bot, app *appCfg.App) *telebot.ReplyMarkup {
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

	bot.Handle(&btnSearchTV, middleware.RequireRegistration(handleSearchTVWrapper(app), app))
	bot.Handle(&btnSearchMovie, middleware.RequireRegistration(handleSearchMovieWrapper(app), app))
	bot.Handle(&btnWatchlist, middleware.RequireRegistration(handleWatchlist, app))
	bot.Handle(&btnInfo, middleware.RequireRegistration(handleInfo, app))

	return menu
}
