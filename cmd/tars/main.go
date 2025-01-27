package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/artemlive/tars/internal/core"
	"github.com/artemlive/tars/pkg/slack"
	"github.com/artemlive/tars/pkg/utils"
)

func main() {
	// Load configuration
	configPath := "configs/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	config, err := utils.LoadConfig(configPath)
	log.Printf("config: %+v", config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context that listens for OS signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := slack.NewSlackClient(config.Slack.BotToken, config.Slack.AppToken)

	// Initialize the bot and inject the Slack client
	bot, err := core.NewBot(ctx, client, config)
	if err != nil {
		log.Fatalf("Failed to initialize TARS: %v", err)
	}

	// Run the bot
	if err := bot.Run(); err != nil {
		log.Fatalf("TARS encountered an error: %v", err)
	}

	log.Println("TARS bot has stopped gracefully.")
}
