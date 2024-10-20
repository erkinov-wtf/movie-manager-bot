package config

import (
	"log"
	"os"
)

var Cfg Config

type Config struct {
	General   General
	Firebase  Firebase
	Endpoints Endpoints
}

type General struct {
	BotToken string
	ApiKey   string
}

type Firebase struct {
	FirebaseCredentials       string
	FirebaseMoviesCollection  string
	FirebaseTvShowsCollection string
}

type Endpoints struct {
	Login           string
	BaseUrl         string
	ImageUrl        string
	SearchMovie     string
	SearchTv        string
	GetMovie        string
	GetTv           string
	Groups          string
	GroupData       string
	Discounts       string
	BranchFromGroup string
	AllStudents     string
}

func MustLoad() {
	Cfg = Config{
		General: General{
			BotToken: getEnv("BOT_TOKEN"),
			ApiKey:   getEnv("API_KEY"),
		},
		Firebase: Firebase{
			FirebaseMoviesCollection:  getEnv("FIREBASE_MOVIES_COLLECTION"),
			FirebaseTvShowsCollection: getEnv("FIREBASE_TVSHOWS_COLLECTION"),
		},
		Endpoints: Endpoints{
			BaseUrl:     getEnv("BASE_URL"),
			ImageUrl:    getEnv("IMAGE_URL"),
			Login:       getEnv("LOGIN"),
			SearchMovie: getEnv("SEARCH_MOVIE"),
			GetMovie:    getEnv("GET_MOVIE"),
			SearchTv:    getEnv("SEARCH_TV"),
			GetTv:       getEnv("GET_TV"),
		},
	}

	// Log the loaded configuration for debugging
	log.Printf("Configuration loaded: %+v", Cfg)
}

func getEnv(key string) string {
	return os.Getenv(key)
}
