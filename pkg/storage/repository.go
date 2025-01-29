package storage

import (
	"fmt"
	"log"
	"time"
)

// StatsRepository defines methods for interacting with stats storage
type StatsRepository interface {
	SaveStats(channelID string, date time.Time, stats map[string]int) error
	GetAggregatedStats(channel string, start, end time.Time) ([]Stats, error)
	GetDailyStats(channel string, start, end time.Time) ([]Stats, error)
}

// NewRepository initializes the database and returns a StatsRepository.
func NewRepository(driver, dsn string) (StatsRepository, error) {
	db, err := InitDB(driver, dsn)
	if err != nil {
		return nil, err
	}

	err = Migrate(db, &Stats{})
	if err != nil {
		return nil, err
	}

	log.Printf("Database initialized with driver: %s", driver)

	switch driver {
	case "sqlite":
		return NewSQLiteStatsRepository(db), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}
