package slack

import (
	"context"
	"fmt"
)

// CommandRequest represents the details of a Slack command invocation.
type CommandRequest struct {
	User    string
	Command string
	Text    string
	Channel string
}

// CommandHandler defines the signature for handling commands.
type CommandHandler func(ctx context.Context, cmd CommandRequest) (string, error)

// CommandsRegistry manages registered Slack commands.
type CommandsRegistry struct {
	commands map[string]CommandHandler
}

// NewCommandsRegistry initializes a new commands registry.
func NewCommandsRegistry() *CommandsRegistry {
	return &CommandsRegistry{
		commands: make(map[string]CommandHandler),
	}
}

// Register adds a new command handler.
func (r *CommandsRegistry) Register(name string, handler CommandHandler) {
	if r.commands == nil {
		r.commands = make(map[string]CommandHandler)
	}
	r.commands[name] = handler
}

// Dispatch invokes the appropriate command handler.
func (r *CommandsRegistry) Dispatch(ctx context.Context, name string, cmd CommandRequest) (string, error) {
	handler, exists := r.commands[name]
	if !exists {
		return "", fmt.Errorf("no handler registered for command: %s", name)
	}
	return handler(ctx, cmd)
}
