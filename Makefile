.PHONY: build test clean install run-examples help

# Build the stackalloc binary
build:
	go build -o stackalloc ./cmd

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f stackalloc

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the tool globally
install:
	go install ./cmd

# Run the analyzer on example code
run-examples: build
	go vet -vettool=./stackalloc ./examples/

# Run with AI suggestions (requires OPENAI_API_KEY)
run-examples-ai: build
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	STACKALLOC_USE_DI=true go vet -vettool=./stackalloc ./examples/

# Run with AI suggestions and automatic fixes (requires OPENAI_API_KEY)
run-examples-autofix: build
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	STACKALLOC_USE_DI=true go vet -vettool=./stackalloc -autofix ./examples/

# Run with metrics enabled
run-examples-metrics: build
	STACKALLOC_USE_DI=true go vet -vettool=./stackalloc -metrics-enabled ./examples/

# Use the helper script for comprehensive analysis
analyze-project: build
	./scripts/analyze-project.sh -b ./stackalloc

# Use the helper script with AI
analyze-project-ai: build
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	./scripts/analyze-project.sh -b ./stackalloc --ai

# Setup for client projects (run this in client's project directory)
setup-client:
	@echo "Setting up stackalloc for client project..."
	@if [ ! -f "go.mod" ]; then \
		echo "Error: This doesn't appear to be a Go project (no go.mod found)"; \
		exit 1; \
	fi
	./scripts/quick-setup.sh --setup-ide --setup-ci

# Make scripts executable
setup-scripts:
	chmod +x scripts/*.sh

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Check for security issues
security:
	gosec ./...

# Run all checks (test, lint, security)
check: test lint security

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the stackalloc binary"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  install         - Install the tool globally"
	@echo "  run-examples    - Run analyzer on example code"
	@echo "  run-examples-ai - Run with AI suggestions (requires OPENAI_API_KEY)"
	@echo "  run-examples-autofix - Run with AI suggestions and automatic fixes"
	@echo "  run-examples-metrics - Run with metrics enabled"
	@echo "  analyze-project - Use helper script for comprehensive analysis"
	@echo "  analyze-project-ai - Use helper script with AI analysis"
	@echo "  setup-client    - Setup stackalloc for client project"
	@echo "  setup-scripts   - Make scripts executable"
	@echo "  lint            - Lint the code"
	@echo "  fmt             - Format the code"
	@echo "  security        - Check for security issues"
	@echo "  check           - Run all checks (test, lint, security)"
	@echo "  help            - Show this help message" 