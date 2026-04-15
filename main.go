package main

import (
	"fmt"
	"os"

	"github.com/Nonnika/agentcmd/commandline/repl"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}
	repl.StartLoop(os.Stdin, os.Stdout)
}
