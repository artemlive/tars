package storage

import (
	"time"
)

type Stats struct {
	ID       uint      `gorm:"primaryKey"`
	Channel  string    `gorm:"not null;uniqueIndex:idx_stats_unique"`
	Category string    `gorm:"not null;uniqueIndex:idx_stats_unique"`
	Reaction string    `gorm:"not null;default:''"`
	Count    int       `gorm:"default:0"`
	Date     time.Time `gorm:"type:DATE;not null;uniqueIndex:idx_stats_unique"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
