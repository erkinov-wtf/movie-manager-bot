package models

import (
	"time"
)

type User struct {
	ID         int64     `gorm:"type:uint;primaryKey"`
	FirstName  *string   `gorm:"type:text"`
	LastName   *string   `gorm:"type:text"`
	Username   *string   `gorm:"type:text"`
	Language   *string   `gorm:"type:text"`
	TmdbApiKey *string   `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}
