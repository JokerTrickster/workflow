#!/bin/bash

# Integration Test Runner for Task History System
# This script runs comprehensive integration tests to validate the complete system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$BACKEND_DIR/frontend"
TEST_TIMEOUT="5m"
VERBOSE=${VERBOSE:-false}

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
TEST_RESULTS=()

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to record test results
record_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ "$status" = "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "$test_name - $status ($duration)"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "$test_name - $status ($duration)"
    fi

    TEST_RESULTS+=("$test_name: $status ($duration)")
}

# Function to run Go tests with timing
run_go_test() {
    local test_name="$1"
    local test_path="$2"
    local test_pattern="${3:-.*}"

    log "Running $test_name..."

    cd "$BACKEND_DIR"

    local start_time=$(date +%s)
    local temp_file=$(mktemp)

    if [ "$VERBOSE" = "true" ]; then
        go test -v -timeout "$TEST_TIMEOUT" -run "$test_pattern" "$test_path" 2>&1 | tee "$temp_file"
        local exit_code=${PIPESTATUS[0]}
    else
        go test -timeout "$TEST_TIMEOUT" -run "$test_pattern" "$test_path" > "$temp_file" 2>&1
        local exit_code=$?
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local duration_str="${duration}s"

    if [ $exit_code -eq 0 ]; then
        record_test_result "$test_name" "PASS" "$duration_str"
    else
        record_test_result "$test_name" "FAIL" "$duration_str"
        log_error "Test output:"
        cat "$temp_file"
    fi

    rm -f "$temp_file"
    return $exit_code
}

# Function to run npm tests with timing
run_npm_test() {
    local test_name="$1"
    local test_pattern="$2"

    log "Running $test_name..."

    cd "$FRONTEND_DIR"

    local start_time=$(date +%s)
    local temp_file=$(mktemp)

    if [ "$VERBOSE" = "true" ]; then
        npm test -- --testPathPattern="$test_pattern" --verbose 2>&1 | tee "$temp_file"
        local exit_code=${PIPESTATUS[0]}
    else
        npm test -- --testPathPattern="$test_pattern" --silent > "$temp_file" 2>&1
        local exit_code=$?
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local duration_str="${duration}s"

    if [ $exit_code -eq 0 ]; then
        record_test_result "$test_name" "PASS" "$duration_str"
    else
        record_test_result "$test_name" "FAIL" "$duration_str"
        log_error "Test output:"
        cat "$temp_file"
    fi

    rm -f "$temp_file"
    return $exit_code
}

# Function to check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    # Check Go
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed"
        exit 1
    fi

    # Check Node.js
    if ! command -v npm >/dev/null 2>&1; then
        log_error "npm is not installed"
        exit 1
    fi

    # Check backend dependencies
    cd "$BACKEND_DIR"
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found in backend directory"
        exit 1
    fi

    # Check frontend dependencies
    cd "$FRONTEND_DIR"
    if [ ! -f "package.json" ]; then
        log_error "package.json not found in frontend directory"
        exit 1
    fi

    if [ ! -d "node_modules" ]; then
        log "Installing frontend dependencies..."
        npm install
    fi

    log_success "Prerequisites check passed"
}

# Function to prepare test environment
prepare_test_environment() {
    log "Preparing test environment..."

    # Set environment variables for testing
    export NODE_ENV=test
    export GO_ENV=test

    # Clean up any previous test artifacts
    cd "$BACKEND_DIR"
    find . -name "*.test" -delete 2>/dev/null || true
    find . -name "coverage.out" -delete 2>/dev/null || true

    cd "$FRONTEND_DIR"
    rm -rf coverage/ 2>/dev/null || true

    log_success "Test environment prepared"
}

# Function to run backend integration tests
run_backend_tests() {
    log "=== Backend Integration Tests ==="

    # Unit tests for components
    run_go_test "Backend: Task History Handler Tests" "./internal/handlers" "TestTaskHistoryHandler"
    run_go_test "Backend: Atomic Service Tests" "./internal/services" "TestAtomicQueueService"

    # Integration tests
    run_go_test "Backend: Integration Tests" "./tests/integration" "TestE2E|TestAtomicTransactions|TestAPIPagination"
    run_go_test "Backend: Error Scenario Tests" "./tests/integration" "TestErrorScenarios"
    run_go_test "Backend: Concurrent Operations" "./tests/integration" "TestConcurrentOperations"

    # End-to-end tests
    run_go_test "Backend: Complete Workflow E2E" "./tests/e2e" "TestE2E_CompleteWorkflow"
    run_go_test "Backend: Multiple Tasks E2E" "./tests/e2e" "TestE2E_MultipleTasksWorkflow"
    run_go_test "Backend: Error Scenarios E2E" "./tests/e2e" "TestE2E_ErrorScenarios"

    # Performance tests
    run_go_test "Backend: Performance Tests" "./tests/performance" "TestAPI_ResponseTimeRequirements"
    run_go_test "Backend: Concurrent Performance" "./tests/performance" "TestAPI_ConcurrentRequestPerformance"
}

# Function to run frontend integration tests
run_frontend_tests() {
    log "=== Frontend Integration Tests ==="

    # Component integration tests
    run_npm_test "Frontend: TaskHistory Integration" "TaskHistoryIntegration.test"

    # API integration tests (if they exist)
    if [ -f "$FRONTEND_DIR/src/__tests__/integration/ApiIntegration.test.tsx" ]; then
        run_npm_test "Frontend: API Integration" "ApiIntegration.test"
    fi

    # Error handling tests
    run_npm_test "Frontend: Error Handling" "TaskHistoryIntegration.test.*Error"

    # Performance tests
    run_npm_test "Frontend: Performance" "TaskHistoryIntegration.test.*Performance"
}

# Function to run load tests (optional)
run_load_tests() {
    log "=== Load Testing ==="

    # Only run load tests if explicitly requested
    if [ "${RUN_LOAD_TESTS}" = "true" ]; then
        run_go_test "Backend: Stress Tests" "./tests/performance" "TestAPI_StressTest"
        run_go_test "Backend: Memory Usage Tests" "./tests/performance" "TestAPI_MemoryUsagePerformance"
    else
        log_warning "Load tests skipped (set RUN_LOAD_TESTS=true to enable)"
    fi
}

# Function to validate system integration
validate_system_integration() {
    log "=== System Integration Validation ==="

    # Run comprehensive end-to-end validation
    run_go_test "System: Complete Integration" "./tests/e2e" "TestE2E.*"

    # Validate performance requirements
    run_go_test "System: Performance Requirements" "./tests/performance" "TestAPI_ResponseTimeRequirements"

    # Validate error handling throughout system
    run_go_test "System: Error Handling" "./tests/integration" "TestErrorScenarios"
}

# Function to generate test report
generate_test_report() {
    log "=== Test Results Summary ==="

    echo
    echo "📊 Test Execution Summary:"
    echo "  Total Tests: $TOTAL_TESTS"
    echo "  Passed: $PASSED_TESTS"
    echo "  Failed: $FAILED_TESTS"

    if [ $FAILED_TESTS -eq 0 ]; then
        echo
        log_success "All tests passed! 🎉"
        echo
        echo "✅ Integration test validation complete:"
        echo "  • End-to-end workflow: VALIDATED"
        echo "  • Atomic transactions: VALIDATED"
        echo "  • API pagination: VALIDATED"
        echo "  • Error scenarios: VALIDATED"
        echo "  • Performance requirements: VALIDATED"
        echo "  • Frontend integration: VALIDATED"
        echo "  • Concurrent operations: VALIDATED"
    else
        echo
        log_error "Some tests failed!"
        echo
        echo "❌ Failed tests:"
        for result in "${TEST_RESULTS[@]}"; do
            if [[ $result == *"FAIL"* ]]; then
                echo "  • $result"
            fi
        done
    fi

    echo
    echo "📈 Detailed Results:"
    for result in "${TEST_RESULTS[@]}"; do
        echo "  $result"
    done
}

# Function to cleanup
cleanup() {
    log "Cleaning up test environment..."

    # Reset environment variables
    unset NODE_ENV
    unset GO_ENV

    # Clean up temporary files
    find /tmp -name "tmp.*" -user "$(whoami)" -delete 2>/dev/null || true

    log_success "Cleanup complete"
}

# Main execution
main() {
    echo "🚀 Task History Integration Test Suite"
    echo "======================================"
    echo

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --load-tests)
                RUN_LOAD_TESTS=true
                shift
                ;;
            --backend-only)
                BACKEND_ONLY=true
                shift
                ;;
            --frontend-only)
                FRONTEND_ONLY=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  -v, --verbose      Enable verbose output"
                echo "  --load-tests       Run load and stress tests"
                echo "  --backend-only     Run only backend tests"
                echo "  --frontend-only    Run only frontend tests"
                echo "  -h, --help         Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Trap cleanup on exit
    trap cleanup EXIT

    # Start time
    START_TIME=$(date +%s)

    # Run test suite
    check_prerequisites
    prepare_test_environment

    if [ "${FRONTEND_ONLY}" != "true" ]; then
        run_backend_tests
        validate_system_integration
        run_load_tests
    fi

    if [ "${BACKEND_ONLY}" != "true" ]; then
        run_frontend_tests
    fi

    # End time and duration
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    echo
    log "Test suite completed in ${DURATION}s"

    generate_test_report

    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"