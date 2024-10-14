package config

import (
	"log"
	"os"
)

var Cfg Config

type Config struct {
	General   General
	Database  Database
	Endpoints Endpoints
}

type General struct {
	BotToken string
	ApiKey   string
}

type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
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
		Database: Database{},
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
