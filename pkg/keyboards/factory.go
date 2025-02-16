package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"gopkg.in/telebot.v3"
)

type KeyboardFactory struct {
	app              *appCfg.App
	watchlistHandler interfaces.WatchlistInterface
	infoHandler      interfaces.InfoInterface
}

func NewKeyboardFactory(app *appCfg.App, w interfaces.WatchlistInterface, i interfaces.InfoInterface) *KeyboardFactory {
	return &KeyboardFactory{
		app:              app,
		watchlistHandler: w,
		infoHandler:      i,
	}
}

func (f *KeyboardFactory) LoadAllKeyboards(bot *telebot.Bot, defaultHandler interfaces.DefaultInterface) {
	f.loadMenuKeyboards(bot)
	f.loadTokenRegistrationKeyboard(bot, defaultHandler, f.app)
}

// LoadMenu public wrapper for internal loadMenuKeyboards method
func (f *KeyboardFactory) LoadMenu(bot *telebot.Bot) *telebot.ReplyMarkup {
	return f.loadMenuKeyboards(bot)
}

// LoadTokenRegistration public wrapper for internal loadTokenRegistrationKeyboard method
func (f *KeyboardFactory) LoadTokenRegistration(bot *telebot.Bot, h interfaces.DefaultInterface) *telebot.ReplyMarkup {
	return f.loadTokenRegistrationKeyboard(bot, h, f.app)
}
