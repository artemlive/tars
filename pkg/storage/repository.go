package storage

import (
	"context"
	"fmt"
	"log"
	"time"
)

// StatsRepository defines methods for interacting with stats storage
type StatsRepository interface {
	IncrementReaction(channel, category, reaction string) error
	GetStats(channel string, start, end time.Time) ([]Stats, error)
	SaveStats(ctx context.Context, stat *Stats) error
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
