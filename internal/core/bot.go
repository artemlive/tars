package core

import (
	"context"
	"log"
	"time"

	slackx "github.com/artemlive/tars/pkg/slack"
	"github.com/artemlive/tars/pkg/utils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// Bot represents the TARS bot.
type Bot struct {
	slackClient slackx.Client
	config      *utils.Config
	ctx         context.Context
}

// NewBot initializes the bot with its dependencies.
func NewBot(ctx context.Context, client slackx.Client, config *utils.Config) (*Bot, error) {
	bot := &Bot{
		slackClient: client,
		config:      config,
		ctx:         ctx,
	}
	// Register event handlers
	client.RegisterEventHandler("app_mention", bot.handleAppMentionEvent)
	client.RegisterEventHandler("reaction_added", bot.handleReactionEvent)

	client.RegisterInteractiveHandler(slack.InteractionTypeShortcut, "pull_stats_for_interval", bot.handleInteractiveEvent)
	client.RegisterInteractiveHandler(slack.InteractionTypeViewSubmission, "pull_stats_for_interval_modal", bot.handleInteractiveEvent)

	return bot, nil
}

// Run starts the bot's main loop.
func (b *Bot) Run() error {
	log.Println("Starting TARS bot...")
	return b.slackClient.ListenEvents(b.ctx)
}

// Handle interactive events
func (b *Bot) handleInteractiveEvent(eventType slack.InteractionType, callback slack.InteractionCallback) error {
	log.Printf("Interactive Event at client: %+v", callback.CallbackID)
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
	switch callback.CallbackID {
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

// Handle pull_stats_for_interval interaction
func (b *Bot) handlePullStatsForInterval(callback slack.InteractionCallback) error {
	log.Printf("Pull Stats For Interval: %+v", callback)
	return nil

}
