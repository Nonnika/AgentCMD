// This package is used for tools index
package tools

import (
	"encoding/json"
	"fmt"
)

// ToolFunc is the type for tool functions
// Each tool receives a JSON string containing the arguments and returns the result or error
type ToolFunc func(jsonArgs string) (string, error)

// IndexFunctions maps tool names to their handler functions
var IndexFunctions = map[string]ToolFunc{
	"CreateFile": CreateFileWrapper,
	"PwdCommand": PwdCommandWrapper,
}

// ToolDef represents a tool definition for the model
type ToolDef struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents the function details in a tool definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolParameter represents a single parameter definition
type ToolParameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// GenerateToolsJSON generates the tools index for the model
func GenerateToolsJSON() []string {
	tools := []ToolDef{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "CreateFile",
				Description: "Create a file at the specified path. If the directory does not exist, it will be created automatically.",
				Parameters: map[string]interface{}{
					"type":     "object",
					"required": []string{"fp"},
					"properties": map[string]interface{}{
						"fp": map[string]interface{}{
							"type":        "string",
							"description": "The path to create the file, including filename.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "PwdCommand",
				Description: "Get the current working directory.",
				Parameters: map[string]interface{}{
					"type":     "object",
					"required": []string{},
					"properties": map[string]interface{}{},
				},
			},
		},
	}

	result := make([]string, len(tools))
	for i, tool := range tools {
		// DeepSeek 使用标准的 OpenAI 格式
		toolJSON := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			},
		}
		data, err := json.Marshal(toolJSON)
		if err != nil {
			fmt.Printf("Error marshaling tool %s: %v\n", tool.Function.Name, err)
			continue
		}
		result[i] = string(data)
	}
	return result
}

// ToolsIndex is the cached tools index for the model
var ToolsIndex = GenerateToolsJSON()
