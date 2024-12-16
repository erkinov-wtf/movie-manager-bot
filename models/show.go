package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TVShows struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    int64          `gorm:"type:uint;uniqueIndex:user_api_unique"`
	ApiID     int64          `gorm:"type:uint;uniqueIndex:user_api_unique"`
	Name      string         `gorm:"type:varchar"`
	Seasons   int64          `gorm:"type:uint"`
	Episodes  int64          `gorm:"type:uint"`
	Runtime   int64          `gorm:"type:uint"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}

func (tv *TVShows) BeforeCreate(*gorm.DB) (err error) {
	tv.ID = uuid.New()
	return nil
}