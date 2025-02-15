package routes

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/handlers"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/middleware"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/keyboards"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"gopkg.in/telebot.v3"
	"log"
	"strings"
)

func SetupDefaultRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	keyboards.LoadAllKeyboards(bot, container.DefaultHandler, app)

	bot.Handle(telebot.OnText, func(context telebot.Context) error {
		userId := context.Sender().ID
		isActive, userCache := app.Cache.UserCache.Fetch(userId)

		if !isActive {
			log.Printf("No active session for user %v", userId)
			return context.Send(messages.InternalError)
		}

		switch {
		case userCache.ApiToken.IsTokenWaiting:
			log.Printf("Handling API token input for user %d", userId)
			return container.DefaultHandler.HandleTextInput(context)

		case userCache.SearchState.IsSearchWaiting:
			return container.DefaultHandler.HandleReplySearch(context)

		default:
			log.Printf("Unknown command from user %d: %s", userId, context.Message().Text)
			return context.Send(fmt.Sprintf("Unknown input '%s'. Please use /help for available commands",
				context.Message().Text))
		}
	})

	bot.Handle(telebot.OnCallback, handleCallback(container))
	bot.Handle("/start", container.DefaultHandler.Start)
	bot.Handle("/token", middleware.RequireRegistration(container.DefaultHandler.GetToken, app))

	bot.Handle("/debug", func(context telebot.Context) error {
		return handlers.DebugMessage(context, app)
	})

}

func SetupMovieRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	bot.Handle("/sm", middleware.RequireRegistration(container.MovieHandler.SearchMovie, app))
}

func SetupTVRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	bot.Handle("/stv", middleware.RequireRegistration(container.TVHandler.SearchTV, app))
}

func SetupInfoRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	bot.Handle("/info", middleware.RequireRegistration(container.InfoHandler.Info, app))
}

func SetupWatchlistRoutes(bot *telebot.Bot, container *api.Resolver, app *appCfg.App) {
	bot.Handle("/w", middleware.RequireRegistration(container.WatchlistHandler.WatchlistInfo, app))
}

func handleCallback(container *api.Resolver) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		trimmed := strings.TrimSpace(c.Callback().Data)
		log.Print(trimmed)
		switch {
		case strings.HasPrefix(trimmed, "movie|"):
			return container.MovieHandler.MovieCallback(c)

		case strings.HasPrefix(trimmed, "tv|"):
			return container.TVHandler.TVCallback(c)

		case strings.HasPrefix(trimmed, "info|"):
			return container.InfoHandler.InfoCallback(c)

		case strings.HasPrefix(trimmed, "default|"):
			return container.DefaultHandler.DefaultCallback(c)

		case strings.HasPrefix(trimmed, "watchlist|"):
			return container.WatchlistHandler.WatchlistCallback(c)

		default:
			return c.Respond(&telebot.CallbackResponse{Text: "Unknown callback type"})
		}
	}
}
