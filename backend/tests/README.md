# Comprehensive Testing Suite for Go Backend Task Queue System

This directory contains a complete testing suite for the Go backend task queue system, implementing comprehensive quality assurance with 90%+ test coverage targets.

## 🎯 Testing Strategy Overview

### Test Categories

#### 1. **Unit Tests** (`/tests/unit/`)
Tests individual components in isolation with mocked dependencies.

**Coverage Areas:**
- **Domain Layer Tests** (`/tests/unit/domain/`)
  - Task entity business logic and state transitions
  - Value objects validation (TaskID, TaskStatus, UserID, etc.)
  - Domain services (TaskValidationService, TaskLifecycleService)
  - Business rule enforcement and error handling

- **Application Layer Tests** (`/tests/unit/application/`)
  - Use case implementations with mocked dependencies
  - Authorization service logic
  - Event service functionality
  - DTO validation and conversion
  - Error handling and edge cases

#### 2. **Integration Tests** (`/tests/integration/`)
Tests component interactions with real external dependencies.

**Coverage Areas:**
- **Database Integration** (`/tests/integration/database/`)
  - MySQL repository operations (CRUD, filtering, pagination)
  - Transaction handling and rollback scenarios
  - Connection pooling and error recovery
  - Database schema validation

- **Queue Integration** (`/tests/integration/queue/`)
  - RabbitMQ message publishing and consumption
  - Connection recovery and health checks
  - Message format validation
  - Queue statistics and monitoring

- **System Integration** (`/tests/integration/system/`)
  - End-to-end task lifecycle (create → process → complete)
  - Cross-component communication
  - Configuration loading and validation
  - Container dependency injection

#### 3. **API Tests** (`/tests/api/`)
Tests HTTP endpoints and API contracts.

**Coverage Areas:**
- All REST endpoints (POST, GET, PUT, DELETE)
- Request/response validation
- Authentication and authorization
- Error responses and status codes
- Middleware functionality (logging, CORS, recovery)

#### 4. **Performance Tests** (`/tests/performance/`)
Tests system performance under load and stress conditions.

**Coverage Areas:**
- Concurrent operations (multiple users, simultaneous requests)
- Database connection pooling under load
- Queue message publishing throughput
- API response times under concurrent load
- Memory usage and resource leak detection

## 🛠️ Testing Infrastructure

### Test Utilities (`/tests/testutils/`)

**TestEnvironment Setup:**
- Docker testcontainers for MySQL and RabbitMQ
- Automated database schema creation
- Test data factories and helpers
- Environment cleanup and isolation

**Key Components:**
- `TestEnvironment`: Complete test environment with containers
- `CreateTestTask()`: Test data factory for tasks
- `AssertTaskEqual()`: Custom assertion helpers
- `BenchmarkConfig`: Performance test configuration
- `MockTime`: Controllable time source for testing

### Testing Framework Stack

- **Core Testing**: Go standard `testing` package with table-driven tests
- **Assertions**: `github.com/stretchr/testify` for rich assertions and mocking
- **Test Containers**: Real MySQL and RabbitMQ instances via Docker
- **Benchmarking**: Go benchmark tests for performance validation
- **Coverage**: `go test -cover` for coverage reporting

## 🚀 Running Tests

### Quick Start
```bash
# Run all tests
./scripts/run-tests.sh

# Run specific test suites
./scripts/run-tests.sh unit
./scripts/run-tests.sh integration  
./scripts/run-tests.sh api
./scripts/run-tests.sh performance
./scripts/run-tests.sh coverage
```

### Manual Test Execution

```bash
# Unit tests (fast, no external dependencies)
go test -v ./tests/unit/... -timeout=30s

# Integration tests (requires Docker)
go test -v ./tests/integration/... -timeout=5m

# API tests
go test -v ./tests/api/... -timeout=3m

# Performance tests
go test -v ./tests/performance/... -timeout=10m

# Coverage analysis
go test -coverprofile=coverage.out ./tests/unit/... ./internal/...
go tool cover -html=coverage.out -o coverage.html
```

### Benchmark Tests
```bash
# Run all benchmarks
go test -bench=. -benchmem ./tests/...

# Specific benchmark categories
go test -bench=BenchmarkTaskRepository ./tests/integration/database/...
go test -bench=BenchmarkRabbitMQ ./tests/integration/queue/...
go test -bench=BenchmarkPerformance ./tests/performance/...
```

## 📊 Test Coverage Goals

### Coverage Targets
- **Domain Layer**: 95%+ coverage (business logic is critical)
- **Application Layer**: 90%+ coverage (use cases and services)
- **Infrastructure Layer**: 80%+ coverage (integration points)
- **API Layer**: 90%+ coverage (all endpoints tested)
- **Overall Target**: 90%+ coverage

### Quality Metrics
- **API Response Time**: <200ms for 95th percentile
- **Concurrent Operations**: Support 100+ simultaneous requests
- **Queue Throughput**: 100+ messages/second enqueue/dequeue
- **Database Performance**: <100ms for 95% of queries
- **Zero Data Loss**: Queued tasks must be persistent
- **Error Recovery**: Graceful handling of external service failures

## 🔧 Test Configuration

### Environment Variables
```bash
GO_ENV=test                    # Set test environment
DOCKER_API_VERSION=1.41       # Docker API compatibility
TEST_DATABASE_URL=...         # Test database connection
TEST_RABBITMQ_URL=...         # Test RabbitMQ connection
```

### Docker Requirements
- Docker Engine 20.10+
- 4GB+ available memory for test containers
- Network access for pulling test images

### Database Setup
Tests automatically create temporary MySQL containers with:
- Database: `workflow_test`
- User: `root` / Password: `testpass`
- Automatic schema migration
- Cleanup after test completion

### Queue Setup
Tests automatically create temporary RabbitMQ containers with:
- User: `admin` / Password: `password`
- Exchange: `tasks_test`
- Queue: `task_queue_test`
- Automatic cleanup after test completion

## 📈 Performance Benchmarks

### Baseline Performance Targets

**Database Operations:**
- Task Creation: >1000 ops/sec
- Task Queries: >5000 ops/sec
- Concurrent Users: 100+ simultaneous
- Transaction Time: <50ms average

**Queue Operations:**
- Message Enqueue: >1000 msgs/sec
- Message Dequeue: >500 msgs/sec
- Queue Latency: <10ms average
- Message Persistence: 100% reliable

**API Endpoints:**
- Task CRUD: <200ms 95th percentile
- Task Listing: <100ms for 100 items
- Concurrent Requests: 100+ simultaneous
- Error Rate: <1% under normal load

## 🐛 Test-Driven Development

### Writing New Tests

1. **Domain Tests**: Test business logic in isolation
```go
func TestTask_BusinessRule(t *testing.T) {
    // Arrange: Create test entities
    // Act: Execute business operation
    // Assert: Verify business rules enforced
}
```

2. **Integration Tests**: Test with real dependencies
```go
func TestRepository_Integration(t *testing.T) {
    env := testutils.SetupTestEnvironment(t)
    defer env.TearDown(t)
    // Test with real database
}
```

3. **API Tests**: Test HTTP endpoints
```go
func TestAPI_Endpoint(t *testing.T) {
    router, env := setupAPITestEnvironment(t)
    defer env.TearDown(t)
    // Test HTTP requests/responses
}
```

4. **Performance Tests**: Test under load
```go
func TestPerformance_Load(t *testing.T) {
    config := testutils.DefaultBenchmarkConfig()
    // Test concurrent operations
}
```

### Test Data Management

- Use factories for consistent test data creation
- Clean database state between tests
- Use realistic but minimal test datasets
- Avoid test interdependencies

### Mocking Strategy

- Mock external dependencies (databases, queues, APIs)
- Use interfaces for dependency injection
- Prefer real implementations for integration tests
- Mock time sources for deterministic tests

## 📝 Test Reports

### Generated Artifacts

- `coverage.html`: Interactive coverage report
- `*-test-results.log`: Detailed test execution logs
- `benchmark-results.log`: Performance benchmark results
- Test timing and resource usage metrics

### Continuous Integration

Tests are designed to run in CI/CD pipelines with:
- Parallel test execution
- Container orchestration
- Coverage reporting
- Performance regression detection
- Quality gate enforcement

## 🎯 Success Criteria

✅ **90%+ test coverage** for business logic components  
✅ **All API endpoints tested** with various scenarios  
✅ **Database operations validated** with real MySQL  
✅ **Queue integration tested** with RabbitMQ  
✅ **Performance benchmarks** meet requirements  
✅ **Concurrent operation handling** (100+ simultaneous requests)  
✅ **Error scenarios covered** for all failure modes  
✅ **CI/CD ready** test suite with automation  

This comprehensive testing suite ensures the Go backend task queue system is production-ready, reliable, and performant. The testing strategy covers all layers of the application with appropriate testing techniques for each component type.