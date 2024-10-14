package tv

type TV struct {
	Id           int64   `json:"id"`
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
