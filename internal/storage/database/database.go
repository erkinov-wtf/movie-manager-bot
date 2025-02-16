package database

import (
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config"
	"github.com/erkinov-wtf/movie-manager-bot/internal/models"
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
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	log.Print("DB connected successfully")

	err = db.AutoMigrate(&models.Movie{}, &models.TVShows{}, &models.User{}, &models.Watchlist{})
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate models: %v", err))
	}

	log.Print("Models migrated successfully")

	return db
}
