package tv

type TV struct {
	Id           int64   `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	Status       string  `json:"status"`
	Adult        bool    `json:"adult"`
	Popularity   float32 `json:"popularity"`
	Seasons      int32   `json:"number_of_seasons"`
	Episodes     int32   `json:"number_of_episodes"`
	BackdropPath string  `json:"backdrop_path"`
	PosterPath   string  `json:"poster_path"`
}

type Season struct {
	Id           int64     `json:"id"`
	SeasonNumber int32     `json:"season_number"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	Episodes     []Episode `json:"episodes"`
}

type Episode struct {
	Id            int64  `json:"id"`
	AirDate       string `json:"air_date"`
	SeasonNumber  int32  `json:"season_number"`
	EpisodeNumber int32  `json:"episode_number"`
	Runtime       int32  `json:"runtime"`
}
