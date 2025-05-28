#!/bin/bash

# stackalloc Project Analysis Script
# This script helps analyze Go projects with different configurations

set -e

# Default values
PROJECT_ROOT="."
OUTPUT_DIR="./stackalloc-reports"
STACKALLOC_BINARY="stackalloc"
MAX_ALLOC_SIZE=32
ENABLE_AI=false
ENABLE_AUTOFIX=false
ENABLE_METRICS=false
EXCLUDE_VENDOR=true
EXCLUDE_TESTS=true
VERBOSE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Analyze Go projects with stackalloc for memory allocation optimization.

OPTIONS:
    -p, --project PATH          Project root directory (default: current directory)
    -o, --output DIR           Output directory for reports (default: ./stackalloc-reports)
    -s, --size SIZE            Maximum allocation size threshold (default: 32)
    -b, --binary PATH          Path to stackalloc binary (default: stackalloc)
    --ai                       Enable AI-powered suggestions (requires OPENAI_API_KEY)
    --autofix                  Enable automatic fix suggestions (requires --ai)
    --metrics                  Enable metrics collection
    --include-vendor           Include vendor directory in analysis
    --include-tests            Include test files in analysis
    -v, --verbose              Verbose output
    -h, --help                 Show this help message

EXAMPLES:
    # Basic analysis
    $0

    # Analyze specific project with AI
    $0 -p /path/to/project --ai

    # Strict analysis for performance-critical code
    $0 -s 16 --ai --autofix

    # Include everything (vendor, tests)
    $0 --include-vendor --include-tests

ENVIRONMENT VARIABLES:
    OPENAI_API_KEY            Required for AI-powered analysis
    STACKALLOC_USE_DI         Set to 'true' for dependency injection mode

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--project)
            PROJECT_ROOT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -s|--size)
            MAX_ALLOC_SIZE="$2"
            shift 2
            ;;
        -b|--binary)
            STACKALLOC_BINARY="$2"
            shift 2
            ;;
        --ai)
            ENABLE_AI=true
            shift
            ;;
        --autofix)
            ENABLE_AUTOFIX=true
            shift
            ;;
        --metrics)
            ENABLE_METRICS=true
            shift
            ;;
        --include-vendor)
            EXCLUDE_VENDOR=false
            shift
            ;;
        --include-tests)
            EXCLUDE_TESTS=false
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
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

# Validate inputs
if [ ! -d "$PROJECT_ROOT" ]; then
    print_error "Project directory does not exist: $PROJECT_ROOT"
    exit 1
fi

if ! command -v "$STACKALLOC_BINARY" &> /dev/null; then
    print_error "stackalloc binary not found: $STACKALLOC_BINARY"
    print_error "Please ensure stackalloc is installed and in PATH, or specify path with -b"
    exit 1
fi

if [ "$ENABLE_AI" = true ] && [ -z "$OPENAI_API_KEY" ]; then
    print_warning "AI analysis requested but OPENAI_API_KEY not set"
    print_warning "Continuing with basic analysis only"
    ENABLE_AI=false
    ENABLE_AUTOFIX=false
fi

if [ "$ENABLE_AUTOFIX" = true ] && [ "$ENABLE_AI" = false ]; then
    print_warning "Autofix requires AI to be enabled. Disabling autofix."
    ENABLE_AUTOFIX=false
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

print_status "Starting stackalloc analysis..."
print_status "Project: $PROJECT_ROOT"
print_status "Output: $OUTPUT_DIR"
print_status "Max allocation size: $MAX_ALLOC_SIZE bytes"

# Change to project directory
cd "$PROJECT_ROOT"

# Build analysis command
ANALYSIS_CMD="go vet -vettool=$STACKALLOC_BINARY"
ANALYSIS_CMD="$ANALYSIS_CMD -stackalloc.max-alloc-size=$MAX_ALLOC_SIZE"

if [ "$ENABLE_METRICS" = true ]; then
    ANALYSIS_CMD="$ANALYSIS_CMD -stackalloc.metrics-enabled"
fi

if [ "$ENABLE_AI" = true ]; then
    export STACKALLOC_USE_DI=true
    if [ "$ENABLE_AUTOFIX" = true ]; then
        ANALYSIS_CMD="$ANALYSIS_CMD -stackalloc.autofix"
    fi
fi

# Determine target packages
if [ "$EXCLUDE_VENDOR" = true ] && [ "$EXCLUDE_TESTS" = true ]; then
    # Find Go packages excluding vendor and test files
    PACKAGES=$(find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go" | xargs dirname | sort -u | tr '\n' ' ')
    if [ -z "$PACKAGES" ]; then
        PACKAGES="./..."
    fi
elif [ "$EXCLUDE_VENDOR" = true ]; then
    # Exclude only vendor
    PACKAGES=$(find . -name "*.go" -not -path "./vendor/*" | xargs dirname | sort -u | tr '\n' ' ')
    if [ -z "$PACKAGES" ]; then
        PACKAGES="./..."
    fi
elif [ "$EXCLUDE_TESTS" = true ]; then
    # Exclude only test files
    PACKAGES=$(find . -name "*.go" -not -name "*_test.go" | xargs dirname | sort -u | tr '\n' ' ')
    if [ -z "$PACKAGES" ]; then
        PACKAGES="./..."
    fi
else
    # Include everything
    PACKAGES="./..."
fi

# Run analysis
print_status "Running analysis on packages: $PACKAGES"

if [ "$VERBOSE" = true ]; then
    print_status "Command: $ANALYSIS_CMD $PACKAGES"
fi

# Basic analysis
BASIC_REPORT="$OUTPUT_DIR/basic_analysis_$TIMESTAMP.txt"
print_status "Running basic analysis..."

if [ "$ENABLE_AI" = false ]; then
    $ANALYSIS_CMD $PACKAGES > "$BASIC_REPORT" 2>&1 || true
else
    # AI-enhanced analysis
    AI_REPORT="$OUTPUT_DIR/ai_analysis_$TIMESTAMP.txt"
    print_status "Running AI-enhanced analysis..."
    $ANALYSIS_CMD $PACKAGES > "$AI_REPORT" 2>&1 || true
    
    # Also run basic analysis for comparison
    BASIC_CMD="go vet -vettool=$STACKALLOC_BINARY -stackalloc.max-alloc-size=$MAX_ALLOC_SIZE"
    if [ "$ENABLE_METRICS" = true ]; then
        BASIC_CMD="$BASIC_CMD -stackalloc.metrics-enabled"
    fi
    $BASIC_CMD $PACKAGES > "$BASIC_REPORT" 2>&1 || true
fi

# Generate summary
SUMMARY_REPORT="$OUTPUT_DIR/summary_$TIMESTAMP.txt"
print_status "Generating summary report..."

cat > "$SUMMARY_REPORT" << EOF
stackalloc Analysis Summary
===========================

Timestamp: $(date)
Project: $PROJECT_ROOT
Max Allocation Size: $MAX_ALLOC_SIZE bytes
AI Analysis: $ENABLE_AI
Auto-fix: $ENABLE_AUTOFIX
Metrics: $ENABLE_METRICS
Excluded Vendor: $EXCLUDE_VENDOR
Excluded Tests: $EXCLUDE_TESTS

Reports Generated:
EOF

if [ -f "$BASIC_REPORT" ]; then
    BASIC_ISSUES=$(grep -c "\.go:" "$BASIC_REPORT" 2>/dev/null || echo "0")
    echo "- Basic Analysis: $BASIC_REPORT ($BASIC_ISSUES issues found)" >> "$SUMMARY_REPORT"
fi

if [ -f "$AI_REPORT" ]; then
    AI_ISSUES=$(grep -c "\.go:" "$AI_REPORT" 2>/dev/null || echo "0")
    echo "- AI Analysis: $AI_REPORT ($AI_ISSUES issues found)" >> "$SUMMARY_REPORT"
fi

# Add top issues to summary
echo "" >> "$SUMMARY_REPORT"
echo "Top Issues Found:" >> "$SUMMARY_REPORT"
echo "=================" >> "$SUMMARY_REPORT"

if [ -f "$BASIC_REPORT" ]; then
    head -20 "$BASIC_REPORT" >> "$SUMMARY_REPORT" 2>/dev/null || true
fi

print_success "Analysis complete!"
print_status "Reports saved to: $OUTPUT_DIR"

# Show summary
if [ "$VERBOSE" = true ]; then
    echo ""
    cat "$SUMMARY_REPORT"
fi

# List generated files
echo ""
print_status "Generated files:"
ls -la "$OUTPUT_DIR"/*_$TIMESTAMP.* 2>/dev/null || true

# Provide next steps
echo ""
print_status "Next steps:"
echo "1. Review the analysis reports in $OUTPUT_DIR"
echo "2. Focus on high-impact issues first"
echo "3. Use AI suggestions to guide optimizations"
if [ "$ENABLE_AI" = false ]; then
    echo "4. Consider running with --ai for enhanced suggestions"
fi
echo "5. Re-run analysis after making changes to measure improvement"

exit 0 