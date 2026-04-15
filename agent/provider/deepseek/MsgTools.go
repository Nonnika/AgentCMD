// This package is used for DeepSeek Model
// Include create chat request, response, and build request, response from config, messages

package deepseek

import (
	"encoding/json"
	"fmt"

	"github.com/Nonnika/agentcmd/agent"
	"github.com/Nonnika/agentcmd/agent/tools"
)

// Response message types
type RespMessage struct {
	Role             string     `json:"role"`
	Content          *string    `json:"content,omitempty"`
	ReasoningContent *string    `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Id       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Request/Response structs
type RespMsg struct {
	Id                string    `json:"id"`
	Object            string    `json:"object"`
	Created           int64     `json:"created"`
	Model             string    `json:"model"`
	Choices           []Choices `json:"choices"`
	Usage             Usage     `json:"usage"`
	SystemFingerprint string    `json:"system_fingerprint"`
}

type ReqMsg struct {
	Model     string            `json:"model"`
	Messages  []interface{}     `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
	Tools     []json.RawMessage `json:"tools"`
}

type Choices struct {
	Index        int         `json:"index"`
	Message      RespMessage `json:"message"`
	Logprobs     string      `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	TotalTokens           int                `json:"total_tokens"`
	PromptTokens          int                `json:"prompt_tokens"`
	CompletionTokens      int                `json:"completion_tokens"`
	PromptTokenDetails    PromptTokenDetails `json:"prompt_token_details"`
	PromptCacheHitTokens  int                `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int                `json:"prompt_cache_miss_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens int64 `json:"cached_tokens"`
}

// BuildReqMsg builds the request message for DeepSeek API
func BuildReqMsg(cfg *Config, messages []agent.Message) ReqMsg {
	toolsRaw := make([]json.RawMessage, len(tools.ToolsIndex))
	for i, tool := range tools.ToolsIndex {
		toolsRaw[i] = json.RawMessage(tool)
	}

	// Convert messages to DeepSeek format
	deepseekMessages := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		dsMsg := map[string]interface{}{
			"role": msg.Role,
		}

		if msg.Content != nil {
			dsMsg["content"] = msg.Content
		}

		// Handle tool_calls in assistant messages
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				toolCalls[i] = map[string]interface{}{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]interface{}{
						"name":      tc.Function.Name,
						"arguments": string(tc.Function.Arguments),
					},
				}
			}
			dsMsg["tool_calls"] = toolCalls
		}

		// Handle tool response messages
		if msg.ToolCallID != "" {
			dsMsg["tool_call_id"] = msg.ToolCallID
			dsMsg["name"] = msg.Name
		}

		deepseekMessages = append(deepseekMessages, dsMsg)
	}

	var msgs []interface{}
	for _, m := range deepseekMessages {
		msgs = append(msgs, m)
	}

	return ReqMsg{
		Model:     models[cfg.Model],
		Messages:  msgs,
		MaxTokens: 4096,
		Tools:     toolsRaw,
	}
}

// ConvertRespMessageToAgent converts DeepSeek response to agent types
func ConvertRespMessageToAgent(resp *RespMsg) (string, agent.Tool) {
	if len(resp.Choices) == 0 {
		return "", agent.Tool{}
	}

	message := resp.Choices[0].Message
	content := ""
	if message.Content != nil {
		content = *message.Content
	}

	tool := agent.Tool{}
	if len(message.ToolCalls) > 0 {
		toolCall := message.ToolCalls[0]
		tool = agent.Tool{
			ID:   toolCall.Id,
			Type: toolCall.Type,
			Function: agent.Function{
				Name:      toolCall.Function.Name,
				Arguments: json.RawMessage(toolCall.Function.Arguments),
			},
		}
	}

	return content, tool
}

// BuildToolCallMessage builds a tool call message for conversation history
func BuildToolCallMessage(tool agent.Tool) agent.ToolCall {
	return agent.ToolCall{
		ID:   tool.ID,
		Type: tool.Type,
		Function: agent.ToolFunction{
			Name:      tool.Function.Name,
			Arguments: []byte(tool.Function.Arguments),
		},
	}
}

// ExecuteTool executes a tool and returns the result
func ExecuteTool(tool agent.Tool) (string, error) {
	toolFunction, ok := tools.IndexFunctions[tool.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool function: %s", tool.Function.Name)
	}

	return toolFunction(string(tool.Function.Arguments))
}
