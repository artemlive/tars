package slack

import (
	"context"
	"fmt"
)

// EventHandler defines a handler function for Slack events.
type EventHandler func(eventType string, event interface{}) error

// EventsRegistry manages registered event handlers.
type EventsRegistry struct {
	events map[string]EventHandler
}

// NewEventsRegistry initializes a new events registry.
func NewEventsRegistry() *EventsRegistry {
	return &EventsRegistry{
		events: make(map[string]EventHandler),
	}
}

// Register adds a new event handler.
func (r *EventsRegistry) Register(eventType string, handler EventHandler) {
	if r.events == nil {
		r.events = make(map[string]EventHandler)
	}
	r.events[eventType] = handler
}

// Dispatch invokes the appropriate event handler.
func (r *EventsRegistry) Dispatch(ctx context.Context, eventType string, event interface{}) error {
	handler, exists := r.events[eventType]
	if !exists {
		return fmt.Errorf("no handler registered for event type: %s", eventType)
	}
	return handler(eventType, event)
}
