package paginators

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"gopkg.in/telebot.v3"
)

func GenerateTVResponse(paginatedTV []tv.TV, currentPage, maxPage, tvCount int) (string, *telebot.ReplyMarkup) {
	var response string
	for _, el := range paginatedTV {
		response += fmt.Sprintf(
			"📺 *Name*: %v\n"+
				"📝 *Overview*: %v\n"+
				"📜 *Status*: %v\n"+
				"🔞 *Is Adult*: %v\n"+
				"🔥 *Popularity*: %v\n"+
				"🎥 *Seasons*: %v\n"+
				"#️⃣ *Episodes*: %v\n\n",
			el.Name,
			el.Overview,
			el.Status,
			el.Adult,
			el.Popularity,
			el.Seasons,
			el.Episodes,
		)
	}

	btn := &telebot.ReplyMarkup{}
	btnRow := telebot.Row{}

	for i, mov := range paginatedTV {
		btnRow = append(btnRow, btn.Data(fmt.Sprintf("%d️⃣", i+1), "", fmt.Sprintf("tv|tv|%v", mov.Id)))
	}

	btn.Inline(
		btnRow,
		btn.Row(
			btn.Data("⏮️ Prev", "", "tv|prev|"),
			btn.Text(fmt.Sprintf("Page %d | %d • %d shows", currentPage, maxPage, tvCount)),
			btn.Data("Next ⏭️", "", "tv|next|"),
		),
	)

	return response, btn
}
