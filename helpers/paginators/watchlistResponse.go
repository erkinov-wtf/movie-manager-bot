package paginators

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/models"
	"gopkg.in/telebot.v3"
)

func GenerateWatchlistResponse(paginatedWatchlists *[]models.Watchlist, currentPage, maxPage, watchlistCount int, watchlistType string) (string, *telebot.ReplyMarkup) {
	var response string
	for _, w := range *paginatedWatchlists {
		var typeStr string
		if w.Type == models.TVShowType {
			typeStr = "📺 Tv Show"
		} else {
			typeStr = "🎥 Movie"
		}
		response += fmt.Sprintf(
			"🎬 *Title*: %v\n"+
				"📝 *Type*: %v\n"+
				"📅 *Added At*: %v\n\n",
			w.Title,
			typeStr,
			w.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	btn := &telebot.ReplyMarkup{}
	btnRow := telebot.Row{}

	for i, w := range *paginatedWatchlists {
		btnRow = append(btnRow, btn.Data(fmt.Sprintf("%d️⃣", i+1), "", fmt.Sprintf("watchlist|info|%v-%v", w.Type, w.ShowApiId)))
	}

	btn.Inline(
		btnRow,
		btn.Row(
			btn.Data("⏮️ Prev", "", fmt.Sprintf("watchlist|prev|%s-%v", watchlistType, currentPage)),
			btn.Text(fmt.Sprintf("%d | %d • %d", currentPage, maxPage, watchlistCount)),
			btn.Data("Next ⏭️", "", fmt.Sprintf("watchlist|next|%s-%v", watchlistType, currentPage)),
		),
	)

	return response, btn
}
