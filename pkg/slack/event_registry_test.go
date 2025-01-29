package slack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewEventsRegistry ensures the registry initializes correctly
func TestNewEventsRegistry(t *testing.T) {
	registry := NewEventsRegistry()

	assert.NotNil(t, registry, "Registry should not be nil")
	assert.NotNil(t, registry.events, "Events map should be initialized")
}

// Mock handler function
func mockEventHandler(eventType string, event interface{}) error {
	return nil
}

// TestEventRegisterHandler verifies handlers are registered correctly
func TestEventRegisterHandler(t *testing.T) {
	registry := NewEventsRegistry()

	registry.Register("test_event", mockEventHandler)

	_, exists := registry.events["test_event"]
	assert.True(t, exists, "Handler should be registered")
}

// TestEventDispatchValidHandler ensures dispatching to a registered handler works
func TestEventDispatchValidHandler(t *testing.T) {
	registry := NewEventsRegistry()
	called := false

	// Mock handler that marks the call as executed
	handler := func(eventType string, event interface{}) error {
		called = true
		return nil
	}

	registry.Register("test_event", handler)

	err := registry.Dispatch(context.Background(), "test_event", nil)

	assert.NoError(t, err, "Expected dispatch to succeed")
	assert.True(t, called, "Handler should have been called")
}

// TestEventDispatchInvalidHandler ensures dispatching an unknown event returns an error
func TestEventDispatchInvalidHandler(t *testing.T) {
	registry := NewEventsRegistry()

	err := registry.Dispatch(context.Background(), "unknown_event", nil)

	assert.Error(t, err, "Expected an error for unregistered event")
	assert.Contains(t, err.Error(), "no handler registered", "Error should indicate missing handler")
}

