package agent

import (
	"context"
)

// Client defines the interface for AI model clients
type Client interface {
	Chat(ctx context.Context, messages []Message) (string, Tool, error)
}

// FunctionCall is the type for function call handlers
type FunctionCall func(jsonArgs string) (string, error)
