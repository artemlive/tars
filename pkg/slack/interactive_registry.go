package slack

import (
	"context"
	"fmt"
	"log"

	"github.com/slack-go/slack"
)

type HandlerMap map[string]InteractiveHandler
type InteractionHandlers map[slack.InteractionType]HandlerMap

type InteractiveHandler func(interType slack.InteractionType, payload slack.InteractionCallback) error

type InteractiveRegistry struct {
	handlers InteractionHandlers
}

func NewInteractiveRegistry() *InteractiveRegistry {
	handlers := make(InteractionHandlers)
	handlers[slack.InteractionTypeShortcut] = make(HandlerMap)

	return &InteractiveRegistry{
		handlers: handlers,
	}
}

func (r *InteractiveRegistry) Register(eventType slack.InteractionType, callBackId string, handler InteractiveHandler) {
	if r.handlers == nil {
		r.handlers = make(InteractionHandlers)
	}
	if r.handlers[eventType] == nil {
		r.handlers[eventType] = make(HandlerMap)
	}
	r.handlers[eventType][callBackId] = handler
}

// Dispatch an interactive event to the appropriate handler
func (r *InteractiveRegistry) Dispatch(ctx context.Context, interType slack.InteractionType, payload slack.InteractionCallback) error {
	var callbackID string

	// Dynamically determine the identifier
	switch interType {
	case slack.InteractionTypeViewSubmission, slack.InteractionTypeViewClosed:
		callbackID = payload.View.CallbackID
	case slack.InteractionTypeBlockActions:
		if len(payload.ActionCallback.BlockActions) > 0 {
			callbackID = payload.ActionCallback.BlockActions[0].ActionID
		}
	default:
		callbackID = payload.CallbackID // Default to the root-level CallbackID
	}

	if callbackID == "" {
		return fmt.Errorf("no valid callback ID found for interaction type: %s", interType)
	}

	log.Printf("Dispatching interactive event: type=%s, callbackID=%s", interType, callbackID)

	// Check if the handler exists and call it
	if handler, exists := r.handlers[interType][callbackID]; exists {
		return handler(interType, payload)
	}
	return fmt.Errorf("no interactive handler registered for callback ID: %s", callbackID)
}
