#!/bin/bash

# stackalloc Quick Setup Script
# This script helps users quickly install and configure stackalloc

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
INSTALL_DIR="/usr/local/bin"
REPO_URL="https://github.com/harriteja/gostackallocator.git"
TEMP_DIR="/tmp/stackalloc-setup"
SETUP_IDE=false
SETUP_CI=false

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Quick setup script for stackalloc - Go static analysis for stack allocation optimization.

OPTIONS:
    --install-dir DIR      Installation directory (default: /usr/local/bin)
    --setup-ide           Setup IDE integration (VS Code tasks)
    --setup-ci            Setup CI/CD templates
    -h, --help            Show this help message

EXAMPLES:
    # Basic installation
    $0

    # Install with IDE and CI setup
    $0 --setup-ide --setup-ci

    # Install to custom directory
    $0 --install-dir ~/bin

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --setup-ide)
            SETUP_IDE=true
            shift
            ;;
        --setup-ci)
            SETUP_CI=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

print_status "Starting stackalloc setup..."

# Check prerequisites
print_status "Checking prerequisites..."

if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.20+ first."
    exit 1
fi

GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
MAJOR_VERSION=$(echo $GO_VERSION | cut -d. -f1)
MINOR_VERSION=$(echo $GO_VERSION | cut -d. -f2)

if [ "$MAJOR_VERSION" -lt 1 ] || ([ "$MAJOR_VERSION" -eq 1 ] && [ "$MINOR_VERSION" -lt 20 ]); then
    print_error "Go 1.20+ is required. Current version: $GO_VERSION"
    exit 1
fi

print_success "Go $GO_VERSION detected"

if ! command -v git &> /dev/null; then
    print_error "Git is not installed. Please install Git first."
    exit 1
fi

print_success "Git detected"

# Create installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
    print_status "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR" || {
        print_error "Failed to create installation directory. You may need sudo privileges."
        print_error "Try: sudo $0"
        exit 1
    }
fi

# Check write permissions
if [ ! -w "$INSTALL_DIR" ]; then
    print_error "No write permission to $INSTALL_DIR. You may need sudo privileges."
    print_error "Try: sudo $0"
    exit 1
fi

# Clean up any existing temp directory
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# Clone repository
print_status "Downloading stackalloc source code..."
git clone "$REPO_URL" "$TEMP_DIR" || {
    print_error "Failed to clone repository"
    exit 1
}

# Build stackalloc
print_status "Building stackalloc..."
cd "$TEMP_DIR"
go build -o stackalloc ./cmd || {
    print_error "Failed to build stackalloc"
    exit 1
}

# Install binary
print_status "Installing stackalloc to $INSTALL_DIR..."
cp stackalloc "$INSTALL_DIR/" || {
    print_error "Failed to install stackalloc"
    exit 1
}

print_success "stackalloc installed successfully!"

# Verify installation
if command -v stackalloc &> /dev/null; then
    print_success "stackalloc is available in PATH"
else
    print_warning "stackalloc is not in PATH. You may need to add $INSTALL_DIR to your PATH"
    echo "Add this to your shell profile (.bashrc, .zshrc, etc.):"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

# Copy helper scripts
print_status "Installing helper scripts..."
SCRIPTS_DIR="$INSTALL_DIR"
cp scripts/analyze-project.sh "$SCRIPTS_DIR/" 2>/dev/null || {
    print_warning "Could not copy helper scripts to $SCRIPTS_DIR"
    print_status "Helper scripts are available in $TEMP_DIR/scripts/"
}

# Setup IDE integration
if [ "$SETUP_IDE" = true ]; then
    print_status "Setting up IDE integration..."
    
    # VS Code setup
    if [ -d ".vscode" ] || [ -f ".vscode/settings.json" ]; then
        print_status "Setting up VS Code integration..."
        mkdir -p .vscode
        
        cat > .vscode/tasks.json << 'EOF'
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "stackalloc: Analyze Current File",
            "type": "shell",
            "command": "go",
            "args": ["vet", "-vettool=stackalloc", "${file}"],
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
            "args": ["vet", "-vettool=stackalloc", "./..."],
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
EOF
        print_success "VS Code tasks configured"
    fi
fi

# Setup CI/CD templates
if [ "$SETUP_CI" = true ]; then
    print_status "Setting up CI/CD templates..."
    
    # GitHub Actions
    mkdir -p .github/workflows
    cat > .github/workflows/stackalloc.yml << 'EOF'
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
      run: go vet -vettool=stackalloc ./... || true
EOF
    print_success "GitHub Actions workflow created"
    
    # Pre-commit hook template
    mkdir -p .git/hooks
    cat > .git/hooks/pre-commit.template << 'EOF'
#!/bin/bash
# stackalloc pre-commit hook
# Copy this to .git/hooks/pre-commit and make it executable

echo "Running stackalloc analysis on changed files..."

CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -z "$CHANGED_FILES" ]; then
    echo "No Go files changed."
    exit 0
fi

for file in $CHANGED_FILES; do
    echo "Analyzing $file..."
    go vet -vettool=stackalloc "./$file" || {
        echo "stackalloc found issues in $file"
        echo "Please review and fix before committing."
        exit 1
    }
done

echo "stackalloc analysis passed!"
EOF
    print_success "Pre-commit hook template created"
fi

# Create sample configuration
print_status "Creating sample configuration..."
cat > .stackalloc.example << 'EOF'
# stackalloc Configuration Example
# Copy this to .stackalloc.config and modify as needed

# Maximum allocation size to consider "small" (in bytes)
MAX_ALLOC_SIZE=32

# Patterns to disable (comma-separated)
# Options: new-allocation, pointer-escape
DISABLE_PATTERNS=""

# Enable metrics collection
METRICS_ENABLED=false

# OpenAI configuration (for AI-powered suggestions)
# OPENAI_API_KEY=your-api-key-here
OPENAI_MODEL=gpt-4
OPENAI_MAX_TOKENS=512
OPENAI_TEMPERATURE=0.2

# Enable dependency injection mode for advanced features
STACKALLOC_USE_DI=false
EOF

# Show usage examples
print_status "Setup complete! Here are some usage examples:"
echo ""
echo "Basic usage:"
echo "  go vet -vettool=stackalloc ./..."
echo ""
echo "Analyze specific package:"
echo "  go vet -vettool=stackalloc ./pkg/mypackage"
echo ""
echo "With custom threshold:"
echo "  go vet -vettool=stackalloc -max-alloc-size=64 ./..."
echo ""
echo "Using the helper script:"
echo "  analyze-project.sh --ai --autofix"
echo ""

if [ "$SETUP_IDE" = true ]; then
    echo "IDE Integration:"
    echo "  - VS Code: Use Ctrl+Shift+P -> 'Tasks: Run Task' -> 'stackalloc'"
    echo ""
fi

if [ "$SETUP_CI" = true ]; then
    echo "CI/CD:"
    echo "  - GitHub Actions workflow created in .github/workflows/stackalloc.yml"
    echo "  - Pre-commit hook template in .git/hooks/pre-commit.template"
    echo ""
fi

print_success "stackalloc is ready to use!"
print_status "For detailed usage instructions, see: https://github.com/harriteja/gostackallocator/blob/main/USAGE.md"

# Clean up
cd - > /dev/null
rm -rf "$TEMP_DIR"

exit 0 