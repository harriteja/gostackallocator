# stackalloc Client Usage Summary

## Quick Start for Clients

### Installation
```bash
# Clone and build
git clone https://github.com/harriteja/gostackallocator.git
cd gostackallocator
go build -o stackalloc ./cmd

# Optional: Add to PATH
sudo mv stackalloc /usr/local/bin/
```

### Basic Usage

#### 1. Entire Project Analysis
```bash
# Analyze entire monolith application
go vet -vettool=stackalloc ./...

# With custom threshold (stricter for performance-critical code)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./...

# Save results to file
go vet -vettool=stackalloc ./... 2> stackalloc-report.txt
```

#### 2. Folder-Level Analysis
```bash
# Analyze specific modules/packages
go vet -vettool=stackalloc ./internal/auth/...
go vet -vettool=stackalloc ./pkg/handlers/...

# Multiple packages
go vet -vettool=stackalloc ./internal/core/... ./pkg/algorithms/...

# Different thresholds for different modules
go vet -vettool=stackalloc -stackalloc.max-alloc-size=8 ./internal/performance-critical/...
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./internal/less-critical/...
```

#### 3. File-Level Analysis
```bash
# Single file
go vet -vettool=stackalloc ./internal/handlers/user.go

# Multiple specific files
go vet -vettool=stackalloc ./main.go ./internal/config/config.go

# Performance-critical files
go vet -vettool=stackalloc \
  ./internal/handlers/high-traffic.go \
  ./pkg/algorithms/sorting.go \
  ./internal/database/queries.go
```

### Helper Scripts

#### Automated Analysis
```bash
# Use the provided helper script
./scripts/analyze-project.sh -b /path/to/stackalloc

# With custom settings
./scripts/analyze-project.sh \
  -b /path/to/stackalloc \
  -s 16 \
  --metrics \
  -v

# AI-powered analysis (requires OPENAI_API_KEY)
export OPENAI_API_KEY="your-key"
./scripts/analyze-project.sh -b /path/to/stackalloc --ai --autofix
```

#### Quick Setup
```bash
# Install and configure for your project
./scripts/quick-setup.sh --setup-ide --setup-ci
```

### Configuration Options

#### Available Flags
- `-stackalloc.max-alloc-size=N` - Maximum bytes to consider 'small' allocation (default: 32)
- `-stackalloc.disable-patterns=P` - Comma-separated list of detectors to skip
- `-stackalloc.metrics-enabled` - Expose Prometheus metrics
- `-stackalloc.openai-api-key=KEY` - OpenAI API key for AI suggestions
- `-stackalloc.openai-model=MODEL` - OpenAI model to use (default: gpt-4)
- `-stackalloc.autofix` - Enable automatic code fixes (use with caution)

#### Example Configurations
```bash
# Strict analysis for high-performance code
go vet -vettool=stackalloc -stackalloc.max-alloc-size=8 ./...

# Disable specific patterns
go vet -vettool=stackalloc -stackalloc.disable-patterns="new-allocation" ./...

# Enable metrics collection
go vet -vettool=stackalloc -stackalloc.metrics-enabled ./...
```

### Real-World Examples

#### E-commerce Monolith
```bash
# Critical path analysis (payment, orders)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=16 ./internal/payment/... ./internal/orders/...

# High-traffic analysis (auth, cart)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./internal/auth/... ./internal/cart/...

# Background services (less critical)
go vet -vettool=stackalloc -stackalloc.max-alloc-size=64 ./cmd/worker/... ./internal/inventory/...
```

#### Microservices in Monorepo
```bash
# Analyze each service separately
for service in user-service order-service payment-service; do
    echo "Analyzing $service..."
    go vet -vettool=stackalloc ./services/$service/...
done
```

#### Performance-Critical Application
```bash
# Ultra-strict analysis for trading engine
go vet -vettool=stackalloc -stackalloc.max-alloc-size=8 ./engine/...

# Normal analysis for other components
go vet -vettool=stackalloc -stackalloc.max-alloc-size=32 ./risk/... ./reporting/...
```

### Integration

#### CI/CD Pipeline
```yaml
# GitHub Actions example
- name: Run stackalloc analysis
  run: |
    git clone https://github.com/harriteja/gostackallocator.git
    cd gostackallocator && go build -o stackalloc ./cmd
    go vet -vettool=./gostackallocator/stackalloc ./... || true
```

#### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')
for file in $CHANGED_FILES; do
    go vet -vettool=stackalloc "./$file" || exit 1
done
```

### What stackalloc Detects

1. **new(T) allocations** that could be stack-allocated
2. **Pointer escapes** that happen only once
3. **Object reuse patterns** to avoid false positives
4. **Small heap allocations** based on configurable thresholds

### Output Example
```
examples/sample.go:13:6: new(T) in return/assignment always allocates on heap; consider stack allocation
examples/sample.go:13:6: new(T) always allocates on heap; consider using stack allocation if object doesn't escape
```

### Best Practices

1. **Start Small**: Begin with a single package or file
2. **Gradual Adoption**: Slowly lower allocation thresholds
3. **Focus on Hot Paths**: Prioritize performance-critical code
4. **Regular Analysis**: Integrate into CI/CD pipeline
5. **Review Suggestions**: Don't blindly apply fixes
6. **Measure Impact**: Profile before and after optimizations

### Getting Help

- Full documentation: [USAGE.md](USAGE.md)
- Implementation details: [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)
- Issues: [GitHub Issues](https://github.com/harriteja/gostackallocator/issues) 