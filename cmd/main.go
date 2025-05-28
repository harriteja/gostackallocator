package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

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

		// Set the API key from environment if available
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.OpenAIAPIKey = apiKey
		}

		// Set default model to gpt-4o-mini
		config.OpenAIModel = "gpt-4o-mini"

		// Parse command line flags
		fs := flag.NewFlagSet("stackalloc", flag.ContinueOnError)
		config.SetupFlags(fs)

		// Parse the arguments (skip program name)
		args := os.Args[1:]
		// Filter out go vet specific args
		var stackallocArgs []string
		for i, arg := range args {
			if strings.HasPrefix(arg, "-openai-") ||
				strings.HasPrefix(arg, "-autofix") ||
				strings.HasPrefix(arg, "-metrics-") ||
				strings.HasPrefix(arg, "-max-alloc-") ||
				strings.HasPrefix(arg, "-disable-") {
				stackallocArgs = append(stackallocArgs, arg)
				// Check if next arg is a value (not starting with -)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					stackallocArgs = append(stackallocArgs, args[i+1])
				}
			}
		}

		if len(stackallocArgs) > 0 {
			fs.Parse(stackallocArgs)
			config.ParseFlags(fs)
		}

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
