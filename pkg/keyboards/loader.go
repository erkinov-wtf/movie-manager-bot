package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	appCfg "github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"gopkg.in/telebot.v3"
)

func LoadAllKeyboards(bot *telebot.Bot, handlers interfaces.DefaultInterface, app *appCfg.App) {
	LoadMenuKeyboards(bot, app)
	LoadTokenRegistrationKeyboard(bot, handlers, app)
}
