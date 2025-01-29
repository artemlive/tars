package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_UnmarshalFailure(t *testing.T) {
	// Create a YAML config with an incorrect data type
	configContent := `
slack:
  app_token: "xapp-test-token"
  bot_token: "xoxb-test-token"
  signing_secret: "some-secret"
bot:
  log_level:  
    - "info"
    - "error"
`

	tempFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(configContent))
	assert.NoError(t, err)
	assert.NoError(t, tempFile.Close())

	_, err = LoadConfig(tempFile.Name())

	assert.Error(t, err, "Expected error due to type mismatch in config file")
}

func TestLoadConfig_WithValidFile(t *testing.T) {
	configContent := `
slack:
  app_token: "xapp-test-token"
  bot_token: "xoxb-test-token"
  signing_secret: "test-secret"
bot:
  log_level: "debug"
db:
  driver: "sqlite"
  dsn: "test.db"
channels:
  - name: "Test Channel"
    id: "C123456"
    beacon_reaction: ":beacon:"
    rules:
      - reaction: ":thumbsup:"
        category: "approval"
      - reaction: ":bug:"
        category: "issue"
`

	tempFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(configContent))
	assert.NoError(t, err)
	assert.NoError(t, tempFile.Close())

	// Load the configuration
	config, err := LoadConfig(tempFile.Name())
	assert.NoError(t, err)

	// Validate the loaded configuration
	assert.Equal(t, "xapp-test-token", config.Slack.AppToken)
	assert.Equal(t, "xoxb-test-token", config.Slack.BotToken)
	assert.Equal(t, "test-secret", config.Slack.SigningSecret)
	assert.Equal(t, "debug", config.Bot.LogLevel)
	assert.Equal(t, "sqlite", config.Database.Driver)
	assert.Equal(t, "test.db", config.Database.DSN)

	assert.Len(t, config.Channels, 1)
	assert.Equal(t, "Test Channel", config.Channels[0].Name)
	assert.Equal(t, "C123456", config.Channels[0].ID)
	assert.Equal(t, ":beacon:", config.Channels[0].BeaconReaction)
	assert.Len(t, config.Channels[0].Rules, 2)
	assert.Equal(t, ":thumbsup:", config.Channels[0].Rules[0].Reaction)
	assert.Equal(t, "approval", config.Channels[0].Rules[0].Category)
	assert.Equal(t, ":bug:", config.Channels[0].Rules[1].Reaction)
	assert.Equal(t, "issue", config.Channels[0].Rules[1].Category)
}

func TestLoadConfig_WithMissingFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
}

func TestBuildReactionCache(t *testing.T) {
	config := &Config{
		Channels: []ChannelConfig{
			{
				ID:   "C123456",
				Name: "Test Channel",
				Rules: []RuleConfig{
					{Reaction: ":thumbsup:", Category: "approval"},
					{Reaction: ":bug:", Category: "issue"},
				},
			},
			{
				ID:   "C654321",
				Name: "Another Channel",
				Rules: []RuleConfig{
					{Reaction: ":fire:", Category: "alert"},
				},
			},
		},
	}

	// Build the reaction cache
	config.BuildReactionCache()

	// Validate the reaction cache
	assert.NotNil(t, config.ReactionCache)
	assert.Len(t, config.ReactionCache, 2)

	assert.Equal(t, "approval", config.ReactionCache["C123456"][":thumbsup:"])
	assert.Equal(t, "issue", config.ReactionCache["C123456"][":bug:"])
	assert.Equal(t, "alert", config.ReactionCache["C654321"][":fire:"])
}
