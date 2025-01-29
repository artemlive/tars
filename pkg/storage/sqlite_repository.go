package storage

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StatsQuery defines query parameters for fetching stats
type StatsQuery struct {
	Channel string
	Start   time.Time
	End     time.Time
	GroupBy []string // Defines what fields to group by (category, date, etc.)
}

// SQLiteStatsRepository is the SQLite implementation of StatsRepository
type SQLiteStatsRepository struct {
	DB *gorm.DB
}

// NewSQLiteStatsRepository creates a new SQLiteStatsRepository
func NewSQLiteStatsRepository(db *gorm.DB) *SQLiteStatsRepository {
	return &SQLiteStatsRepository{DB: db}
}

func (r *SQLiteStatsRepository) GetAggregatedStats(channel string, start, end time.Time) ([]Stats, error) {
	query := StatsQuery{
		Channel: channel,
		Start:   start,
		End:     end,
		GroupBy: []string{"category"},
	}
	return r.getStats(query)
}

func (r *SQLiteStatsRepository) getStats(query StatsQuery) ([]Stats, error) {
	var results []Stats
	db := r.DB

	// Base selection
	db = db.Select("SUM(count) as count")

	// Append group fields dynamically
	if len(query.GroupBy) > 0 {
		groupClause := strings.Join(query.GroupBy, ", ")
		db = db.Select(groupClause + ", SUM(count) as count").Group(groupClause)
	}

	// Apply filters
	db = db.Where("channel = ? AND date BETWEEN ? AND ?", query.Channel, query.Start, query.End)

	// Execute query
	err := db.Find(&results).Error
	if err != nil {
		log.Printf("❌ Failed to fetch stats: %v", err)
	}
	return results, err
}

func (r *SQLiteStatsRepository) GetDailyStats(channel string, start, end time.Time) ([]Stats, error) {
	query := StatsQuery{
		Channel: channel,
		Start:   start,
		End:     end,
		GroupBy: []string{"date", "category"},
	}
	return r.getStats(query)
}

func (r *SQLiteStatsRepository) SaveStats(channelID string, date time.Time, stats map[string]int) error {
	tx := r.DB.Begin()

	for category, count := range stats {
		stat := Stats{
			Channel:   channelID,
			Category:  category,
			Count:     count,
			Date:      date,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "channel"}, {Name: "category"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"count", "updated_at"}),
		}).Create(&stat).Error

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save stats: %w", err)
		}
	}

	return tx.Commit().Error
}
