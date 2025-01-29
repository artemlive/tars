package slack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewCommandsRegistry ensures the registry initializes correctly
func TestNewCommandsRegistry(t *testing.T) {
	registry := NewCommandsRegistry()

	assert.NotNil(t, registry, "Registry should not be nil")
	assert.NotNil(t, registry.commands, "Commands map should be initialized")
}

// Mock command handler function
func mockCommandHandler(ctx context.Context, cmd CommandRequest) (string, error) {
	return "Command executed", nil
}

// TestRegisterCommandHandler verifies command handlers are registered correctly
func TestRegisterCommandHandler(t *testing.T) {
	registry := NewCommandsRegistry()

	registry.Register("test_command", mockCommandHandler)

	_, exists := registry.commands["test_command"]
	assert.True(t, exists, "Command handler should be registered")
}

// TestDispatchValidCommand ensures dispatching to a registered command works
func TestDispatchValidCommand(t *testing.T) {
	registry := NewCommandsRegistry()

	// Mock handler that returns a static response
	handler := func(ctx context.Context, cmd CommandRequest) (string, error) {
		return "Command executed successfully", nil
	}

	registry.Register("test_command", handler)

	response, err := registry.Dispatch(context.Background(), "test_command", CommandRequest{})

	assert.NoError(t, err, "Expected dispatch to succeed")
	assert.Equal(t, "Command executed successfully", response, "Handler response should match expected output")
}

// TestDispatchInvalidCommand ensures dispatching an unknown command returns an error
func TestDispatchInvalidCommand(t *testing.T) {
	registry := NewCommandsRegistry()

	response, err := registry.Dispatch(context.Background(), "unknown_command", CommandRequest{})

	assert.Error(t, err, "Expected an error for unknown command")
	assert.Equal(t, "", response, "Response should be empty for unknown command")
	assert.Contains(t, err.Error(), "no handler registered for command", "Error message should indicate missing handler")
}
