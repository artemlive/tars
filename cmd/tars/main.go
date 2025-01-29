package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/artemlive/tars/internal/core"
	"github.com/artemlive/tars/pkg/slack"
	"github.com/artemlive/tars/pkg/storage"
	"github.com/artemlive/tars/pkg/utils"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to the configuration file")
	flag.Parse()

	config, err := utils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a context that listens for OS signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client := slack.NewSlackClient(config.Slack.BotToken, config.Slack.AppToken)

	repo, err := storage.NewRepository(config.Database.Driver, config.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	// Initialize the bot and inject the Slack client and repository
	bot, err := core.NewBot(ctx, client, repo, config)
	if err != nil {
		log.Fatalf("Failed to initialize TARS: %v", err)
	}

	// Run the bot
	if err := bot.Run(); err != nil {
		log.Fatalf("TARS encountered an error: %v", err)
	}

	log.Println("TARS bot has stopped gracefully.")
}
