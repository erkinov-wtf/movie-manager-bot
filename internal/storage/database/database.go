package database

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	models2 "github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func MustLoadDb(config *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	log.Print("DB connected successfully")

	err = db.AutoMigrate(&models2.Movie{}, &models2.TVShows{}, &models2.User{}, &models2.Watchlist{})
	if err != nil {
		panic(err)
	}

	log.Print("Models migrated successfully")

	return db
}
