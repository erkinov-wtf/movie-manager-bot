package movie

type Movie struct {
	ID               int64   `json:"id"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	Runtime          int64   `json:"runtime"`
	Status           string  `json:"status"`
	OriginalLanguage string  `json:"original_language"`
	Adult            bool    `json:"adult"`
	Popularity       float32 `json:"popularity"`
	BackdropPath     string  `json:"backdrop_path"`
	PosterPath       string  `json:"poster_path"`
}
