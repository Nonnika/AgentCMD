// This package is used for DeepSeek Model
// Include create chat client , reasoner client
// Each client has its own config and http client
// The config is used to set the model, api key, and baseURL for each client
// The http client is used to send requests to DeepSeek Model
package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Default BaseURL for DeepSeek Model
const BaseURL = "https://api.deepseek.com"

type Config struct {
	Model   int
	ApiKey  string
	BaseURL string
}

// Client for DeepSeek Model
type Client struct {
	cfg        *Config
	httpClient *http.Client // To avoid goroutine leak
	// http.DefaultClient is safe for concurrent use, but if someone change in somewhere ,the global config will be changed
}

// Create a New Client
func NewClient(cfg *Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = BaseURL
	}
	if cfg.Model == 0 {
		cfg.Model = DeepseekChat
	}

	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// Create and Send a chat request
func (c *Client) Chat(ctx context.Context, messages []Message) (string, string, error) {
	ReqMsg := BuildReqMsg(c.cfg, messages)

	reqBody, err := json.Marshal(ReqMsg)
	if err != nil {
		return "", "", fmt.Errorf("marshal req msg to json: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", "", fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.ApiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("http status code not ok: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response body: %w", err)
	}

	var respMsg RespMsg
	if err := json.Unmarshal(body, &respMsg); err != nil {
		return "", "", fmt.Errorf("unmarshal response body to resp msg: %w", err)
	}

	content := ""
	if respMsg.Choices[0].Message.Content != nil {
		content = *respMsg.Choices[0].Message.Content
	}
	var toolCalls strings.Builder
	if respMsg.Choices[0].Message.ToolCalls != nil {
		for _, tool := range respMsg.Choices[0].Message.ToolCalls {
			toolCalls.WriteString(tool.Function.Name + " " + tool.Function.Arguments + "\n")
		}
	}

	return content, toolCalls.String(), nil
}
