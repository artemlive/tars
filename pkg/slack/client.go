package slack

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Client interface {
	RegisterEventHandler(eventType string, handler EventHandler)
	RegisterCommandHandler(command string, handler CommandHandler)
	RegisterInteractiveHandler(interactionType slack.InteractionType, callbackID string, handler InteractiveHandler)
	ListenEvents(ctx context.Context) error
	FetchMessages(ctx context.Context, channelID string, from, to time.Time) ([]slack.Message, error)
	OpenViewContext(ctx context.Context, triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error)
	PostEphemeralContext(ctx context.Context, channel, user string, options ...slack.MsgOption) (string, error)
	FetchReactions(ctx context.Context, channelID, timestamp string) ([]slack.ItemReaction, error)
	PostMessageContext(ctx context.Context, channel string, options ...slack.MsgOption) (string, string, error)
}

// Client wraps the Slack API and socket mode client.
type SlackClient struct {
	api                *slack.Client
	socket             *socketmode.Client
	socketHandler      *socketmode.SocketmodeHandler
	eventsRegistry     *EventsRegistry
	commandsRegistry   *CommandsRegistry
	interaciveRegistry *InteractiveRegistry
	context            context.Context
}

// NewSlackClient initializes a new Slack client for socket mode.
func NewSlackClient(botToken, appToken string) *SlackClient {
	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(appToken),
	)
	socket := socketmode.New(
		api,
		socketmode.OptionDebug(true),
	)
	client := &SlackClient{
		api:                api,
		socket:             socket,
		socketHandler:      socketmode.NewSocketmodeHandler(socket),
		eventsRegistry:     NewEventsRegistry(),
		commandsRegistry:   NewCommandsRegistry(),
		interaciveRegistry: NewInteractiveRegistry(),
	}
	client.registerConnectionHandlers()
	return client
}

// registerConnectionHandlers sets up middleware for connection events.
func (c *SlackClient) registerConnectionHandlers() {
	c.socketHandler.Handle(socketmode.EventTypeHello, func(e *socketmode.Event, client *socketmode.Client) {
		log.Println("Hello event received.")
	})
	c.socketHandler.Handle(socketmode.EventTypeConnecting, func(e *socketmode.Event, client *socketmode.Client) {
		log.Println("Connecting to Slack...")
	})

	c.socketHandler.Handle(socketmode.EventTypeConnectionError, func(e *socketmode.Event, client *socketmode.Client) {
		log.Printf("Connection error: %v\n", e)
	})

	c.socketHandler.Handle(socketmode.EventTypeConnected, func(e *socketmode.Event, client *socketmode.Client) {
		log.Println("Connected to Slack.")
	})
}

// RegisterEventHandler registers a handler for specific event types.
func (s *SlackClient) RegisterEventHandler(eventType string, handler EventHandler) {
	log.Printf("Registered handler for event type: %s", eventType)
	s.eventsRegistry.Register(eventType, handler)
}

// RegisterCommandHandler registers a handler for a specific command.
func (s *SlackClient) RegisterCommandHandler(command string, handler CommandHandler) {
	log.Printf("Registered handler for command: %s", command)
	s.commandsRegistry.Register(command, handler)
}

// RegisterInteractiveHandler registers a handler for interactive payloads.
func (s *SlackClient) RegisterInteractiveHandler(interactionType slack.InteractionType, callbackID string, handler InteractiveHandler) {
	log.Printf("Registered handler for interaction type: %s", interactionType)
	s.interaciveRegistry.Register(interactionType, callbackID, handler)
}

// ListenEvents starts listening for events via Socket Mode.
func (s *SlackClient) ListenEvents(ctx context.Context) error {
	s.dispatchEventsAPIHandlers()
	s.dispatchInteractiveHandlers()
	s.dispatchEventsCommandHandlers()

	go func() {
		<-ctx.Done()
		log.Println("Context canceled. Stopping the handler...")
	}()
	return s.socketHandler.RunEventLoopContext(ctx)
}

// registerEventsAPIHandlers sets up middleware for handling Events API payloads.
func (s *SlackClient) dispatchEventsAPIHandlers() {
	s.socketHandler.Handle(socketmode.EventTypeEventsAPI, func(e *socketmode.Event, client *socketmode.Client) {
		// Acknowledge the event
		client.Ack(*e.Request)

		eventData, ok := e.Data.(slackevents.EventsAPIEvent)
		if !ok {
			log.Printf("Ignored unsupported Events API event: %+v", e)
			return
		}

		// Dispatch to the specific event handler
		log.Printf("Received event: %+v\n", eventData.InnerEvent)
		if err := s.eventsRegistry.Dispatch(context.Background(), eventData.InnerEvent.Type, eventData.InnerEvent.Data); err != nil {
			log.Printf("Error handling event %s: %v", eventData.InnerEvent.Type, err)
		}
	})
}

// registerEventsCommandHandlers sets up middleware for handling slash commands.
func (s *SlackClient) dispatchEventsCommandHandlers() {
	s.socketHandler.Handle(socketmode.EventTypeSlashCommand, func(e *socketmode.Event, client *socketmode.Client) {
		log.Printf("Slash command event received: %+v", e)
	})
}

// registerInteractiveHandlers sets up middleware for handling interactive payloads.
func (s *SlackClient) dispatchInteractiveHandlers() {
	s.socketHandler.Handle(socketmode.EventTypeInteractive, func(e *socketmode.Event, client *socketmode.Client) {
		client.Ack(*e.Request)

		callback, ok := e.Data.(slack.InteractionCallback)
		if !ok {
			log.Printf("Ignored unsupported interactive event: %+v", e)
			return
		}
		if err := s.interaciveRegistry.Dispatch(context.Background(), callback.Type, callback); err != nil {
			log.Printf("Error handling interactive event %s: %v", callback.Type, err)
		}
	})
}

func (s *SlackClient) FetchMessages(ctx context.Context, channelID string, from, to time.Time) ([]slack.Message, error) {
	log.Printf("Fetching messages from %s between %s and %s", channelID, from, to)
	var allMessages []slack.Message
	cursor := ""
	for {
		history, err := s.api.GetConversationHistoryContext(
			ctx,
			&slack.GetConversationHistoryParameters{
				ChannelID: channelID,
				Oldest:    fmt.Sprintf("%f", float64(from.Unix())),
				Latest:    fmt.Sprintf("%f", float64(to.Unix())),
				Cursor:    cursor,
			},
		)
		if err != nil {
			return nil, err
		}
		allMessages = append(allMessages, history.Messages...)
		if history.ResponseMetaData.NextCursor == "" {
			break
		}
		cursor = history.ResponseMetaData.NextCursor

	}
	return allMessages, nil
}

func (s *SlackClient) OpenViewContext(ctx context.Context, triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error) {
	return s.api.OpenViewContext(ctx, triggerID, view)
}

func (s *SlackClient) PostEphemeralContext(ctx context.Context, channel, user string, options ...slack.MsgOption) (string, error) {
	return s.api.PostEphemeralContext(ctx, channel, user, options...)
}

func (s *SlackClient) FetchReactions(ctx context.Context, channelID, timestamp string) ([]slack.ItemReaction, error) {
	reactions, err := s.api.GetReactionsContext(ctx, slack.ItemRef{
		Channel:   channelID,
		Timestamp: timestamp,
	}, slack.GetReactionsParameters{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}
	return reactions, nil
}

func (s *SlackClient) PostMessageContext(ctx context.Context, channel string, options ...slack.MsgOption) (string, string, error) {
	return s.api.PostMessageContext(ctx, channel, options...)
}
