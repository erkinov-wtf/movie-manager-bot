package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Movie struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    int64          `gorm:"type:uint;uniqueIndex:user_api_unique"`
	ApiID     int64          `gorm:"type:uint;uniqueIndex:user_api_unique"`
	Title     string         `gorm:"type:varchar"`
	Runtime   int64          `gorm:"type:uint"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (m *Movie) BeforeCreate(*gorm.DB) (err error) {
	m.ID = uuid.New()
	return nil
}
