package helpers

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"movie-manager-bot/api/media/tv"
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
		btnRow = append(btnRow, btn.Data(fmt.Sprintf("%d️⃣", i+1), "", fmt.Sprintf("tv|tv|%v", mov.ID)))
	}

	btn.Inline(
		btnRow,
		btn.Row(
			btn.Data("⏮️ Prev", "", "tv|prev|"),
			btn.Text(fmt.Sprintf("%v | %v | %v", currentPage, maxPage, tvCount)),
			btn.Data("Next ⏭️", "", "tv|next|"),
		),
	)

	return response, btn
}
