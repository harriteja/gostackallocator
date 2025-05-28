# stackalloc Usage Guide

This guide explains how to use `stackalloc` to analyze and optimize memory allocation patterns in your Go projects.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Project-Level Analysis](#project-level-analysis)
- [Folder-Level Analysis](#folder-level-analysis)
- [File-Level Analysis](#file-level-analysis)
- [Configuration Options](#configuration-options)
- [AI-Powered Analysis](#ai-powered-analysis)
- [Continuous Integration](#continuous-integration)
- [IDE Integration](#ide-integration)
- [Troubleshooting](#troubleshooting)

## Installation

### Option 1: Install from Source
```bash
git clone https://github.com/harriteja/gostackallocator.git
cd gostackallocator
go build -o stackalloc ./cmd
sudo mv stackalloc /usr/local/bin/  # Optional: add to PATH
```

### Option 2: Go Install (when published)
```bash
go install github.com/harriteja/gostackallocator/cmd@latest
```

### Option 3: Download Binary
Download the latest release from GitHub releases and add to your PATH.

## Basic Usage

### Quick Start
```bash
# Analyze current directory and subdirectories
go vet -vettool=stackalloc ./...

# Analyze specific package
go vet -vettool=stackalloc ./pkg/mypackage

# Analyze single file
go vet -vettool=stackalloc ./main.go
```

## Project-Level Analysis

### Entire Monolith Application

For a large monolith Go application, you'll want to analyze the entire codebase:

```bash
# Navigate to your project root
cd /path/to/your/monolith

# Analyze entire project
go vet -vettool=stackalloc ./...

# With custom configuration
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./...

# Save output to file for review
go vet -vettool=stackalloc ./... 2> stackalloc-report.txt
```

### Example Project Structure
```
your-monolith/
├── cmd/
│   ├── api/
│   └── worker/
├── internal/
│   ├── auth/
│   ├── database/
│   └── handlers/
├── pkg/
│   ├── models/
│   └── utils/
└── vendor/
```

### Analyzing with Exclusions
```bash
# Exclude vendor and test files
go vet -vettool=stackalloc $(find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go" | xargs dirname | sort -u)

# Or use build tags to exclude certain packages
go vet -vettool=stackalloc -tags="!integration" ./...
```

## Folder-Level Analysis

### Specific Modules/Packages

```bash
# Analyze specific service/module
go vet -vettool=stackalloc ./internal/auth/...

# Analyze multiple related packages
go vet -vettool=stackalloc ./internal/handlers/... ./internal/services/...

# Analyze with different configurations per module
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./internal/performance-critical/...
go vet -vettool=stackalloc -stackalloc.max-alloc-size=128 ./internal/less-critical/...
```

### Batch Analysis Script

Create a script for analyzing different parts of your application:

```bash
#!/bin/bash
# analyze-project.sh

echo "Analyzing critical performance modules..."
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./internal/core/... ./pkg/algorithms/...

echo "Analyzing API handlers..."
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./internal/handlers/...

echo "Analyzing background services..."
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./internal/workers/...

echo "Analysis complete!"
```

## File-Level Analysis

### Single File Analysis

```bash
# Analyze specific file
go vet -vettool=stackalloc ./internal/handlers/user.go

# Multiple specific files
go vet -vettool=stackalloc ./main.go ./internal/config/config.go

# With detailed output
go vet -vettool=stackalloc -v ./internal/critical-path.go
```

### Targeted Analysis for Hot Paths

```bash
# Analyze performance-critical files
go vet -vettool=stackalloc \
  ./internal/handlers/high-traffic.go \
  ./pkg/algorithms/sorting.go \
  ./internal/database/queries.go
```

## Configuration Options

### Basic Configuration

```bash
# Set allocation size threshold
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./...

# Disable specific patterns
go vet -vettool=stackalloc -stackalloc.disable-patterns="new-allocation" ./...

# Enable metrics
go vet -vettool=stackalloc -stackalloc.metrics-enabled ./...
```

### Advanced Configuration

Create a configuration file for complex setups:

```bash
# .stackalloc.config
MAX_ALLOC_SIZE=32
DISABLE_PATTERNS=new-allocation,pointer-escape
METRICS_ENABLED=true
OPENAI_MODEL=gpt-4
```

```bash
# Use configuration file
source .stackalloc.config
go vet -vettool=stackalloc \
  -stackalloc.max-alloc-size=$MAX_ALLOC_SIZE \
  -stackalloc.disable-patterns=$DISABLE_PATTERNS \
  -stackalloc.metrics-enabled=$METRICS_ENABLED \
  ./...
```

## AI-Powered Analysis

### With OpenAI Integration

```bash
# Set up OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# Basic AI analysis
STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./...

# AI with automatic fix suggestions
STACKALLOC_USE_DI=true go vet -vettool=stackalloc -stackalloc.autofix ./...

# Custom AI model
STACKALLOC_USE_DI=true go vet -vettool=stackalloc \
  -stackalloc.openai-model=gpt-4 \
  -stackalloc.openai-temperature=0.1 \
  ./...
```

### AI Analysis for Different Project Parts

```bash
# High-precision analysis for critical code
STACKALLOC_USE_DI=true go vet -vettool=stackalloc \
  -stackalloc.openai-temperature=0.1 \
  -stackalloc.autofix \
  ./internal/core/...

# More creative suggestions for experimental code
STACKALLOC_USE_DI=true go vet -vettool=stackalloc \
  -stackalloc.openai-temperature=0.5 \
  ./experimental/...
```

## Continuous Integration

### GitHub Actions

Create `.github/workflows/stackalloc.yml`:

```yaml
name: Stack Allocation Analysis

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  stackalloc:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.22
    
    - name: Install stackalloc
      run: |
        git clone https://github.com/harriteja/gostackallocator.git
        cd gostackallocator
        go build -o stackalloc ./cmd
        sudo mv stackalloc /usr/local/bin/
    
    - name: Run stackalloc analysis
      run: |
        go vet -vettool=stackalloc ./... || true
        
    - name: Run stackalloc with AI (if API key available)
      if: ${{ secrets.OPENAI_API_KEY }}
      env:
        OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
      run: |
        STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./... || true
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    environment {
        OPENAI_API_KEY = credentials('openai-api-key')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Install stackalloc') {
            steps {
                sh '''
                    git clone https://github.com/harriteja/gostackallocator.git
                    cd gostackallocator
                    go build -o stackalloc ./cmd
                '''
            }
        }
        
        stage('Stack Allocation Analysis') {
            steps {
                sh '''
                    export PATH=$PATH:$(pwd)/gostackallocator
                    go vet -vettool=stackalloc ./... > stackalloc-report.txt 2>&1 || true
                '''
                archiveArtifacts artifacts: 'stackalloc-report.txt'
            }
        }
        
        stage('AI-Enhanced Analysis') {
            when {
                environment name: 'OPENAI_API_KEY', value: ''
                not { equals expected: '', actual: env.OPENAI_API_KEY }
            }
            steps {
                sh '''
                    export PATH=$PATH:$(pwd)/gostackallocator
                    STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./... > stackalloc-ai-report.txt 2>&1 || true
                '''
                archiveArtifacts artifacts: 'stackalloc-ai-report.txt'
            }
        }
    }
}
```

## IDE Integration

### VS Code

Create `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "stackalloc: Analyze Current File",
            "type": "shell",
            "command": "go",
            "args": [
                "vet",
                "-vettool=stackalloc",
                "${file}"
            ],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "stackalloc: Analyze Project",
            "type": "shell",
            "command": "go",
            "args": [
                "vet",
                "-vettool=stackalloc",
                "./..."
            ],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "stackalloc: AI Analysis",
            "type": "shell",
            "command": "bash",
            "args": [
                "-c",
                "STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./..."
            ],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
```

### GoLand/IntelliJ

1. Go to **File** → **Settings** → **Tools** → **External Tools**
2. Click **+** to add a new tool:
   - **Name**: stackalloc
   - **Program**: `go`
   - **Arguments**: `vet -vettool=stackalloc $FileDir$`
   - **Working Directory**: `$ProjectFileDir$`

## Practical Examples

### Example 1: E-commerce Monolith

```bash
# Your e-commerce application structure
ecommerce-app/
├── cmd/api/           # API server
├── cmd/worker/        # Background workers
├── internal/
│   ├── auth/         # Authentication
│   ├── cart/         # Shopping cart
│   ├── payment/      # Payment processing
│   ├── inventory/    # Inventory management
│   └── orders/       # Order processing
└── pkg/
    ├── models/       # Data models
    └── utils/        # Utilities

# Analysis strategy
# 1. Critical path analysis (payment, orders)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./internal/payment/... ./internal/orders/...

# 2. High-traffic analysis (auth, cart)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./internal/auth/... ./internal/cart/...

# 3. Background services (less critical)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./cmd/worker/... ./internal/inventory/...

# 4. Full project analysis with AI
export OPENAI_API_KEY="your-key"
STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./...
```

### Example 2: Microservices in Monorepo

```bash
# Monorepo structure
services/
├── user-service/
├── order-service/
├── payment-service/
├── notification-service/
└── shared/

# Analyze each service separately
for service in user-service order-service payment-service notification-service; do
    echo "Analyzing $service..."
    go vet -vettool=stackalloc ./services/$service/...
done

# Analyze shared libraries
go vet -vettool=stackalloc ./services/shared/...
```

### Example 3: Performance-Critical Application

```bash
# High-frequency trading application
trading-app/
├── engine/           # Trading engine (ultra-critical)
├── market-data/      # Market data processing
├── risk/            # Risk management
└── reporting/       # Reporting (less critical)

# Ultra-strict analysis for trading engine
go vet -vettool=stackalloc -stackalloc.max-alloc-size=8 ./engine/...

# Strict analysis for market data
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./market-data/...

# Normal analysis for other components
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./risk/... ./reporting/...
```

## Automation Scripts

### Daily Analysis Script

```bash
#!/bin/bash
# daily-stackalloc.sh

PROJECT_ROOT="/path/to/your/project"
REPORT_DIR="$PROJECT_ROOT/reports/stackalloc"
DATE=$(date +%Y-%m-%d)

mkdir -p "$REPORT_DIR"

cd "$PROJECT_ROOT"

echo "Running daily stackalloc analysis..."

# Basic analysis
go vet -vettool=stackalloc ./... > "$REPORT_DIR/basic-$DATE.txt" 2>&1

# AI analysis (if API key is available)
if [ ! -z "$OPENAI_API_KEY" ]; then
    echo "Running AI-enhanced analysis..."
    STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./... > "$REPORT_DIR/ai-$DATE.txt" 2>&1
fi

# Generate summary
echo "Analysis complete. Reports saved to $REPORT_DIR"
ls -la "$REPORT_DIR"/*$DATE*
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running stackalloc analysis on changed files..."

# Get list of changed Go files
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -z "$CHANGED_FILES" ]; then
    echo "No Go files changed."
    exit 0
fi

# Analyze changed files
for file in $CHANGED_FILES; do
    echo "Analyzing $file..."
    go vet -vettool=stackalloc "./$file" || {
        echo "stackalloc found issues in $file"
        echo "Please review and fix before committing."
        exit 1
    }
done

echo "stackalloc analysis passed!"
```

## Troubleshooting

### Common Issues

#### 1. "stackalloc not found"
```bash
# Solution: Ensure stackalloc is in PATH or use full path
/path/to/stackalloc/stackalloc ./...
# or
go vet -vettool=/path/to/stackalloc ./...
```

#### 2. "No issues found" (but you expect some)
```bash
# Try lowering the allocation size threshold
go vet -vettool=stackalloc -stackalloc.max-alloc-size=8 ./...

# Check if patterns are disabled
go vet -vettool=stackalloc -stackalloc.disable-patterns="" ./...
```

#### 3. "Too many issues"
```bash
# Start with higher threshold and gradually lower it
go vet -vettool=stackalloc -stackalloc.max-alloc-size=128 ./...

# Focus on specific patterns
go vet -vettool=stackalloc -stackalloc.disable-patterns="new-allocation" ./...
```

#### 4. AI integration not working
```bash
# Check API key
echo $OPENAI_API_KEY

# Ensure DI mode is enabled
STACKALLOC_USE_DI=true go vet -vettool=stackalloc ./...

# Check network connectivity
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

### Performance Tips

1. **Large Projects**: Analyze in chunks rather than all at once
2. **CI/CD**: Use parallel jobs for different modules
3. **Caching**: Cache stackalloc binary in CI environments
4. **Filtering**: Use build tags to exclude test files and vendor code

### Getting Help

- Check the [GitHub Issues](https://github.com/harriteja/gostackallocator/issues)
- Review the [README.md](README.md) for basic setup
- See [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for technical details

## Best Practices

1. **Start Small**: Begin with a single package or file
2. **Gradual Adoption**: Slowly lower allocation thresholds
3. **Focus on Hot Paths**: Prioritize performance-critical code
4. **Regular Analysis**: Integrate into CI/CD pipeline
5. **Review AI Suggestions**: Don't blindly apply automatic fixes
6. **Team Training**: Ensure team understands allocation patterns
7. **Measure Impact**: Profile before and after optimizations 