// files.go
// This file contains the tools related to files.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateFileWrapper is the JSON-compatible wrapper for CreateFile
// It receives a JSON string {"fp": "/path/to/file"} and returns the result
func CreateFileWrapper(jsonArgs string) (string, error) {
	var args struct {
		Fp string `json:"fp"`
	}
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}
	return CreateFile(args.Fp)
}

// CreateFile creates a file at the specified path.
func CreateFile(fp string) (string, error) {
	dir := filepath.Dir(fp)
	// check the directory exists, if not, create it
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s : %w", dir, err)
	}

	if _, err := os.Stat(fp); err == nil {
		return "", fmt.Errorf("file %s already exists", fp)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check if file exists: %w", err)
	}

	// create the file
	file, err := os.Create(fp)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return "File created successfully", nil
}

// PwdCommandWrapper is the JSON-compatible wrapper for PwdCommand
func PwdCommandWrapper(jsonArgs string) (string, error) {
	// PwdCommand takes no arguments, ignore jsonArgs
	return PwdCommand()
}

// Get current working directory (not implement yet, just for test)
func PwdCommand() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}
