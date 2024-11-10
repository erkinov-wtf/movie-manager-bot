package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/config"
	"movie-manager-bot/models"
)

var DB *gorm.DB

func DBConnect() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Cfg.Database.Host,
		config.Cfg.Database.User,
		config.Cfg.Database.Password,
		config.Cfg.Database.Name,
		config.Cfg.Database.Port,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	log.Print("DB connected successfully")

	err = DB.AutoMigrate(&models.Movie{}, &models.TVShows{}, &models.User{})
	if err != nil {
		panic(err)
	}

	log.Print("Models migrated successfully")
}
