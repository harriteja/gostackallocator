package analyzer

import (
	"context"
	"go/token"
)

// Issue represents a detected allocation issue
type Issue struct {
	Pos     token.Position // file:line:col
	Message string         // suggestion text
}

// Config holds configuration options for the analyzer
type Config struct {
	MaxAllocSize      int      // Maximum bytes to consider "small"
	DisablePatterns   []string // List of detectors to skip
	MetricsEnabled    bool     // Expose Prometheus metrics
	OpenAIAPIKey      string   // OpenAI API key
	OpenAIModel       string   // OpenAI model to use
	OpenAIMaxTokens   int      // Maximum tokens for OpenAI response
	OpenAITemperature float32  // Temperature for OpenAI requests
	OpenAIDisable     bool     // Disable AI suggestions
	AutoFix           bool     // Enable automatic code fixes
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		MaxAllocSize:      32,
		DisablePatterns:   []string{},
		MetricsEnabled:    false,
		OpenAIModel:       "gpt-4",
		OpenAIMaxTokens:   512,
		OpenAITemperature: 0.2,
		OpenAIDisable:     false,
		AutoFix:           false, // Disabled by default for safety
	}
}

// AIClient interface for AI-powered code suggestions
type AIClient interface {
	SuggestFix(ctx context.Context, snippet, issueMsg string) (string, error)
}

// MetricsClient interface for telemetry
type MetricsClient interface {
	IncrementFilesAnalyzed()
	IncrementIssuesFound()
	RecordAnalysisDuration(duration float64)
}

// Detector interface for extensible pattern detection
type Detector interface {
	Name() string
	Detect(file FileContext) []Issue
}

// FileContext provides context for analysis
type FileContext struct {
	File     interface{} // *ast.File
	TypeInfo interface{} // *types.Info
	Fset     *token.FileSet
}
