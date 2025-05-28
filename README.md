# stackalloc - Go Static Analysis for Stack Allocation Optimization

[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`stackalloc` is a Go static analysis tool that detects small heap allocations and suggests stack-friendly alternatives. It helps optimize memory allocation patterns by identifying cases where stack allocation could be used instead of heap allocation.

## Features

- 🔍 **Pattern Detection**: Identifies common allocation patterns that could benefit from stack allocation
- 🤖 **AI-Powered Suggestions**: Optional integration with OpenAI for enhanced code suggestions with actual code fixes
- 📊 **Metrics & Telemetry**: Prometheus metrics support for monitoring analysis performance
- 🔧 **Configurable**: Extensive configuration options for different use cases
- 🚀 **Fast**: Efficient AST-based analysis with minimal overhead
- 🔌 **Extensible**: Plugin architecture for custom detectors

## Quick Start

### Installation

```bash
# Option 1: Install from source
git clone https://github.com/harriteja/gostackallocator.git
cd gostackallocator
go build -o stackalloc ./cmd

# Option 2: Using go install (when published)
go install github.com/harriteja/gostackallocator/cmd@latest
```

### Basic Usage

```bash
# Analyze entire project
go vet -vettool=stackalloc ./...

# Analyze specific package
go vet -vettool=stackalloc ./internal/handlers/...

# Analyze single file
go vet -vettool=stackalloc ./main.go

# With custom threshold (stricter for performance-critical code)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./...
```

### AI-Powered Analysis

```bash
# Set up OpenAI API key
export OPENAI_API_KEY="your-api-key"

# Basic AI analysis
STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./...

# AI with automatic code fixes (experimental)
STACKALLOC_USE_DI=true go vet -vettool=stackalloc -stackalloc.autofix ./...
```

### Helper Scripts

```bash
# Use automated analysis script
./scripts/analyze-project.sh -b ./stackalloc --ai --autofix

# Quick setup for your project
./scripts/quick-setup.sh --setup-ide --setup-ci
```

## What It Detects

### Example Output

```
examples/sample.go:13:6: new(T) always allocates on heap; consider using stack allocation if object doesn't escape
    AI suggested fix:
    - s := new(string)
    + return "hello"  // Direct value return

examples/sample.go:15:6: pointer to x escapes only once; consider using stack allocation
    AI suggested fix:
    - return &x
    + return x  // Return by value instead of pointer
```

## Documentation

- **[USAGE.md](USAGE.md)**: Comprehensive usage guide with real-world examples, CI/CD integration, and advanced configuration
- **[Examples](examples/)**: Sample code demonstrating detected patterns
- **[Scripts](scripts/)**: Helper scripts for automated analysis and project setup

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

`stackalloc` follows a clean 3-layer architecture:

```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐
│  CLI / Vet  │ -> │ Analyzer Core│ -> │ Report Engine│ -> Output
└─────────────┘    └──────────────┘    └──────────────┘
```

**Key Components:**
- **Analyzer Core**: AST inspection and pattern detection
- **AI Integration**: OpenAI adapter for enhanced suggestions
- **Metrics**: Prometheus integration for monitoring
- **Extensible**: Plugin architecture for custom detectors

## AI Integration

`stackalloc` integrates with OpenAI to provide **concrete code fixes**, not just enhanced problem descriptions.

### How It Works

1. **Issue Detection**: Detects memory allocation issues using static analysis
2. **Context Extraction**: Extracts code snippet around the problematic line  
3. **AI Analysis**: Sends code and issue to OpenAI requesting specific fixes
4. **Code Generation**: Returns actual replacement code with before/after examples
5. **Smart Application**: Generates `analysis.SuggestedFix` objects for IDE integration

### Automatic Code Fixes

When `-stackalloc.autofix` is enabled, the analyzer provides concrete code fixes:

**⚠️ Important**: Automatic fixes are experimental. Always review changes before applying.

**Example:**
```go
// Before (problematic)
func createString() *string {
    s := new(string)  // Heap allocation
    *s = "hello"
    return s
}

// After (AI-suggested fix)
func createString() string {
    return "hello"  // Direct value return
}
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