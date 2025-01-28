package storage

import "time"

// Stats stores reaction stats for a channel
type Stats struct {
	ID        uint   `gorm:"primaryKey"`
	Channel   string `gorm:"index;not null"`
	Category  string `gorm:"not null"`
	Reaction  string `gorm:"not null"`
	Count     int    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
