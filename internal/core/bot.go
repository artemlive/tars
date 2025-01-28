package core

import (
	"context"
	"fmt"
	"log"
	"time"

	slackx "github.com/artemlive/tars/pkg/slack"
	"github.com/artemlive/tars/pkg/storage"
	"github.com/artemlive/tars/pkg/utils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// Bot represents the TARS bot.
type Bot struct {
	slackClient slackx.Client
	config      *utils.Config
	ctx         context.Context
	repo        storage.StatsRepository
	rulesCache  map[string]map[string]string // Channel ID -> Reaction -> Category

}

// NewBot initializes the bot with its dependencies.
func NewBot(ctx context.Context, client slackx.Client, dbRepo storage.StatsRepository, config *utils.Config) (*Bot, error) {
	bot := &Bot{
		slackClient: client,
		config:      config,
		ctx:         ctx,
		repo:        dbRepo,
	}
	// Register event handlers
	client.RegisterEventHandler("app_mention", bot.handleAppMentionEvent)
	client.RegisterEventHandler("reaction_added", bot.handleReactionEvent)

	client.RegisterInteractiveHandler(slack.InteractionTypeShortcut, "pull_stats_for_interval", bot.handleInteractiveEvent)
	client.RegisterInteractiveHandler(slack.InteractionTypeViewSubmission, "pull_stats_for_interval_modal", bot.handleInteractiveEvent)
	bot.loadRulesCache()
	return bot, nil
}

// Run starts the bot's main loop.
func (b *Bot) Run() error {
	log.Println("Starting TARS bot...")
	return b.slackClient.ListenEvents(b.ctx)
}

// Handle interactive events
func (b *Bot) handleInteractiveEvent(eventType slack.InteractionType, callback slack.InteractionCallback) error {
	log.Printf("Interactive Event at client: %+v", eventType)
	switch eventType {
	case slack.InteractionTypeShortcut:
		return b.handleInteractiveShortcut(callback)
	case slack.InteractionTypeViewSubmission:
		return b.handelInteractiveViewSubmission(callback)
	default:
		log.Println("Unsupported interactive event type")
	}
	return nil
}

// Handle interactive view submissions
func (b *Bot) handelInteractiveViewSubmission(callback slack.InteractionCallback) error {
	log.Printf("Interactive View Submission: %+v", callback)

	switch callback.View.CallbackID {
	case "pull_stats_for_interval_modal":
		return b.handlePullStatsForInterval(callback)
	default:
		return nil
	}
}

func (b *Bot) handleInteractiveShortcut(callback slack.InteractionCallback) error {
	switch callback.CallbackID {
	case "pull_stats_for_interval":
		return b.openDatePickerModal(callback.TriggerID)
	default:
		return nil
	}
}

func (b *Bot) openDatePickerModal(triggerID string) error {
	curDate := time.Now().Format("2006-01-02")

	startDate := slack.NewDatePickerBlockElement("start_date_picker")
	endDate := slack.NewDatePickerBlockElement("end_date_picker")

	startDate.InitialDate = curDate
	endDate.InitialDate = curDate

	modal := slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: "pull_stats_for_interval_modal",
		Title: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Pull Stats for Interval📊",
		},
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewInputBlock(
					"channel_picker",
					slack.NewTextBlockObject(slack.PlainTextType, "Select the channel: 🎯", false, false),
					nil,
					slack.NewOptionsSelectBlockElement(
						slack.OptTypeChannels,
						slack.NewTextBlockObject(slack.PlainTextType, "Select a channel", false, false),
						"channel_picker",
					),
				),
				slack.NewInputBlock(
					"start_date",
					slack.NewTextBlockObject(slack.PlainTextType, "Select the start date 📅", false, false),
					slack.NewTextBlockObject(slack.PlainTextType, "start date", false, false),
					startDate,
				),
				slack.NewInputBlock(
					"end_date",
					slack.NewTextBlockObject(slack.PlainTextType, "Select the end date 📅", false, false),
					slack.NewTextBlockObject(slack.PlainTextType, "end date", false, false),
					endDate,
				),
			},
		},
		Submit: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Submit",
		},
	}

	_, err := b.slackClient.OpenViewContext(b.ctx, triggerID, modal)
	if err != nil {
		log.Printf("Error opening modal: %v", err)
	}
	return err
}

// Handle app_mention event
func (b *Bot) handleAppMentionEvent(eventType string, rawEvent interface{}) error {
	event, err := utils.DecodeEvent[slackevents.AppMentionEvent](rawEvent)
	if err != nil {
		return err
	}
	log.Printf("App Mention: %+v", event)
	return nil
}

// Handle reaction_added event
func (b *Bot) handleReactionEvent(eventType string, rawEvent interface{}) error {
	event, err := utils.DecodeEvent[slackevents.ReactionAddedEvent](rawEvent)
	if err != nil {
		return err
	}
	log.Printf("Reaction Added: %+v", event)
	return nil
}

func (b *Bot) handlePullStatsForInterval(callback slack.InteractionCallback) error {
	// Extract channel and date range from the callback
	channelID := callback.View.State.Values["channel_picker"]["channel_picker"].SelectedChannel
	if channelID == "" {
		return fmt.Errorf("channel is required")
	}
	// Extract the selected start date
	startDateString := callback.View.State.Values["start_date"]["start_date_picker"].SelectedDate
	startDate, err := time.Parse("2006-01-02", startDateString)
	if err != nil {
		return fmt.Errorf("invalid start date: %s", startDateString)
	}

	// Extract the selected end date
	endDateString := callback.View.State.Values["end_date"]["end_date_picker"].SelectedDate
	endDate, err := time.Parse("2006-01-02", endDateString)
	if err != nil {
		return fmt.Errorf("invalid end date: %s", endDateString)
	}
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)

	err = b.processChannelStats(channelID, startDate, endDate)
	if err != nil {
		errPost := b.postEphemeralError(channelID, callback.User.ID, err.Error())
		if errPost != nil {
			log.Printf("Failed to send error message: %v", errPost)
		}
		return err // return the original error
	}

	userID := callback.User.ID
	stats, err := b.repo.GetStats(channelID, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching stats: %v", err)
		errPost := b.postEphemeralError(channelID, userID, err.Error())
		if errPost != nil {
			log.Printf("Failed to send error message: %v", err)
		}
		return err
	}

	log.Printf("Stats: %+v", stats)
	return nil
}

func (b *Bot) postEphemeralError(channelID, userID, message string) error {
	errMessage := slack.MsgOptionText(fmt.Sprintf(":x: Sorry, I couldn't process your request: %s", message), false)
	_, err := b.slackClient.PostEphemeralContext(b.ctx, channelID, userID, errMessage)
	if err != nil {
		return fmt.Errorf("failed to send error message: %v", err)
	}
	return nil
}

func (b *Bot) processChannelStats(channelID string, startDate, endDate time.Time) error {
	// Fetch messages for the channel
	messages, err := b.slackClient.FetchMessages(b.ctx, channelID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	log.Println("Messages fetched successfully")

	// Initialize a map to store stats by category
	stats := make(map[string]int)

	// Process messages and update stats
	for _, message := range messages {
		log.Printf("Processing message: %s", message.Text)

		// Fetch reactions for the message
		reactions, err := b.slackClient.FetchReactions(b.ctx, channelID, message.Timestamp)
		if err != nil {
			log.Printf("Failed to fetch reactions for message %s: %v", message.Timestamp, err)
			continue
		}

		// I know that this is a nested loop, but we don't expect a large number of reactions per message
		// so it should be fine
		for _, reaction := range reactions {
			log.Printf("Found reaction: %s (count: %d)", reaction.Name, reaction.Count)

			// Match the reaction to a category
			category, exists := b.getCategoryForReaction(channelID, reaction.Name)
			if exists {
				stats[category] += reaction.Count
				log.Printf("Matched reaction '%s' to category '%s'", reaction.Name, category)
			}
		}
	}

	// Save stats to the database
	for category, count := range stats {
		stat := storage.Stats{
			Channel:   channelID,
			Category:  category,
			Count:     count,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := b.repo.SaveStats(b.ctx, &stat); err != nil {
			return fmt.Errorf("failed to save stats: %w", err)
		}
		log.Printf("Saved stats: %+v", stat)
	}
	log.Printf("Stats for channel %s: %+v", channelID, stats)

	return nil
}

func (b *Bot) getControllingReaction(channelID string) string {
	for _, channel := range b.config.Channels {
		if channel.ID == channelID {
			return channel.BeaconReaction
		}
	}
	return ""
}

func (b *Bot) loadRulesCache() {
	b.rulesCache = make(map[string]map[string]string)
	for _, channel := range b.config.Channels {
		ruleMap := make(map[string]string)
		for _, rule := range channel.Rules {
			ruleMap[rule.Reaction] = rule.Category
		}
		b.rulesCache[channel.Name] = ruleMap
	}
}

func (b *Bot) getCategoryForReaction(channelID, reaction string) (string, bool) {
	channelRules, exists := b.rulesCache[channelID]
	if !exists {
		return "", false
	}
	category, exists := channelRules[reaction]
	return category, exists
}
