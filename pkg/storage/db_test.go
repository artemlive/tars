package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitDB tests database initialization for supported and unsupported drivers
func TestInitDB(t *testing.T) {
	// Test SQLite initialization
	db, err := InitDB("sqlite", ":memory:")
	assert.NoError(t, err, "Failed to initialize SQLite DB")
	assert.NotNil(t, db, "Database connection should not be nil")

	// Test unsupported driver
	db, err = InitDB("postgres", "invalid_dsn")
	assert.Error(t, err, "Expected an error for unsupported driver")
	assert.Nil(t, db, "Database should be nil for unsupported driver")
}

// TestMigrate ensures that database migrations work correctly
func TestMigrate(t *testing.T) {
	// Initialize an in-memory SQLite database
	db, err := InitDB("sqlite", ":memory:")
	assert.NoError(t, err, "Failed to initialize SQLite DB")

	// Apply migrations for Stats model
	err = Migrate(db, &Stats{})
	assert.NoError(t, err, "Failed to migrate Stats model")

	// Ensure table exists
	hasTable := db.Migrator().HasTable(&Stats{})
	assert.True(t, hasTable, "Stats table should exist after migration")

	// Ensure columns exist
	hasColumn := db.Migrator().HasColumn(&Stats{}, "Channel")
	assert.True(t, hasColumn, "Stats table should have 'Channel' column")
}
