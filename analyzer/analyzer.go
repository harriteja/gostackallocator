package analyzer

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"time"

	"golang.org/x/tools/go/analysis"
)

// NoOpMetricsAdapter is a no-op implementation of MetricsClient
type NoOpMetricsAdapter struct{}

func (n *NoOpMetricsAdapter) IncrementFilesAnalyzed()                 {}
func (n *NoOpMetricsAdapter) IncrementIssuesFound()                   {}
func (n *NoOpMetricsAdapter) RecordAnalysisDuration(duration float64) {}

// Analyzer is the main static analysis analyzer for stack allocation detection
var Analyzer = &analysis.Analyzer{
	Name:  "stackalloc",
	Doc:   "detects small heap allocations and suggests stack-friendly alternatives",
	Run:   run,
	Flags: flag.FlagSet{},
}

func init() {
	// Setup default flags for the analyzer
	config := DefaultConfig()
	config.SetupFlags(&Analyzer.Flags)
}

// AnalyzerWithDeps creates an analyzer with injected dependencies
func NewAnalyzer(aiClient AIClient, metricsClient MetricsClient, config *Config) *analysis.Analyzer {
	analyzer := &analysis.Analyzer{
		Name: "stackalloc",
		Doc:  "detects small heap allocations and suggests stack-friendly alternatives",
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return runWithDeps(pass, aiClient, metricsClient, config)
		},
	}

	// Setup flags if config is provided
	if config != nil {
		config.SetupFlags(&analyzer.Flags)
	}

	return analyzer
}

// run is the main entry point for the analyzer
func run(pass *analysis.Pass) (interface{}, error) {
	// Create config from flags
	config := DefaultConfig()
	config.ParseFlags(&pass.Analyzer.Flags)

	// Create metrics client (no-op for now)
	metricsClient := &NoOpMetricsAdapter{}

	// Create AI client if enabled (simplified for now)
	var aiClient AIClient

	// Track analysis start time
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metricsClient.RecordAnalysisDuration(duration)
	}()

	// Analyze each file
	for _, file := range pass.Files {
		metricsClient.IncrementFilesAnalyzed()

		// Use the existing InspectFile function
		InspectFile(file, pass.TypesInfo, pass.Fset, func(pos token.Pos, msg string) {
			metricsClient.IncrementIssuesFound()

			// Create issue
			issue := Issue{
				Pos:     pass.Fset.Position(pos),
				Message: msg,
			}

			// Report issue
			ReportIssue(pass, issue, aiClient, config)
		})
	}

	return nil, nil
}

// runWithDeps runs the analysis with injected dependencies
func runWithDeps(pass *analysis.Pass, aiClient AIClient, metricsClient MetricsClient, config *Config) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			pass.Reportf(token.NoPos, "stackalloc panicked: %v", r)
		}
	}()

	startTime := time.Now()

	// Increment files analyzed metric
	if metricsClient != nil {
		metricsClient.IncrementFilesAnalyzed()
		defer func() {
			duration := time.Since(startTime).Seconds()
			metricsClient.RecordAnalysisDuration(duration)
		}()
	}

	var issuesFound int

	// Analyze each file in the package
	for _, file := range pass.Files {
		issues := analyzeFile(file, pass.TypesInfo, pass.Fset, config)

		for _, issue := range issues {
			ReportIssue(pass, issue, aiClient, config)
			issuesFound++
		}
	}

	// Record metrics
	if metricsClient != nil && issuesFound > 0 {
		for i := 0; i < issuesFound; i++ {
			metricsClient.IncrementIssuesFound()
		}
	}

	return nil, nil
}

// analyzeFile analyzes a single file for allocation patterns
func analyzeFile(file *ast.File, info *types.Info, fset *token.FileSet, config *Config) []Issue {
	var issues []Issue

	// Collect issues using the inspector
	InspectFile(file, info, fset, func(pos token.Pos, msg string) {
		position := fset.Position(pos)
		issue := Issue{
			Pos:     position,
			Message: msg,
		}
		issues = append(issues, issue)
	})

	return issues
}

// GetVersion returns the analyzer version
func GetVersion() string {
	return "v0.1.0"
}

// GetDescription returns a detailed description of the analyzer
func GetDescription() string {
	return fmt.Sprintf(`stackalloc %s - Static Analysis for Stack Allocation Optimization

This analyzer detects small heap allocations that could potentially be optimized
to use stack allocation instead. It identifies patterns such as:

1. Pointers to local variables that escape only once
2. new(T) calls that could be replaced with stack allocation
3. Object reuse patterns to avoid false positives

The analyzer provides suggestions for optimizing memory allocation patterns
and can integrate with AI services for enhanced code suggestions.

Flags:
  -max-alloc-size=N     Maximum bytes to consider 'small' allocation (default: 32)
  -disable-patterns=P   Comma-separated list of detectors to skip
  -metrics-enabled      Expose Prometheus metrics (default: false)
  -openai-api-key=KEY   OpenAI API key for AI suggestions
  -openai-model=MODEL   OpenAI model to use (default: gpt-4)
  -openai-disable       Disable AI-powered suggestions (default: false)

Environment Variables:
  OPENAI_API_KEY        OpenAI API key (alternative to -openai-api-key flag)
`, GetVersion())
}
