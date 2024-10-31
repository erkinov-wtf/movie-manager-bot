package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TVShows struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ApiID     int64          `gorm:"type:uint"`
	Name      string         `gorm:"type:varchar"`
	Seasons   int64          `gorm:"type:uint"`
	Episodes  int64          `gorm:"type:uint"`
	Runtime   int64          `gorm:"type:uint"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (tv *TVShows) BeforeCreate(*gorm.DB) (err error) {
	tv.ID = uuid.New()
	return nil
}
