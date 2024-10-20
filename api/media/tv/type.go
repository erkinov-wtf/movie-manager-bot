package tv

type TV struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	Status       string  `json:"status"`
	Adult        bool    `json:"adult"`
	Popularity   float32 `json:"popularity"`
	Seasons      int64   `json:"number_of_seasons"`
	Episodes     int64   `json:"number_of_episodes"`
	BackdropPath string  `json:"backdrop_path"`
	PosterPath   string  `json:"poster_path"`
}

type Season struct {
	ID           int64     `json:"id"`
	SeasonNumber int64     `json:"season_number"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	Episodes     []Episode `json:"episodes"`
}

type Episode struct {
	ID            int64  `json:"id"`
	AirDate       string `json:"air_date"`
	SeasonNumber  int64  `json:"season_number"`
	EpisodeNumber int64  `json:"episode_number"`
	Runtime       int64  `json:"runtime"`
}
