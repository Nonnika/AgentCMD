// This package implements a simple Read-Evaluate-Print Loop (REPL) for the AgentCmd application.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Nonnika/agentcmd/animations/header"
)

// Read Execute Print Loop
func StartLoop(input io.Reader, output io.Writer) error {
	// Display header animation
	header.StartHeader()

	scanner := bufio.NewScanner(input)
	for {
		fmt.Fprintf(output, "\033[36mEnter a command: \033[0m\n")
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
			fmt.Fprintf(output, "Exiting REPL. Goodbye!\n")
			break
		}
		err := ExecuteCommand(line, output)
		if err != nil {
			fmt.Fprintf(output, "Error executing command: %v\n", err)
			return err
		}
	}
	return nil
}

// Execute Command or Send Message to Agent
func ExecuteCommand(cmd string, output io.Writer) error {
	args := []string{cmd}
	if args[0] == "/help" {
		fmt.Fprintf(output, "Available commands:\n")
		fmt.Fprintf(output, "/help - Show this help message\n")
		fmt.Fprintf(output, "/exit - Exit the REPL\n")
		fmt.Fprintf(output, "Any other input will be sent to the agent for processing.\n")
		return nil
	}
	return nil
}
