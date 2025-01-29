package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	slackx "github.com/artemlive/tars/pkg/slack"
	"github.com/artemlive/tars/pkg/storage"
	"github.com/artemlive/tars/pkg/utils"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	charts "github.com/vicanso/go-charts/v2"
)

// Bot represents the TARS bot.
type Bot struct {
	slackClient slackx.Client
	config      *utils.Config
	ctx         context.Context
	repo        storage.StatsRepository
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
	client.RegisterInteractiveHandler(slack.InteractionTypeShortcut, "draw_stats_for_interval", bot.handleInteractiveEvent)
	client.RegisterInteractiveHandler(slack.InteractionTypeViewSubmission, "pull_stats_for_interval_modal", bot.handleInteractiveEvent)
	client.RegisterInteractiveHandler(slack.InteractionTypeViewSubmission, "draw_stats_for_interval_modal", bot.handleInteractiveEvent)

	bot.config.BuildReactionCache()
	return bot, nil
}

// Run starts the bot's main loop.
func (b *Bot) Run() error {
	log.Println("Starting TARS bot...")
	return b.slackClient.ListenEvents(b.ctx)
}

func (b *Bot) handleInteractiveEvent(eventType slack.InteractionType, callback slack.InteractionCallback) error {
	log.Printf("Interactive Event: %s", eventType)
	if eventType == slack.InteractionTypeShortcut {
		return b.handleInteractiveShortcut(callback)
	} else if eventType == slack.InteractionTypeViewSubmission {
		return b.handleViewSubmission(callback)
	}
	log.Println("Unsupported interactive event type")
	return nil
}

func (b *Bot) handleViewSubmission(callback slack.InteractionCallback) error {
	switch callback.View.CallbackID {
	case "pull_stats_for_interval_modal", "draw_stats_for_interval_modal":
		return b.handlePullStatsForInterval(callback)
	default:
		log.Printf("Unhandled view submission callback: %s", callback.View.CallbackID)
		return nil
	}
}
func (b *Bot) handleInteractiveShortcut(callback slack.InteractionCallback) error {
	switch callback.CallbackID {
	case "pull_stats_for_interval", "draw_stats_for_interval":
		return b.openDatePickerModal(callback.CallbackID, callback.TriggerID)
	default:
		return nil
	}
}

func (b *Bot) openDatePickerModal(modalType, triggerID string) error {
	curDate := time.Now().Format("2006-01-02")

	startDate := slack.NewDatePickerBlockElement("start_date_picker")
	endDate := slack.NewDatePickerBlockElement("end_date_picker")

	startDate.InitialDate = curDate
	endDate.InitialDate = curDate

	modal := slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: fmt.Sprintf("%s_modal", modalType),
		Title: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Stats for Interval📊",
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

	if !b.channelConfigExists(channelID) {
		// TODO: no hardcode actual user-friendly message
		errPost := b.postDM(callback.User.ID, "Sorry, this channel is not configured for stats exporting")
		if errPost != nil {
			log.Printf("Failed to send DM: %v", errPost)
		}
		return fmt.Errorf("channel %s is not configured", channelID)
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

	if callback.View.CallbackID == "pull_stats_for_interval_modal" {
		err = b.processChannelStats(channelID, startDate, endDate)
		if err != nil {
			errPost := b.postEphemeralError(channelID, callback.User.ID, err.Error())
			if errPost != nil {
				log.Printf("Failed to send error message: %v", errPost)
			}
			return err // return the original error
		}
	}

	err = b.GenerateAndSendStatsPieChart(b.ctx, channelID, startDate, endDate, callback.User.ID)
	return err
}

func (b *Bot) channelConfigExists(channelID string) bool {
	for _, channel := range b.config.Channels {
		if channel.ID == channelID {
			return true
		}
	}
	return false
}

func (b *Bot) postDM(userID, message string) error {
	_, _, err := b.slackClient.PostMessageContext(b.ctx, userID, slack.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("failed to send DM: %v", err)
	}
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
	// Fetch messages for the entire range in one go
	messages, err := b.slackClient.FetchMessages(b.ctx, channelID, startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}
	statsProcessor := NewStatsProcessor(b.config)
	beaconReaction := utils.GetControllingReaction(b.config, channelID)
	log.Printf("Fetched %d messages for range %s to %s", len(messages), startDate, endDate)

	// Organize stats by day
	statsByDay := make(map[string]map[string]int) // date -> category -> count
	for _, message := range messages {
		timestampFloat, err := strconv.ParseFloat(message.Timestamp, 64)
		if err != nil {
			log.Printf("Invalid timestamp for message: %s, error: %v", message.Timestamp, err)
			continue
		}
		msgDate := time.Unix(int64(timestampFloat), 0).Format("2006-01-02")
		log.Printf("Message Date: %s", msgDate)
		if !statsProcessor.ShouldProcessMessage(channelID, message, beaconReaction) {
			continue
		}

		reactions, err := b.slackClient.FetchReactions(b.ctx, channelID, message.Timestamp)
		if err != nil {
			log.Printf("Failed to fetch reactions for message %s: %v", message.Timestamp, err)
			continue
		}

		if _, exists := statsByDay[msgDate]; !exists {
			statsByDay[msgDate] = make(map[string]int)
		}
		statsProcessor.UpdateStats(channelID, reactions, statsByDay[msgDate])
	}

	// Save stats for each day
	for day, stats := range statsByDay {
		date, _ := time.Parse("2006-01-02", day)
		if err := b.saveStatsToDB(channelID, date, stats); err != nil {
			log.Printf("Failed to save stats for %s: %v", day, err)
		}
	}
	return nil
}

func (b *Bot) saveStatsToDB(channelID string, date time.Time, stats map[string]int) error {
	return b.repo.SaveStats(channelID, date, stats)
}

func (b *Bot) GenerateAndSendStatsPieChart(ctx context.Context, channelID string, startDate, endDate time.Time, userID string) error {
	// Fetch stats from DB
	stats, err := b.repo.GetAggregatedStats(channelID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch stats: %w", err)
	}

	// Check if stats are empty
	if len(stats) == 0 {
		_, _, err := b.slackClient.PostMessageContext(ctx, userID, slack.MsgOptionText("📉 No stats available for this period.", false))
		return err
	}

	// Aggregate stats by category
	names := []string{}
	values := []float64{}
	for _, stat := range stats {
		names = append(names, stat.Category)
		values = append(values, float64(stat.Count))
	}

	p, err := charts.PieRender(
		values,
		charts.TitleOptionFunc(charts.TitleOption{
			Text:    "Reaction Stats Pie Chart",
			Subtext: fmt.Sprintf("From %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
			Left:    charts.PositionCenter,
		}),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 20,
			Left:   20,
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Theme:  charts.NewTheme(charts.ThemeLight),
			Orient: charts.OrientVertical,
			Data:   names,
			Left:   charts.PositionRight,
		}),
		charts.ThemeOptionFunc(charts.ThemeDark),
		PieSeriesShowLabel(),
	)
	if err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	// Save the chart as an image
	filePath := "/tmp/stats_pie_chart.png"
	buf, err := p.Bytes()
	if err != nil {
		return fmt.Errorf("failed to generate chart bytes: %w", err)
	}
	err = os.WriteFile(filePath, buf, 0644)
	if err != nil {
		return fmt.Errorf("failed to save chart to file: %w", err)
	}

	// Upload to Slack
	err = b.uploadGraphToSlack(userID, filePath, "📊 Reaction Stats Pie Chart")
	if err != nil {
		return fmt.Errorf("failed to upload chart: %w", err)
	}
	os.Remove(filePath)

	return nil
}

func (b *Bot) uploadGraphToSlack(userID string, filePath string, title string) error {
	log.Printf("Channel ID: %s", userID)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	fs, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	// Open a direct message (DM) with the user
	conversation, _, _, err := b.slackClient.OpenConversationContext(b.ctx, &slack.OpenConversationParameters{
		Users: []string{userID}, // Open DM with the user
	})
	if err != nil {
		return fmt.Errorf("failed to open DM with user %s: %w", userID, err)
	}
	// Upload file to Slack
	params := slack.UploadFileV2Parameters{
		Channel:  conversation.ID,
		Reader:   file,
		File:     file.Name(),
		Filename: file.Name(),
		Title:    title,
		FileSize: int(fs.Size()),
	}

	_, err = b.slackClient.UploadFileV2Context(b.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("Successfully uploaded %s to Slack", title)
	return nil
}

func PieSeriesShowLabel() charts.OptionFunc {
	return func(opt *charts.ChartOption) {
		for index := range opt.SeriesList {
			opt.SeriesList[index].Label.Show = true
			opt.SeriesList[index].Label.Formatter = "{b}: {c} ({d})"
		}
	}
}
