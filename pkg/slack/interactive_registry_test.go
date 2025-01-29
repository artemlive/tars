package slack

import (
	"context"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

// TestNewInteractiveRegistry ensures the registry initializes correctly
func TestNewInteractiveRegistry(t *testing.T) {
	registry := NewInteractiveRegistry()

	assert.NotNil(t, registry, "Registry should not be nil")
	assert.NotNil(t, registry.handlers, "Handlers map should be initialized")
	assert.NotNil(t, registry.handlers[slack.InteractionTypeShortcut], "Shortcut handlers should be initialized")
}

// Mock handler function
func mockHandler(interType slack.InteractionType, payload slack.InteractionCallback) error {
	return nil
}

// TestRegisterHandler verifies handlers are registered correctly
func TestRegisterHandler(t *testing.T) {
	registry := NewInteractiveRegistry()

	registry.Register(slack.InteractionTypeShortcut, "test_callback", mockHandler)

	_, exists := registry.handlers[slack.InteractionTypeShortcut]["test_callback"]
	assert.True(t, exists, "Handler should be registered")
}

// TestDispatchValidHandler ensures dispatching to a registered handler works
func TestDispatchValidHandler(t *testing.T) {
	registry := NewInteractiveRegistry()
	called := false

	// Mock handler that marks the call as executed
	handler := func(interType slack.InteractionType, payload slack.InteractionCallback) error {
		called = true
		return nil
	}

	registry.Register(slack.InteractionTypeShortcut, "test_callback", handler)

	payload := slack.InteractionCallback{CallbackID: "test_callback"}
	err := registry.Dispatch(context.Background(), slack.InteractionTypeShortcut, payload)

	assert.NoError(t, err, "Expected dispatch to succeed")
	assert.True(t, called, "Handler should have been called")
}

// TestDispatchInvalidHandler ensures dispatching an unknown event returns an error
func TestDispatchInvalidHandler(t *testing.T) {
	registry := NewInteractiveRegistry()

	payload := slack.InteractionCallback{CallbackID: "unknown_callback"}
	err := registry.Dispatch(context.Background(), slack.InteractionTypeShortcut, payload)

	assert.Error(t, err, "Expected an error for unregistered callback")
	assert.Contains(t, err.Error(), "no interactive handler registered", "Error should indicate missing handler")
}

// TestDispatchViewSubmission ensures correct callback ID is extracted from ViewSubmission
func TestDispatchViewSubmission(t *testing.T) {
	registry := NewInteractiveRegistry()
	called := false

	handler := func(interType slack.InteractionType, payload slack.InteractionCallback) error {
		called = true
		return nil
	}

	registry.Register(slack.InteractionTypeViewSubmission, "view_callback", handler)

	payload := slack.InteractionCallback{
		View: slack.View{CallbackID: "view_callback"},
	}
	err := registry.Dispatch(context.Background(), slack.InteractionTypeViewSubmission, payload)

	assert.NoError(t, err, "Expected dispatch to succeed")
	assert.True(t, called, "Handler should have been called")
}

// TestDispatchBlockActions ensures correct callback ID is extracted from BlockActions
func TestDispatchBlockActions(t *testing.T) {
	registry := NewInteractiveRegistry()
	called := false

	handler := func(interType slack.InteractionType, payload slack.InteractionCallback) error {
		called = true
		return nil
	}

	registry.Register(slack.InteractionTypeBlockActions, "block_action", handler)

	payload := slack.InteractionCallback{
		ActionCallback: slack.ActionCallbacks{
			BlockActions: []*slack.BlockAction{
				{ActionID: "block_action"},
			},
		},
	}

	err := registry.Dispatch(context.Background(), slack.InteractionTypeBlockActions, payload)

	assert.NoError(t, err, "Expected dispatch to succeed")
	assert.True(t, called, "Handler should have been called")
}
