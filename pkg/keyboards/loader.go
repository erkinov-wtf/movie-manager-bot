package keyboards

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/interfaces"
	"gopkg.in/telebot.v3"
)

func LoadAllKeyboards(bot *telebot.Bot, handlers interfaces.DefaultInterface) {
	LoadMenuKeyboards(bot)
	LoadTokenRegistrationKeyboard(bot, handlers)
}
