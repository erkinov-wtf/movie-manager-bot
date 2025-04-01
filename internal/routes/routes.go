package routes

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"strings"
)

func SetupDefaultRoutes(bot *telebot.Bot, resolver *api.Resolver, app *appCfg.App) {
	const op = "routes.SetupDefaultRoutes"
	resolver.KeyboardFactory.LoadAllKeyboards(bot, resolver.DefaultHandler)

	bot.Handle(telebot.OnText, func(context telebot.Context) error {
		const handlerOp = "routes.OnTextHandler"
		userId := context.Sender().ID

		app.Logger.Debug(handlerOp, context, "Handling text message",
			"user_id", userId, "text", context.Message().Text)

		isActive, userCache := app.Cache.UserCache.Fetch(userId)

		if !isActive {
			app.Logger.Warning(handlerOp, context, "No active session for user")
			return context.Send(messages.InternalError)
		}

		switch {
		case userCache.ApiToken.IsTokenWaiting:
			app.Logger.Info(handlerOp, context, "Handling API token input")
			return resolver.DefaultHandler.HandleTextInput(context)

		case userCache.SearchState.IsSearchWaiting:
			app.Logger.Info(handlerOp, context, "Handling search reply")
			return resolver.DefaultHandler.HandleReplySearch(context)

		default:
			app.Logger.Info(handlerOp, context, "Unknown command from user",
				"user_id", userId, "text", context.Message().Text)
			return context.Send(fmt.Sprintf("Unknown input '%s'. Please use /help for available commands",
				context.Message().Text))
		}
	})

	bot.Handle(telebot.OnCallback, handleCallback(resolver, app))

	bot.Handle("/start", resolver.DefaultHandler.Start)
	bot.Handle("/token", middleware.RequireRegistration(resolver.DefaultHandler.GetToken, app))

	bot.Handle("/debug", func(context telebot.Context) error {
		app.Logger.Info(op+".", context, "Debug command triggered")
		return handlers.DebugMessage(context, app)
	})
}

func SetupMovieRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	const op = "routes.SetupMovieRoutes"
	bot.Handle("/sm", middleware.RequireTMDBToken(container.MovieHandler.SearchMovie, app))
}

func SetupTVRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	const op = "routes.SetupTVRoutes"
	bot.Handle("/stv", middleware.RequireTMDBToken(container.TVHandler.SearchTV, app))
}

func SetupInfoRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	const op = "routes.SetupInfoRoutes"
	bot.Handle("/info", middleware.RequireTMDBToken(container.InfoHandler.Info, app))
}

func SetupWatchlistRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	const op = "routes.SetupWatchlistRoutes"
	bot.Handle("/w", middleware.RequireTMDBToken(container.WatchlistHandler.WatchlistInfo, app))
}

func handleCallback(container *api.Resolver, app *appCfg.App) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		const op = "routes.handleCallback"
		trimmed := strings.TrimSpace(c.Callback().Data)
		app.Logger.Info(op, c, "Processing callback", "callback_data", trimmed)

		switch {
		case strings.HasPrefix(trimmed, "movie|"):
			app.Logger.Debug(op, c, "Routing to movie callback handler")
			return container.MovieHandler.MovieCallback(c)

		case strings.HasPrefix(trimmed, "tv|"):
			app.Logger.Debug(op, c, "Routing to TV callback handler")
			return container.TVHandler.TVCallback(c)

		case strings.HasPrefix(trimmed, "info|"):
			app.Logger.Debug(op, c, "Routing to info callback handler")
			return container.InfoHandler.InfoCallback(c)

		case strings.HasPrefix(trimmed, "default|"):
			app.Logger.Debug(op, c, "Routing to default callback handler")
			return container.DefaultHandler.DefaultCallback(c)

		case strings.HasPrefix(trimmed, "watchlist|"):
			app.Logger.Debug(op, c, "Routing to watchlist callback handler")
			return container.WatchlistHandler.WatchlistCallback(c)

		default:
			app.Logger.Warning(op, c, "Unknown callback type received", "callback_data", trimmed)
			return c.Respond(&telebot.CallbackResponse{Text: "Unknown callback type"})
		}
	}
}
