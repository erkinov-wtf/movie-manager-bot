package defaults

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"strings"
	"time"
)

func (h *DefaultHandler) Start(ctx telebot.Context) error {
	const op = "defaults.Start"
	h.app.Logger.Info(op, ctx, "Start command received")

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	userExists, err := h.app.Repository.Users.UserExists(ctxDb, ctx.Sender().ID)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to check if user exists", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	if !userExists {
		h.app.Logger.Info(op, ctx, "New user detected, showing privacy policy")

		btn := &telebot.ReplyMarkup{}
		btnRows := []telebot.Row{
			btn.Row(btn.Data("✅ I Agree", "", fmt.Sprint("default|start|"))),
		}

		btn.Inline(btnRows...)
		err = ctx.Send(messages.PrivacyPolicy, &telebot.SendOptions{
			ParseMode:   telebot.ModeMarkdown,
			ReplyMarkup: btn,
		})
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to send privacy policy", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
	} else {
		h.app.Logger.Info(op, ctx, "Existing user detected, showing help menu")
		menu := h.keyboards.LoadMenu(ctx.Bot())
		err = ctx.Send(messages.UseHelp, telebot.ModeMarkdown, menu)
		if err != nil {
			h.app.Logger.Error(op, ctx, "Failed to send help message", "error", err.Error())
			return ctx.Send(messages.InternalError)
		}
	}

	h.app.Logger.Info(op, ctx, "Start command handled successfully")
	return nil
}

func (h *DefaultHandler) handleStartCallback(ctx telebot.Context) error {
	const op = "defaults.handleStartCallback"
	h.app.Logger.Info(op, ctx, "Processing start callback")

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	newUser := database.CreateUserParams{
		TgID:       ctx.Sender().ID,
		FirstName:  &ctx.Sender().FirstName,
		LastName:   &ctx.Sender().LastName,
		Username:   &ctx.Sender().Username,
		Language:   ctx.Sender().LanguageCode,
		TmdbApiKey: nil,
	}

	h.app.Logger.Debug(op, ctx, "Creating new user in database",
		"username", ctx.Sender().Username, "language", ctx.Sender().LanguageCode)
	err := h.app.Repository.Users.CreateUser(ctxDb, newUser)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to create user", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	keyboard := h.keyboards.LoadTokenRegistration(ctx.Bot(), h)
	h.app.Logger.Info(op, ctx, "User registered successfully")
	return ctx.Send(messages.Registered, keyboard, telebot.ModeMarkdown)
}

func (h *DefaultHandler) GetToken(ctx telebot.Context) error {
	const op = "defaults.GetToken"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Token request received")

	isActive, userCache := h.app.Cache.UserCache.Get(userId)

	if isActive && !userCache.ApiToken.IsTokenWaiting {
		h.app.Logger.Warning(op, ctx, "User already has token, rejecting request")
		return ctx.Send(messages.TokenAlreadyExists, telebot.ModeMarkdown)
	}

	h.app.Logger.Info(op, ctx, "Sending token instructions")
	return ctx.Send(messages.TokenInstructions, telebot.ModeMarkdown)
}

func (h *DefaultHandler) HandleReplySearch(ctx telebot.Context) error {
	const op = "defaults.HandleReplySearch"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Processing search reply")

	_, uc := h.app.Cache.UserCache.Fetch(userId)

	if uc.SearchState.IsTVShowSearch {
		h.app.Logger.Info(op, ctx, "TV show search detected")
		return h.handleTVShowSearch(ctx)
	}

	h.app.Logger.Info(op, ctx, "Movie search detected")
	return h.handleMovieSearch(ctx)
}

func (h *DefaultHandler) handleTVShowSearch(ctx telebot.Context) error {
	const op = "defaults.handleTVShowSearch"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Processing TV show search")

	h.app.Cache.UserCache.SetSearchStartFalse(userId)

	return h.tvHandler.SearchTV(ctx)
}

func (h *DefaultHandler) handleMovieSearch(ctx telebot.Context) error {
	const op = "defaults.handleMovieSearch"
	userId := ctx.Sender().ID
	h.app.Logger.Info(op, ctx, "Processing movie search")

	h.app.Cache.UserCache.SetSearchStartFalse(userId)

	return h.movieHandler.SearchMovie(ctx)
}

func (h *DefaultHandler) HandleTextInput(ctx telebot.Context) error {
	const op = "defaults.HandleTextInput"
	userId := ctx.Sender().ID
	inputText := ctx.Message().Text
	h.app.Logger.Info(op, ctx, "Processing text input")

	h.app.Logger.Debug(op, ctx, "Testing API token")
	if !utils.TestApiToken(h.app, inputText) {
		h.app.Logger.Warning(op, ctx, "API token test failed")
		return ctx.Send(messages.TokenTestFailed, telebot.ModeMarkdown)
	}

	h.app.Logger.Debug(op, ctx, "Encrypting API token")
	encrypted, err := h.app.Encryptor.Encrypt(inputText)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to encrypt API token", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.app.Logger.Debug(op, ctx, "Updating user's TMDB key in database")
	err = h.app.Repository.Users.UpdateUserTMDBKey(ctxDb, userId, encrypted)
	if err != nil {
		h.app.Logger.Error(op, ctx, "Failed to update user's TMDB key", "error", err.Error())
		return ctx.Send(messages.InternalError)
	}

	h.app.Cache.UserCache.UpdateTokenState(userId, false)

	menu := h.keyboards.LoadMenu(ctx.Bot())
	h.app.Logger.Info(op, ctx, "Token saved successfully")
	return ctx.Send(messages.TokenSaved, menu, telebot.ModeMarkdown)
}

func (h *DefaultHandler) DefaultCallback(ctx telebot.Context) error {
	const op = "defaults.DefaultCallback"
	callback := ctx.Callback()
	h.app.Logger.Info(op, ctx, "Processing callback data", "callback_data", callback.Data)

	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "default|") {
		h.app.Logger.Warning(op, ctx, "Invalid callback prefix", "callback_data", trimmed)
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		h.app.Logger.Warning(op, ctx, "Malformed callback data", "callback_data", callback.Data,
			"parts_count", len(dataParts))
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	h.app.Logger.Debug(op, ctx, "Processing callback action", "action", action)

	switch action {
	case "start":
		return h.handleStartCallback(ctx)
	default:
		h.app.Logger.Warning(op, ctx, "Unknown callback action", "action", action)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
