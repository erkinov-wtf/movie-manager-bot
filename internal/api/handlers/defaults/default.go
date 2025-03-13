package defaults

import (
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/messages"
	"github.com/erkinov-wtf/movie-manager-bot/pkg/utils"
	"gopkg.in/telebot.v3"
	"log"
	"time"

	"strings"
)

func (h *DefaultHandler) Start(ctx telebot.Context) error {
	log.Print(messages.StartCommand)

	ctxDb, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	userExists, err := h.app.Repository.Users.UserExists(ctxDb, ctx.Sender().ID)
	if err != nil {
		log.Println(err)
		return ctx.Send(messages.InternalError)
	}

	if !userExists {
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
			log.Printf(err.Error())
			return ctx.Send(messages.InternalError)
		}
	} else {
		menu := h.keyboards.LoadMenu(ctx.Bot())
		err = ctx.Send(messages.UseHelp, telebot.ModeMarkdown, menu)
		if err != nil {
			log.Print(err)
			return ctx.Send(messages.InternalError)
		}
	}

	return nil
}

func (h *DefaultHandler) handleStartCallback(ctx telebot.Context) error {
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

	err := h.app.Repository.Users.CreateUser(ctxDb, newUser)
	if err != nil {
		log.Printf("cant create user: %v\n", err.Error())
		return ctx.Send(messages.InternalError)
	}

	keyboard := h.keyboards.LoadTokenRegistration(ctx.Bot(), h)
	return ctx.Send(messages.Registered, keyboard, telebot.ModeMarkdown)
}

func (h *DefaultHandler) GetToken(ctx telebot.Context) error {
	userId := ctx.Sender().ID
	isActive, userCache := h.app.Cache.UserCache.Get(userId)

	if isActive && !userCache.ApiToken.IsTokenWaiting {
		return ctx.Send(messages.TokenAlreadyExists, telebot.ModeMarkdown)
	}

	return ctx.Send(messages.TokenInstructions, telebot.ModeMarkdown)
}

func (h *DefaultHandler) HandleReplySearch(ctx telebot.Context) error {
	_, uc := h.app.Cache.UserCache.Get(ctx.Sender().ID)

	if uc.SearchState.IsTVShowSearch {
		return h.handleTVShowSearch(ctx)
	}
	return h.handleMovieSearch(ctx)
}

func (h *DefaultHandler) handleTVShowSearch(ctx telebot.Context) error {
	h.app.Cache.UserCache.SetSearchStartFalse(ctx.Sender().ID)
	return h.tvHandler.SearchTV(ctx)
}

func (h *DefaultHandler) handleMovieSearch(ctx telebot.Context) error {
	h.app.Cache.UserCache.SetSearchStartFalse(ctx.Sender().ID)
	return h.movieHandler.SearchMovie(ctx)
}

func (h *DefaultHandler) HandleTextInput(ctx telebot.Context) error {
	userId := ctx.Sender().ID
	inputText := ctx.Message().Text

	if !utils.TestApiToken(h.app, inputText) {
		return ctx.Send(messages.TokenTestFailed, telebot.ModeMarkdown)
	}

	encrypted, err := h.app.Encryptor.Encrypt(inputText)
	if err != nil {
		log.Printf("error encrypting api key: %v", err.Error())
		return ctx.Send(messages.InternalError)
	}

	ctxDb, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = h.app.Repository.Users.UpdateUserTMDBKey(ctxDb, ctx.Sender().ID, encrypted)
	if err != nil {
		log.Print(err)
		return ctx.Send(messages.InternalError)
	}

	h.app.Cache.UserCache.UpdateTokenState(userId, false)
	menu := h.keyboards.LoadMenu(ctx.Bot())

	return ctx.Send(messages.TokenSaved, menu, telebot.ModeMarkdown)
}

func (h *DefaultHandler) DefaultCallback(ctx telebot.Context) error {
	callback := ctx.Callback()
	trimmed := strings.TrimSpace(callback.Data)

	if !strings.HasPrefix(trimmed, "default|") {
		return ctx.Send(messages.InternalError)
	}

	dataParts := strings.Split(trimmed, "|")
	if len(dataParts) != 3 {
		log.Printf("Received malformed callback data: %s", callback.Data)
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.MalformedData})
	}

	action := dataParts[1]
	switch action {
	case "start":
		return h.handleStartCallback(ctx)

	default:
		return ctx.Respond(&telebot.CallbackResponse{Text: messages.UnknownAction})
	}
}
