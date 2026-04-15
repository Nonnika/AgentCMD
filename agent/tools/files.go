// files.go
// This file contains the tools related to files.
package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateFile creates a file at the specified path.
func CreateFile(fp string) error {
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s : %w", dir, err)
	}

	if _, err := os.Stat(fp); err == nil {
		return fmt.Errorf("file %s already exists", fp)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	file, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return nil
}

func PwdCommand() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}
