package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Watchlist struct {
	Id        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserId    int64          `gorm:"type:uint;uniqueIndex:user_show_api_unique"`
	ShowApiId int64          `gorm:"type:uint;uniqueIndex:user_show_api_unique"`
	Type      Type           `gorm:"type:varchar"`
	Title     string         `gorm:"type:varchar"`
	Image     string         `gorm:"type:varchar"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`

	User User `gorm:"foreignKey:UserId" json:"user"`
}

func (m *Watchlist) BeforeCreate(*gorm.DB) (err error) {
	m.Id = uuid.New()
	return nil
}

type Type string

const (
	MovieType  Type = "MOVIE"
	TVShowType Type = "TV_SHOW"
	AllType    Type = "ALL"
)
