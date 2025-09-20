# Testing Suite Documentation

This directory contains the comprehensive testing suite for the Local Backend Server. The test suite covers all application layers with 90%+ code coverage goal and includes various types of tests to ensure system reliability and performance.

## Test Structure

```
tests/
├── integration/          # Integration tests for external services
│   ├── database_integration_test.go     # SQLite database integration tests
│   └── claude_integration_test.go       # Claude API integration tests
├── e2e/                  # End-to-end tests
│   └── workflow_test.go              # Complete workflow tests
├── benchmarks/           # Performance benchmarks
│   └── performance_test.go          # Critical path performance tests
└── README.md            # This documentation
```

Additionally, unit tests are located alongside the source code:
- `internal/domain/*_test.go` - Domain layer unit tests
- `internal/infrastructure/*_test.go` - Infrastructure layer unit tests  
- `internal/usecase/*_test.go` - Application layer unit tests

## Test Categories

### 1. Unit Tests
**Location**: `internal/*/` (alongside source files)  
**Purpose**: Test individual components in isolation  
**Coverage**: Business logic, entity validation, service implementations

Key test files:
- `internal/domain/entities_test.go` - Domain entity behavior
- `internal/infrastructure/config_validator_test.go` - Configuration validation
- `internal/infrastructure/error_handler_test.go` - Error handling logic
- `internal/infrastructure/logger_test.go` - Logging functionality
- `internal/usecase/message_processor_test.go` - Message processing logic
- `internal/usecase/request_service_test.go` - Request service operations

### 2. Integration Tests
**Location**: `tests/integration/`  
**Purpose**: Test interactions with external services  
**Dependencies**: SQLite database, mock Claude API server

#### Database Integration Tests (`database_integration_test.go`)
Tests SQLite repository operations with actual database:
- Message CRUD operations with complex payloads
- Request lifecycle management
- Processing context operations with conversation history
- Transaction handling and rollback scenarios
- Concurrent access patterns
- Data persistence and retrieval accuracy
- Constraint validation and error handling

#### Claude API Integration Tests (`claude_integration_test.go`)
Tests Claude service integration with mock API server:
- Request/response processing with various payloads
- Context management for conversations
- Error handling for API failures, rate limiting, timeouts
- Request validation and format verification
- Concurrent request handling
- Large payload processing
- Special character and encoding handling

### 3. End-to-End Tests
**Location**: `tests/e2e/`  
**Purpose**: Test complete workflows from message ingestion to completion  
**Scope**: Full system integration with all components

#### Complete Workflow Tests (`workflow_test.go`)
- Single work request processing from start to finish
- Conversational workflows with context preservation
- Cancellation workflows and request lifecycle management
- Orchestrator startup, operation, and shutdown
- Error handling across system boundaries
- Concurrent processing scenarios
- Data persistence throughout workflows

### 4. Performance Benchmarks
**Location**: `tests/benchmarks/`  
**Purpose**: Measure system performance and identify bottlenecks  
**Targets**: 
- Message processing < 100ms (excluding Claude API)
- Database operations < 10ms
- Memory usage < 50MB during normal operation

#### Performance Tests (`performance_test.go`)
- Entity creation performance (Message, Request, ProcessingContext)
- Database operation benchmarks (save, retrieve, update, delete)
- End-to-end message processing performance
- Concurrent processing throughput
- Memory usage profiling
- Large payload handling
- Context operation performance
- Cleanup operation efficiency

## Running Tests

### Quick Test Execution

```bash
# Run all unit tests
go test ./internal/...

# Run integration tests
go test ./tests/integration/...

# Run end-to-end tests  
go test ./tests/e2e/...

# Run performance benchmarks
go test -bench=. ./tests/benchmarks/...
```

### Comprehensive Test Suite

Use the automated test runner for complete testing:

```bash
# Run complete test suite with coverage
./scripts/run-tests.sh
```

The test runner performs:
1. Environment validation
2. Code quality checks (fmt, vet, lint)
3. Unit tests with coverage reporting
4. Integration tests with external service mocking
5. End-to-end workflow testing
6. Performance benchmarking
7. Combined coverage analysis

### Coverage Requirements

- **Overall Target**: 90%+ code coverage
- **Unit Tests**: Focus on business logic correctness
- **Integration Tests**: Verify external service interactions
- **E2E Tests**: Validate complete user scenarios

### Test Environment Setup

#### Required Environment Variables
```bash
# For integration tests
export DATABASE_DSN=":memory:"              # Use in-memory SQLite
export CLAUDE_API_KEY="sk-test-key"         # Test API key

# Optional for specific test scenarios
export LOG_LEVEL="debug"                    # Increase logging detail
export TEST_TIMEOUT="10m"                   # Adjust test timeout
```

#### Dependencies
- Go 1.21+ 
- SQLite (for integration tests)
- testify/assert, testify/mock (included in go.mod)

## Test Patterns and Best Practices

### Unit Test Patterns

1. **Table-Driven Tests**: Used extensively for validation scenarios
```go
tests := []struct {
    name        string
    input       InputType
    expected    OutputType
    expectError bool
}{
    {"valid input", validInput, expectedOutput, false},
    {"invalid input", invalidInput, nil, true},
}
```

2. **Mock Objects**: Comprehensive mocking for external dependencies
```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, entity *Entity) error {
    args := m.Called(ctx, entity)
    return args.Error(0)
}
```

3. **Test Fixtures**: Reusable test data creation
```go
func createTestMessage() *domain.Message {
    return domain.NewMessage("test-id", domain.MessageTypeWorkRequest, testPayload)
}
```

### Integration Test Patterns

1. **Test Database Setup**: Temporary databases for isolation
```go
func setupTestDatabase(t *testing.T) (*Repository, func()) {
    // Create temp DB
    // Return repository and cleanup function
}
```

2. **Mock External Services**: HTTP test servers for API testing
```go
func createMockClaudeServer(t *testing.T) *httptest.Server {
    // Mock API endpoints with realistic responses
}
```

3. **Resource Cleanup**: Proper cleanup in defer statements
```go
defer func() {
    repo.Close()
    os.RemoveAll(tempDir)
}()
```

### End-to-End Test Patterns

1. **Complete System Setup**: All components initialized
2. **Realistic Scenarios**: Actual user workflow simulation
3. **State Verification**: Database and memory state checks
4. **Error Scenario Testing**: Failure handling validation

### Performance Test Patterns

1. **Baseline Measurements**: Consistent performance targets
2. **Memory Profiling**: Resource usage monitoring
3. **Concurrent Load Testing**: Multi-goroutine scenarios
4. **Scalability Testing**: Large data volume handling

## Troubleshooting Tests

### Common Issues

1. **Database Lock Errors**: Ensure proper cleanup of test databases
2. **Context Timeouts**: Increase timeout for slow tests
3. **Port Conflicts**: Use dynamic port allocation for test servers
4. **Race Conditions**: Enable race detection with `-race` flag
5. **Mock Configuration**: Verify mock expectations match test scenarios

### Debugging Tips

1. **Verbose Output**: Use `go test -v` for detailed test logs
2. **Single Test Execution**: `go test -run TestSpecificFunction`
3. **Coverage Analysis**: Review coverage reports for missed code paths
4. **Benchmark Comparison**: Track performance changes over time
5. **Log Analysis**: Check test logs in `test-results/` directory

### Test Maintenance

1. **Regular Updates**: Keep tests current with code changes
2. **Mock Maintenance**: Update mocks when interfaces change
3. **Performance Baselines**: Update targets as system evolves
4. **Test Data**: Refresh test fixtures periodically
5. **Dependency Updates**: Keep testing libraries current

## Continuous Integration

The test suite is designed for CI/CD integration:

- **Fast Feedback**: Unit tests run quickly for rapid iteration
- **Comprehensive Coverage**: Integration and E2E tests for release validation
- **Performance Monitoring**: Benchmarks track performance regressions
- **Artifact Generation**: Coverage reports and test logs for analysis

### CI Configuration Example

```yaml
test:
  script:
    - ./scripts/run-tests.sh
  coverage: '/total:.*(\d+.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/coverage.xml
    paths:
      - coverage/
      - test-results/
      - benchmark-results.txt
```

This comprehensive testing approach ensures the Local Backend Server maintains high quality, performance, and reliability standards across all operational scenarios.