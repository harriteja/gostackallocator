package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/harriteja/gostackallocator/adapter"
	"github.com/harriteja/gostackallocator/analyzer"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	// Check if we should use dependency injection mode
	if shouldUseDI() {
		runWithDI()
	} else {
		// Standard mode - use the default analyzer
		unitchecker.Main(analyzer.Analyzer)
	}
}

// shouldUseDI determines if dependency injection should be used
func shouldUseDI() bool {
	// Use DI if specific flags are present or environment variables are set
	for _, arg := range os.Args {
		if arg == "-openai-api-key" || arg == "-metrics-enabled" {
			return true
		}
	}
	return os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("STACKALLOC_USE_DI") == "true"
}

// runWithDI runs the analyzer with dependency injection
func runWithDI() {
	container := buildContainer()

	err := container.Invoke(func(a *analysis.Analyzer) {
		unitchecker.Main(a)
	})

	if err != nil {
		log.Fatalf("Failed to run analyzer with DI: %v", err)
	}
}

// buildContainer sets up the dependency injection container
func buildContainer() *dig.Container {
	container := dig.New()

	// Provide configuration
	container.Provide(func() *analyzer.Config {
		config := analyzer.DefaultConfig()

		// Parse command line flags
		fs := flag.NewFlagSet("stackalloc", flag.ExitOnError)
		config.SetupFlags(fs)

		return config
	})

	// Provide logger
	container.Provide(func() *zap.Logger {
		logger, err := zap.NewDevelopment()
		if err != nil {
			// Fallback to no-op logger
			return zap.NewNop()
		}
		return logger
	})

	// Provide AI client
	container.Provide(func(config *analyzer.Config, logger *zap.Logger) analyzer.AIClient {
		if config.OpenAIDisable || config.OpenAIAPIKey == "" {
			return &NoOpAIClient{}
		}

		return adapter.NewOpenAIAdapter(
			config.OpenAIAPIKey,
			config.OpenAIModel,
			config.OpenAIMaxTokens,
			config.OpenAITemperature,
			logger,
		)
	})

	// Provide metrics client
	container.Provide(func(config *analyzer.Config, logger *zap.Logger) analyzer.MetricsClient {
		if !config.MetricsEnabled {
			return adapter.NewNoOpMetricsAdapter()
		}

		return adapter.NewMetricsAdapter(logger)
	})

	// Provide analyzer
	container.Provide(analyzer.NewAnalyzer)

	return container
}

// NoOpAIClient provides a no-op implementation when AI is disabled
type NoOpAIClient struct{}

func (n *NoOpAIClient) SuggestFix(ctx context.Context, snippet, issueMsg string) (string, error) {
	return "", nil
}
