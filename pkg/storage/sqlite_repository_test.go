package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *SQLiteStatsRepository {
	t.Helper()

	// Initialize in-memory SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate schema
	err = db.AutoMigrate(&Stats{})
	assert.NoError(t, err)

	return NewSQLiteStatsRepository(db)
}

func TestSaveStats(t *testing.T) {
	repo := setupTestDB(t)

	// Test Data
	channelID := "C123"
	date := time.Date(2025, 01, 29, 0, 0, 0, 0, time.UTC)
	stats := map[string]int{"CI/CD": 5, "Infra Bug": 2}

	// Save Stats
	err := repo.SaveStats(channelID, date, stats)
	assert.NoError(t, err, "Failed to save stats")

	// Fetch and verify
	var storedStats []Stats
	repo.DB.Find(&storedStats)
	assert.Len(t, storedStats, 2)
	assert.Equal(t, 5, storedStats[0].Count)
	assert.Equal(t, "CI/CD", storedStats[0].Category)
}

func TestGetAggregatedStats(t *testing.T) {
	repo := setupTestDB(t)

	// Insert test data
	date := time.Date(2025, 01, 29, 0, 0, 0, 0, time.UTC)
	repo.SaveStats("C123", date, map[string]int{"CI/CD": 3, "Infra Bug": 7})

	// Fetch aggregated stats
	stats, err := repo.GetAggregatedStats("C123", date, date)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)

	// Verify results
	assert.Equal(t, "CI/CD", stats[0].Category)
	assert.Equal(t, 3, stats[0].Count)
	assert.Equal(t, "Infra Bug", stats[1].Category)
	assert.Equal(t, 7, stats[1].Count)
}

func TestGetDailyStats(t *testing.T) {
	repo := setupTestDB(t)

	// Insert data across multiple days
	repo.SaveStats("C123", time.Date(2025, 01, 28, 0, 0, 0, 0, time.UTC), map[string]int{"CI/CD": 1})
	repo.SaveStats("C123", time.Date(2025, 01, 29, 0, 0, 0, 0, time.UTC), map[string]int{"CI/CD": 2})
	repo.SaveStats("C123", time.Date(2025, 01, 30, 0, 0, 0, 0, time.UTC), map[string]int{"Infra Bug": 3})

	// Fetch daily stats
	stats, err := repo.GetDailyStats("C123", time.Date(2025, 01, 28, 0, 0, 0, 0, time.UTC), time.Date(2025, 01, 30, 0, 0, 0, 0, time.UTC))
	assert.NoError(t, err)
	assert.Len(t, stats, 3)

	// Verify each day's stats
	assert.Equal(t, "CI/CD", stats[0].Category)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "CI/CD", stats[1].Category)
	assert.Equal(t, 2, stats[1].Count)
	assert.Equal(t, "Infra Bug", stats[2].Category)
	assert.Equal(t, 3, stats[2].Count)
}
