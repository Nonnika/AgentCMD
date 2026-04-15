// This package implements a simple Read-Evaluate-Print Loop (REPL) for the AgentCmd application.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Nonnika/agentcmd/agent"
	"github.com/Nonnika/agentcmd/agent/provider/deepseek"
	"github.com/Nonnika/agentcmd/agent/tools"
	"github.com/Nonnika/agentcmd/animations/header"
	"github.com/Nonnika/agentcmd/animations/loading"
)

// Read Execute Print Loop
func StartLoop(input io.Reader, output io.Writer) error {
	// Display header animation
	header.StartHeader()

	scanner := bufio.NewScanner(input)
	chatMessages := []agent.Message{
		{Role: "system", Content: &[]string{`You are a helpful assistant. You have access to the following tools:

1. CreateFile(fp: string) - Creates a file at the specified path. Use this when the user asks to create a file.
2. PwdCommand() - Returns the current working directory.

When the user asks you to create a file, you should call the CreateFile function with the appropriate file path parameter.`}[0]},
	}

	var cfg = deepseek.Config{ApiKey: os.Getenv("API_KEY")}
	var client = deepseek.NewClient(&cfg)

	for {
		fmt.Fprintf(output, "\033[33mUser: \033[0m")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(output, "Error reading input: %v\n", err)
				return err
			}
			// EOF
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "/exit" {
			fmt.Fprintf(output, "\nExiting REPL. Goodbye!\n")
			break
		}
		err := ExecuteCommand(line, output, &chatMessages, client)
		if err != nil {
			fmt.Fprintf(output, "\nError executing command: %v\n", err)
			return err
		}
	}
	return nil
}

// executeFunctionCall executes a tool function call and returns the result
func executeFunctionCall(tool agent.Tool, output io.Writer) string {
	if tool.Function.Name == "" {
		return ""
	}

	toolFunction, ok := tools.IndexFunctions[tool.Function.Name]
	if !ok {
		return fmt.Sprintf("Error: unknown function %s", tool.Function.Name)
	}

	// Convert json.RawMessage arguments to string
	args := string(tool.Function.Arguments)

	result, err := toolFunction(args)
	if err != nil {
		return fmt.Sprintf("Error executing %s: %v", tool.Function.Name, err)
	}

	return result
}

// handleToolCall handles a single tool call from the model
// It executes the tool and returns the result that should be added to conversation history
func handleToolCall(tool agent.Tool, output io.Writer) (toolCallMsg agent.Message, toolResultMsg agent.Message) {
	// Build tool call message
	toolCallMsg = agent.Message{
		Role: "assistant",
		ToolCalls: []agent.ToolCall{{
			ID:   tool.ID,
			Type: tool.Type,
			Function: agent.ToolFunction{
				Name:      tool.Function.Name,
				Arguments: tool.Function.Arguments,
			},
		}},
	}

	// Execute the tool
	result := executeFunctionCall(tool, output)

	// Build tool result message
	toolResultMsg = agent.Message{
		Role:       "tool",
		Content:    &result,
		ToolCallID: tool.ID,
		Name:       tool.Function.Name,
	}

	return toolCallMsg, toolResultMsg
}

// Execute Command or Send Message to Agent
func ExecuteCommand(cmd string, output io.Writer, chatMessages *[]agent.Message, client agent.Client) error {
	fmt.Fprintf(output, "\033[34m")
	args := []string{cmd}
	if args[0] == "/help" {
		fmt.Fprintf(output, "Available commands:\n")
		fmt.Fprintf(output, "\t/help - Show this help message\n")
		fmt.Fprintf(output, "\t/exit - Exit the REPL\n")
		fmt.Fprintf(output, "\tAny other input will be sent to the agent for processing.\n")
		fmt.Fprintf(output, "\033[0m")
		return nil
	}
	fmt.Fprintf(output, "\033[0m")

	// Add user message to chat history
	userMsg := agent.Message{Role: "user", Content: &cmd}
	*chatMessages = append(*chatMessages, userMsg)

	// Call the model with current conversation history
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	resp := make(chan string, 1)
	toolCh := make(chan agent.Tool, 1)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		response, tool, err := client.Chat(ctx, *chatMessages)
		if err != nil {
			resp <- fmt.Sprintf("Error: %v", err)
			return
		}
		resp <- response
		toolCh <- tool
	}()

	go loading.LoadingNormal(done, "Thinking...", "Done.", output)
	wg.Wait()
	done <- struct{}{}
	close(done)

	response := <-resp
	if strings.HasPrefix(response, "Error:") {
		fmt.Fprintf(output, "%s\n", response)
		return fmt.Errorf("agent error: %s", response)
	}

	tool := <-toolCh

	// Clear the loading animation line
	fmt.Fprintf(output, "\r                                           ")

	// Check if the model made a tool call
	if tool.Function.Name != "" {
		// Display the tool call
		fmt.Fprintf(output, "\r\033[34mBroith: \033[0mCalling function %s\n", tool.Function.Name)
		fmt.Fprintf(output, "Arguments: %s\n", string(tool.Function.Arguments))

		// Handle the tool call
		toolCallMsg, toolResultMsg := handleToolCall(tool, output)

		// Add messages to chat history
		*chatMessages = append(*chatMessages, toolCallMsg, toolResultMsg)

		// Call the model again with the tool result
		return ExecuteCommand("", output, chatMessages, client)
	}

	// No tool call, just display the response
	fmt.Fprintf(output, "\r\033[34mBroith: \033[0m%s\n", response)

	// Add assistant response to chat history
	*chatMessages = append(*chatMessages, agent.Message{Role: "assistant", Content: &response})

	return nil
}
