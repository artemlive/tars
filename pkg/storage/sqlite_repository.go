package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SQLiteStatsRepository is the SQLite implementation of StatsRepository
type SQLiteStatsRepository struct {
	DB *gorm.DB
}

// NewSQLiteStatsRepository creates a new SQLiteStatsRepository
func NewSQLiteStatsRepository(db *gorm.DB) *SQLiteStatsRepository {
	return &SQLiteStatsRepository{DB: db}
}

// IncrementReaction increments the count for a reaction in a channel
func (r *SQLiteStatsRepository) IncrementReaction(channel, category, reaction string) error {
	var stats Stats
	err := r.DB.Where("channel = ? AND category = ? AND reaction = ?", channel, category, reaction).
		First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		stats = Stats{Channel: channel, Category: category, Reaction: reaction, Count: 1}
		return r.DB.Create(&stats).Error
	}

	if err != nil {
		return err
	}

	stats.Count++
	return r.DB.Save(&stats).Error
}

// GetStats retrieves all stats for a specific channel
func (r *SQLiteStatsRepository) GetStats(channel string, start, end time.Time) ([]Stats, error) {
	var stats []Stats
	err := r.DB.Where("channel = ?", channel).Find(&stats).Error
	return stats, err
}

func (r *SQLiteStatsRepository) SaveStats(ctx context.Context, stat *Stats) error {
	if err := r.DB.WithContext(ctx).Create(stat).Error; err != nil {
		return fmt.Errorf("failed to save stat: %w", err)
	}
	return nil
}
