package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRepository tests initializing different database drivers
func TestNewRepository(t *testing.T) {
	// Test SQLite initialization
	repo, err := NewRepository("sqlite", ":memory:")
	assert.NoError(t, err, "Failed to initialize SQLite repository")
	assert.NotNil(t, repo, "Repository should not be nil")

	// Ensure it implements the interface
	_, ok := repo.(StatsRepository)
	assert.True(t, ok, "Repository should implement StatsRepository")
}
