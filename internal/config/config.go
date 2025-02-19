package config

import (
	"log"
	"os"
)

type Config struct {
	General   General
	Database  Database
	Endpoints Endpoints
}

type General struct {
	BotToken  string
	SecretKey string
}

type Database struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	Timezone string
}

type Endpoints struct {
	LoginUrl        string
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

func MustLoad() *Config {
	cfg := Config{
		General: General{
			BotToken:  getEnv("BOT_TOKEN"),
			SecretKey: getEnv("SECRET_KEY"),
		},
		Database: Database{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			Name:     getEnv("DB_NAME"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
		},
		Endpoints: Endpoints{
			BaseUrl:     getEnv("BASE_URL"),
			ImageUrl:    getEnv("IMAGE_URL"),
			LoginUrl:    getEnv("LOGIN_URL"),
			SearchMovie: getEnv("SEARCH_MOVIE"),
			GetMovie:    getEnv("GET_MOVIE"),
			SearchTv:    getEnv("SEARCH_TV"),
			GetTv:       getEnv("GET_TV"),
		},
	}

	// Log the loaded configuration for debugging
	log.Printf("Configuration loaded: %+v", cfg)

	return &cfg
}

func getEnv(key string) string {
	return os.Getenv(key)
}
