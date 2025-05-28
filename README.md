# stackalloc - Go Static Analysis for Stack Allocation Optimization

[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`stackalloc` is a Go static analysis tool that detects small heap allocations and suggests stack-friendly alternatives. It helps optimize memory allocation patterns by identifying cases where stack allocation could be used instead of heap allocation.

## Features

- 🔍 **Pattern Detection**: Identifies common allocation patterns that could benefit from stack allocation
- 🤖 **AI-Powered Suggestions**: Optional integration with OpenAI for enhanced code suggestions
- 📊 **Metrics & Telemetry**: Prometheus metrics support for monitoring analysis performance
- 🔧 **Configurable**: Extensive configuration options for different use cases
- 🚀 **Fast**: Efficient AST-based analysis with minimal overhead
- 🔌 **Extensible**: Plugin architecture for custom detectors

## Installation

### Using go install

```bash
go install github.com/harriteja/gostackallocator/cmd@latest
```

### Building from source

```bash
git clone https://github.com/harriteja/gostackallocator.git
cd gostackallocator
go build -o stackalloc ./cmd
```

## Usage

### Basic Usage

Run as a standalone analyzer:

```bash
stackalloc ./...
```

### Integration with go vet

```bash
go vet -vettool=stackalloc ./...
```

### With Configuration

```bash
stackalloc -stackalloc.max-alloc-size=64 -stackalloc.metrics-enabled ./...
```

### With AI Suggestions

```bash
export OPENAI_API_KEY="your-api-key"
stackalloc -openai-model=gpt-4 ./...
```

### With AI Suggestions and Automatic Fixes

```bash
export OPENAI_API_KEY="your-api-key"
STACKALLOC_USE_DI=true go vet -vettool=./stackalloc -autofix ./...
```

## Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `-max-alloc-size` | `32` | Maximum bytes to consider "small" allocation |
| `-disable-patterns` | `""` | Comma-separated list of detectors to skip |
| `-stackalloc.metrics-enabled` | `false` | Expose Prometheus metrics |
| `-openai-api-key` | `""` | OpenAI API key for AI suggestions |
| `-openai-model` | `gpt-4` | OpenAI model to use |
| `-openai-max-tokens` | `512` | Maximum tokens for OpenAI response |
| `-openai-temperature` | `0.2` | Temperature for OpenAI requests |
| `-openai-disable` | `false` | Disable AI-powered suggestions |
| `-autofix` | `false` | Enable automatic code fixes (use with caution) |

### Environment Variables

- `OPENAI_API_KEY`: OpenAI API key (alternative to `-openai-api-key` flag)
- `STACKALLOC_USE_DI`: Enable dependency injection mode (`true`/`false`)

## Detection Patterns

### 1. Pointer to Local Variable Escape

Detects when a pointer to a local variable escapes the function scope:

```go
func bad() *int {
    x := 42
    return &x  // ❌ Detected: pointer escapes only once
}

func good() int {
    x := 42
    return x   // ✅ Better: return value directly
}
```

### 2. new(T) Allocations

Identifies `new(T)` calls that could be replaced with stack allocation:

```go
func bad() *string {
    s := new(string)  // ❌ Detected: new() always allocates on heap
    *s = "hello"
    return s
}

func good() string {
    return "hello"    // ✅ Better: return value directly
}
```

### 3. Object Reuse Analysis

Avoids false positives by tracking object usage patterns:

```go
func reused() {
    data := make([]int, 100)
    process(data)
    process(data)     // ✅ Not flagged: object is reused
    process(data)
}
```

## Architecture

```
┌─────────────┐        ┌──────────────┐        ┌──────────────┐
│  CLI / Vet  │ ──>    │ Analyzer Core│ ──>    │ Report Engine│ ──> Output
└─────────────┘        └──────────────┘        └──────────────┘
        │                     │                       │
        │                     │                       │
        ▼                     ▼                       ▼
  Parse command       Walk AST & type info    Format diagnostics
        │                     │                       │
        └───────────>─────────┴─────────>─────────────┘
```

### Package Structure

```
stackalloc/
├── analyzer/          # Core analysis logic
│   ├── analyzer.go    # Main analyzer definition
│   ├── inspector.go   # AST inspection & pattern matching
│   ├── reporter.go    # Suggestion formatting
│   ├── types.go       # Data models
│   └── config.go      # Configuration handling
├── adapter/           # External service adapters
│   ├── openai_adapter.go    # OpenAI integration
│   └── metrics_adapter.go   # Prometheus metrics
├── cmd/               # CLI entry point
│   └── main.go
├── internal/          # Internal utilities
│   ├── utils.go       # Common utilities
│   └── metrics.go     # Telemetry definitions
└── examples/          # Sample code
    └── sample.go
```

## AI Integration

The analyzer can integrate with OpenAI's API to provide enhanced code suggestions:

1. **Setup**: Provide your OpenAI API key via environment variable or flag
2. **Analysis**: The analyzer extracts code snippets around detected issues
3. **Enhancement**: OpenAI generates context-aware suggestions
4. **Output**: Suggestions are included in the analysis diagnostics

### Automatic Code Fixes

When the `-autofix` flag is enabled along with AI integration, the analyzer can automatically generate code fixes:

1. **AI Analysis**: OpenAI analyzes the problematic code and suggests improvements
2. **Code Parsing**: The analyzer parses AI suggestions to extract actionable code changes
3. **Smart Replacement**: Generates `analysis.SuggestedFix` with actual code replacements
4. **Safety**: Validates suggested code before applying fixes

**⚠️ Important**: Automatic fixes are experimental. Always review changes before applying them to production code.

Example output with automatic fixes:

```
example.go:10:2: new(T) always allocates on heap; consider using stack allocation if object doesn't escape
    Suggested fix: Replace new(T) with stack allocation
    - s := new(string)
    + var s string; return &s
```

### AI Suggestion Examples

Example output with AI suggestions:

```
example.go:10:2: pointer to x escapes only once; consider using stack allocation
    AI suggestion: Instead of returning &x, consider returning the value directly
    or using a value receiver pattern to avoid heap allocation.
```

## Metrics

When metrics are enabled, the analyzer exposes Prometheus metrics:

- `stackalloc_files_analyzed_total`: Total number of files analyzed
- `stackalloc_issues_found_total`: Total number of issues found
- `stackalloc_analysis_duration_seconds`: Time spent analyzing files

Access metrics at `http://localhost:8080/metrics` (when running in server mode).

## Development

### Prerequisites

- Go 1.20 or later
- Optional: OpenAI API key for AI features

### Building

```bash
go build ./cmd
```

### Testing

```bash
go test ./...
```

### Running on Examples

```bash
go vet -vettool=./stackalloc ./examples/
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built using [golang.org/x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- AI integration powered by [OpenAI](https://openai.com/)
- Metrics provided by [Prometheus](https://prometheus.io/)
- Dependency injection via [Uber Dig](https://github.com/uber-go/dig)
- Logging with [Uber Zap](https://github.com/uber-go/zap)

## Roadmap

- [ ] Support for more allocation patterns
- [ ] Integration with other AI providers
- [ ] Performance optimizations
- [ ] IDE plugin support
- [ ] Custom rule definitions
- [ ] Batch processing mode 