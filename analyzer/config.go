package analyzer

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

// SetupFlags configures command-line flags for the analyzer
func (c *Config) SetupFlags(fs *flag.FlagSet) {
	fs.IntVar(&c.MaxAllocSize, "max-alloc-size", c.MaxAllocSize,
		"Maximum bytes to consider 'small' allocation")

	var disablePatterns string
	fs.StringVar(&disablePatterns, "disable-patterns", "",
		"Comma-separated list of detectors to skip")

	fs.BoolVar(&c.MetricsEnabled, "metrics-enabled", c.MetricsEnabled,
		"Expose Prometheus metrics")

	fs.StringVar(&c.OpenAIAPIKey, "openai-api-key", c.OpenAIAPIKey,
		"OpenAI API key (can also use OPENAI_API_KEY env var)")

	fs.StringVar(&c.OpenAIModel, "openai-model", c.OpenAIModel,
		"OpenAI model to use for suggestions")

	fs.IntVar(&c.OpenAIMaxTokens, "openai-max-tokens", c.OpenAIMaxTokens,
		"Maximum tokens for OpenAI response")

	var temperature string
	fs.StringVar(&temperature, "openai-temperature", "0.2",
		"Temperature for OpenAI requests (0.0-1.0)")

	fs.BoolVar(&c.OpenAIDisable, "openai-disable", c.OpenAIDisable,
		"Disable AI-powered suggestions")

	fs.BoolVar(&c.AutoFix, "autofix", c.AutoFix,
		"Enable automatic code fixes (use with caution)")

	// Note: We don't call Parse here as the analysis framework handles that

	// Process disable patterns if provided
	if disablePatterns != "" {
		c.DisablePatterns = strings.Split(disablePatterns, ",")
		for i := range c.DisablePatterns {
			c.DisablePatterns[i] = strings.TrimSpace(c.DisablePatterns[i])
		}
	}

	// Parse temperature
	if temp, err := strconv.ParseFloat(temperature, 32); err == nil {
		c.OpenAITemperature = float32(temp)
	}

	// Check environment variable for API key if not provided
	if c.OpenAIAPIKey == "" {
		c.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	}
}

// ParseFlags processes flag values after they've been parsed
func (c *Config) ParseFlags(fs *flag.FlagSet) {
	// This can be called after flag parsing to process complex flag values
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "disable-patterns":
			if f.Value.String() != "" {
				c.DisablePatterns = strings.Split(f.Value.String(), ",")
				for i := range c.DisablePatterns {
					c.DisablePatterns[i] = strings.TrimSpace(c.DisablePatterns[i])
				}
			}
		case "openai-temperature":
			if temp, err := strconv.ParseFloat(f.Value.String(), 32); err == nil {
				c.OpenAITemperature = float32(temp)
			}
		}
	})

	// Check environment variable for API key if not provided via flag
	if c.OpenAIAPIKey == "" {
		c.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	}
}

// IsPatternDisabled checks if a specific pattern detector is disabled
func (c *Config) IsPatternDisabled(pattern string) bool {
	for _, disabled := range c.DisablePatterns {
		if disabled == pattern {
			return true
		}
	}
	return false
}
