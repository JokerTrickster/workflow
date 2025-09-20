#!/bin/bash

# Comprehensive Test Runner Script for Local Backend Server
# This script runs all tests and generates coverage reports

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
BENCHMARK_RESULTS_FILE="$PROJECT_ROOT/benchmark-results.txt"

# Test configuration
COVERAGE_THRESHOLD=90
TIMEOUT="10m"
PARALLEL_TESTS=4

# Create necessary directories
mkdir -p "$COVERAGE_DIR"
mkdir -p "$TEST_RESULTS_DIR"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}     Local Backend Server Test Suite${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Function to print step headers
print_step() {
    echo -e "${YELLOW}[STEP] $1${NC}"
    echo ""
}

# Function to print success messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    echo ""
}

# Function to print error messages
print_error() {
    echo -e "${RED}✗ $1${NC}"
    echo ""
}

# Function to print info messages
print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Change to project directory
cd "$PROJECT_ROOT"

print_step "Environment Setup"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3)
print_info "Using Go version: $GO_VERSION"

# Download dependencies
print_info "Downloading dependencies..."
go mod download
print_success "Dependencies downloaded"

print_step "Code Quality Checks"

# Run go fmt
print_info "Running go fmt..."
if ! go fmt ./...; then
    print_error "Code formatting issues found"
    exit 1
fi
print_success "Code formatting passed"

# Run go vet
print_info "Running go vet..."
if ! go vet ./...; then
    print_error "Static analysis issues found"
    exit 1
fi
print_success "Static analysis passed"

# Run golangci-lint if available
if command -v golangci-lint &> /dev/null; then
    print_info "Running golangci-lint..."
    if ! golangci-lint run; then
        print_error "Linting issues found"
        exit 1
    fi
    print_success "Linting passed"
else
    print_info "golangci-lint not found, skipping advanced linting"
fi

print_step "Unit Tests"

print_info "Running unit tests with coverage..."

# Run unit tests with coverage
if ! go test -v -timeout="$TIMEOUT" -race -coverprofile="$COVERAGE_DIR/unit.out" \
    -covermode=atomic \
    ./internal/domain/... \
    ./internal/usecase/... \
    ./internal/infrastructure/... \
    2>&1 | tee "$TEST_RESULTS_DIR/unit-tests.log"; then
    print_error "Unit tests failed"
    exit 1
fi

print_success "Unit tests passed"

# Generate unit test coverage report
go tool cover -html="$COVERAGE_DIR/unit.out" -o "$COVERAGE_DIR/unit-coverage.html"
UNIT_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/unit.out" | grep total | awk '{print $3}' | sed 's/%//')

print_info "Unit test coverage: ${UNIT_COVERAGE}%"

if (( $(echo "$UNIT_COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    print_error "Unit test coverage ($UNIT_COVERAGE%) is below threshold ($COVERAGE_THRESHOLD%)"
    exit 1
fi

print_success "Unit test coverage meets threshold"

print_step "Integration Tests"

print_info "Running integration tests..."

# Set up test environment variables
export DATABASE_DSN=":memory:"
export CLAUDE_API_KEY="sk-test-key-for-integration-testing"

# Run integration tests
if ! go test -v -timeout="$TIMEOUT" -race -coverprofile="$COVERAGE_DIR/integration.out" \
    -covermode=atomic \
    -tags=integration \
    ./tests/integration/... \
    2>&1 | tee "$TEST_RESULTS_DIR/integration-tests.log"; then
    print_error "Integration tests failed"
    exit 1
fi

print_success "Integration tests passed"

# Generate integration test coverage report
go tool cover -html="$COVERAGE_DIR/integration.out" -o "$COVERAGE_DIR/integration-coverage.html"
INTEGRATION_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/integration.out" | grep total | awk '{print $3}' | sed 's/%//')

print_info "Integration test coverage: ${INTEGRATION_COVERAGE}%"

print_step "End-to-End Tests"

print_info "Running end-to-end tests..."

# Run e2e tests
if ! go test -v -timeout="$TIMEOUT" -race \
    ./tests/e2e/... \
    2>&1 | tee "$TEST_RESULTS_DIR/e2e-tests.log"; then
    print_error "End-to-end tests failed"
    exit 1
fi

print_success "End-to-end tests passed"

print_step "Performance Benchmarks"

print_info "Running performance benchmarks..."

# Run benchmarks
go test -v -timeout="$TIMEOUT" -bench=. -benchmem -run=^$ \
    ./tests/benchmarks/... \
    2>&1 | tee "$BENCHMARK_RESULTS_FILE"

print_success "Performance benchmarks completed"

# Analyze benchmark results
print_info "Benchmark Results Summary:"
echo ""

# Extract key metrics from benchmark results
if [ -f "$BENCHMARK_RESULTS_FILE" ]; then
    echo "Database Operations:"
    grep "BenchmarkDatabaseOperations" "$BENCHMARK_RESULTS_FILE" | head -5
    echo ""
    
    echo "Message Processing:"
    grep "BenchmarkMessageProcessing" "$BENCHMARK_RESULTS_FILE" | head -5
    echo ""
    
    echo "Memory Usage:"
    grep "BenchmarkMemoryUsage" "$BENCHMARK_RESULTS_FILE" | head -3
    echo ""
fi

print_step "Combined Coverage Report"

print_info "Generating combined coverage report..."

# Combine coverage profiles
echo "mode: atomic" > "$COVERAGE_DIR/combined.out"
tail -n +2 "$COVERAGE_DIR/unit.out" >> "$COVERAGE_DIR/combined.out"
tail -n +2 "$COVERAGE_DIR/integration.out" >> "$COVERAGE_DIR/combined.out"

# Generate combined coverage report
go tool cover -html="$COVERAGE_DIR/combined.out" -o "$COVERAGE_DIR/combined-coverage.html"
COMBINED_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/combined.out" | grep total | awk '{print $3}' | sed 's/%//')

print_info "Combined test coverage: ${COMBINED_COVERAGE}%"

print_step "Test Results Summary"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}            TEST RESULTS SUMMARY${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

print_success "✓ Code quality checks passed"
print_success "✓ Unit tests passed (${UNIT_COVERAGE}% coverage)"
print_success "✓ Integration tests passed (${INTEGRATION_COVERAGE}% coverage)"
print_success "✓ End-to-end tests passed"
print_success "✓ Performance benchmarks completed"
print_success "✓ Combined coverage: ${COMBINED_COVERAGE}%"

echo ""
echo -e "${GREEN}All tests passed successfully!${NC}"
echo ""

print_info "Test artifacts generated:"
print_info "  • Unit test coverage: $COVERAGE_DIR/unit-coverage.html"
print_info "  • Integration test coverage: $COVERAGE_DIR/integration-coverage.html"
print_info "  • Combined coverage: $COVERAGE_DIR/combined-coverage.html"
print_info "  • Unit test log: $TEST_RESULTS_DIR/unit-tests.log"
print_info "  • Integration test log: $TEST_RESULTS_DIR/integration-tests.log"
print_info "  • E2E test log: $TEST_RESULTS_DIR/e2e-tests.log"
print_info "  • Benchmark results: $BENCHMARK_RESULTS_FILE"

echo ""

# Final coverage check
if (( $(echo "$COMBINED_COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    print_error "Combined coverage ($COMBINED_COVERAGE%) is below threshold ($COVERAGE_THRESHOLD%)"
    exit 1
fi

print_success "All coverage thresholds met!"

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}     Testing completed successfully! 🎉${NC}"
echo -e "${BLUE}================================================${NC}"

exit 0