package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	AppName     string    `yaml:"app_name"`
	Env         string    `yaml:"env"`
	VersionsUrl string    `yaml:"versions_url"`
	General     General   `yaml:"general"`
	Database    Database  `yaml:"database"`
	Endpoints   Endpoints `yaml:"tmdb_endpoints"`
}

type General struct {
	BotToken  string `yaml:"bot_token"`
	SecretKey string `yaml:"secret_key"`
}

type Database struct {
	Host     string `yaml:"host"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
}

type Endpoints struct {
	BaseUrl   string `yaml:"base_url"`
	ImageUrl  string `yaml:"image_url"`
	LoginUrl  string `yaml:"login_url"`
	Resources struct {
		GetMovie string `yaml:"get_movie"`
		GetTV    string `yaml:"get_tv"`
		Search   struct {
			Prefix string `yaml:"prefix"`
			Movie  string `yaml:"movie"`
			TV     string `yaml:"tv"`
		} `yaml:"search"`
	} `yaml:"resources"`
}

func MustLoad() *Config {
	var Cfg Config

	const configPath = "./config/config.yml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &Cfg); err != nil {
		log.Fatalf("cannot read config file: %s", err.Error())
	}

	updateCredentials(&Cfg)

	log.Println("Configurations loaded")

	return &Cfg
}

func updateCredentials(cfg *Config) {
	if env := os.Getenv("ENV"); env != "" {
		cfg.Env = env
	}
	if botToken := os.Getenv("BOT_TOKEN"); botToken != "" {
		cfg.General.BotToken = botToken
	}
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		cfg.General.SecretKey = secretKey
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		cfg.Database.Port = dbPort
	}
}
