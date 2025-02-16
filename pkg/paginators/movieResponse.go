package paginators

import (
	"fmt"
	movieType "github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/movie"
	"gopkg.in/telebot.v3"
)

func GenerateMovieResponse(paginatedMovies []movieType.Movie, currentPage, maxPage, movieCount int) (string, *telebot.ReplyMarkup) {
	var response string
	for _, mov := range paginatedMovies {
		response += fmt.Sprintf(
			"🎬 *Title*: %v\n"+
				"📝 *Overview*: %v\n"+
				"📅 *Release Date*: %s\n"+
				"⏳ *Runtime*: %v minutes\n"+
				"🔞 *Is Adult*: %v\n"+
				"🔥 *Popularity*: %v\n\n",
			mov.Title,
			mov.Overview,
			mov.ReleaseDate,
			mov.Runtime,
			mov.Adult,
			mov.Popularity,
		)
	}

	btn := &telebot.ReplyMarkup{}
	btnRow := telebot.Row{}

	for i, mov := range paginatedMovies {
		btnRow = append(btnRow, btn.Data(fmt.Sprintf("%d️⃣", i+1), "", fmt.Sprintf("movie|movie|%v", mov.ID)))
	}

	btn.Inline(
		btnRow,
		btn.Row(
			btn.Data("⏮️ Prev", "", "movie|prev|"),
			btn.Text(fmt.Sprintf("Page %d | %d • %d movies", currentPage, maxPage, movieCount)),
			btn.Data("Next ⏭️", "", "movie|next|"),
		),
	)

	return response, btn
}
