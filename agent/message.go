package agent

import (
	"encoding/json"
)

// Message represents a chat message for the model
type Message struct {
	Role          string          `json:"role"`
	Content       *string         `json:"content,omitempty"`
	ToolCalls     []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID    string          `json:"tool_call_id,omitempty"`
	Name          string          `json:"name,omitempty"`
}

// ToolCall represents a single tool call from the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents the function details in a tool call
type ToolFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Function represents the function type for tool execution
type Function struct {
	Name      string
	Arguments json.RawMessage
}

// Tool represents a tool call result
type Tool struct {
	ID       string
	Type     string
	Function Function
}
