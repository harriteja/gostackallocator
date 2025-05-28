package internal

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// Logger provides a global logger instance
var Logger *zap.Logger

// InitLogger initializes the global logger
func InitLogger(development bool) error {
	var err error
	if development {
		Logger, err = zap.NewDevelopment()
	} else {
		Logger, err = zap.NewProduction()
	}

	if err != nil {
		Logger = zap.NewNop()
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}

// GetLogger returns the global logger or a no-op logger if not initialized
func GetLogger() *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger
}

// ReadSourceFile reads the source file for a given position
func ReadSourceFile(pos token.Position) ([]byte, error) {
	if pos.Filename == "" {
		return nil, fmt.Errorf("no filename in position")
	}

	// Check if file exists
	if _, err := os.Stat(pos.Filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", pos.Filename)
	}

	return ioutil.ReadFile(pos.Filename)
}

// GetProjectRoot finds the project root by looking for go.mod
func GetProjectRoot(startPath string) (string, error) {
	dir := startPath
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			return "", err
		}
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found in any parent directory")
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// SafeString returns a safe string representation, handling nil pointers
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
