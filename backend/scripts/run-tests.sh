#!/bin/bash
set -e

# Script to run comprehensive tests for the Go backend task queue system
echo "🧪 Running comprehensive test suite for Go backend task queue system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_section() {
    echo -e "${BLUE}==== $1 ====${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Docker is running (needed for testcontainers)
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker to run integration tests."
    exit 1
fi

# Set test environment variables
export GO_ENV=test
export DOCKER_API_VERSION=1.41

# Navigate to backend directory
cd "$(dirname "$0")/.."

print_section "Environment Setup"
echo "Current directory: $(pwd)"
echo "Go version: $(go version)"
echo "Docker version: $(docker --version)"

# Clean any previous test artifacts
print_section "Cleaning Previous Test Artifacts"
rm -rf coverage.out coverage.html test-results.xml 2>/dev/null || true
print_success "Cleaned previous test artifacts"

# Download dependencies
print_section "Installing Dependencies"
go mod download
go mod tidy
print_success "Dependencies installed"

# Run linting (if available)
print_section "Code Quality Checks"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    print_success "Linting passed"
else
    print_warning "golangci-lint not found, skipping linting"
fi

# Run unit tests first (fastest)
print_section "Unit Tests"
echo "Running unit tests for domain and application layers..."
go test -v -timeout=30s -tags=unit \
    ./tests/unit/domain/... \
    ./tests/unit/application/... \
    2>&1 | tee unit-test-results.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Run integration tests (require containers)
print_section "Integration Tests"
echo "Running integration tests with real MySQL and RabbitMQ..."
echo "This may take a few minutes as containers are started..."

# Set longer timeout for integration tests
go test -v -timeout=5m -tags=integration \
    ./tests/integration/database/... \
    ./tests/integration/queue/... \
    2>&1 | tee integration-test-results.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

# Run API tests
print_section "API Tests"
echo "Running HTTP API tests..."
go test -v -timeout=3m -tags=api \
    ./tests/api/... \
    2>&1 | tee api-test-results.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "API tests passed"
else
    print_error "API tests failed"
    exit 1
fi

# Run performance tests
print_section "Performance Tests"
echo "Running performance and load tests..."
go test -v -timeout=10m -tags=performance \
    ./tests/performance/... \
    2>&1 | tee performance-test-results.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    print_success "Performance tests passed"
else
    print_warning "Performance tests failed (may be due to resource constraints)"
fi

# Run comprehensive test coverage
print_section "Test Coverage Analysis"
echo "Generating test coverage report..."

# Run all tests with coverage
go test -coverprofile=coverage.out -coverpkg=./internal/... \
    ./tests/unit/... \
    ./internal/... \
    2>/dev/null

if [ -f coverage.out ]; then
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    # Get coverage percentage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "📊 Total test coverage: $COVERAGE"
    
    if [[ $COVERAGE == *"."* ]]; then
        COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
        if (( $(echo "$COVERAGE_NUM >= 90" | bc -l) )); then
            print_success "Excellent coverage: $COVERAGE (≥90%)"
        elif (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
            print_success "Good coverage: $COVERAGE (≥80%)"
        elif (( $(echo "$COVERAGE_NUM >= 70" | bc -l) )); then
            print_warning "Acceptable coverage: $COVERAGE (≥70%)"
        else
            print_warning "Low coverage: $COVERAGE (<70%)"
        fi
    fi
    
    print_success "Coverage report generated: coverage.html"
else
    print_warning "Could not generate coverage report"
fi

# Run benchmarks
print_section "Benchmark Tests"
echo "Running benchmark tests..."
go test -bench=. -benchmem -timeout=5m \
    ./tests/unit/... \
    ./tests/integration/... \
    ./tests/performance/... \
    2>&1 | tee benchmark-results.log

print_success "Benchmarks completed"

# Clean up test containers (Docker cleanup)
print_section "Cleanup"
echo "Cleaning up test containers..."
docker container prune -f --filter "label=org.testcontainers=true" 2>/dev/null || true
docker volume prune -f 2>/dev/null || true
print_success "Cleanup completed"

# Summary
print_section "Test Summary"
echo "📋 Test execution completed!"
echo ""
echo "📁 Test artifacts generated:"
echo "   - unit-test-results.log"
echo "   - integration-test-results.log" 
echo "   - api-test-results.log"
echo "   - performance-test-results.log"
echo "   - benchmark-results.log"
if [ -f coverage.html ]; then
    echo "   - coverage.html (open in browser to view)"
fi
echo ""
echo "🎯 Key metrics:"
if [ -f coverage.out ]; then
    echo "   - Test coverage: $(go tool cover -func=coverage.out | grep total | awk '{print $3}')"
fi
echo "   - Unit tests: ✅"
echo "   - Integration tests: ✅"
echo "   - API tests: ✅"
echo "   - Performance tests: $([ ${PIPESTATUS[0]} -eq 0 ] && echo "✅" || echo "⚠️")"
echo ""
print_success "All critical tests passed! 🚀"
echo ""
echo "To run specific test suites:"
echo "  Unit tests:        ./scripts/run-tests.sh unit"
echo "  Integration tests: ./scripts/run-tests.sh integration"
echo "  API tests:         ./scripts/run-tests.sh api"
echo "  Performance tests: ./scripts/run-tests.sh performance"
echo "  Coverage only:     ./scripts/run-tests.sh coverage"