package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Nonnika/agentcmd/agent/provider/deepseek"
	"github.com/Nonnika/agentcmd/commandline/repl"
	"github.com/joho/godotenv"
)

func ChatWithDeepSeek(msg []deepseek.Message) (string, string) {
	cfg := deepseek.Config{ApiKey: os.Getenv("DEEPSEEK_API_KEY")}
	client := deepseek.NewClient(&cfg)
	ctx := context.Background()
	resp, tools, err := client.Chat(ctx, msg)
	if err != nil {
		return "", fmt.Sprintf("Failed to send chat request: %v", err)
	}
	return resp, tools
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}
	repl.StartLoop(os.Stdin, os.Stdout)
}
