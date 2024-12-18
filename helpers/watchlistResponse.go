package helpers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"movie-manager-bot/models"
)

func GenerateWatchlistResponse(paginatedWatchlists *[]models.Watchlist, currentPage, maxPage, watchlistCount int) (string, *telebot.ReplyMarkup) {
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
		btnRow = append(btnRow, btn.Data(fmt.Sprintf("%d️⃣", i+1), "", fmt.Sprintf("watchlist|info|%v", w.ID)))
	}

	btn.Inline(
		btnRow,
		btn.Row(
			btn.Data("⏮️ Prev", "", "watchlist|prev|"),
			btn.Text(fmt.Sprintf("Page %d | %d • %d shows", currentPage, maxPage, watchlistCount)),
			btn.Data("Next ⏭️", "", "watchlist|next|"),
		),
	)

	return response, btn
}
