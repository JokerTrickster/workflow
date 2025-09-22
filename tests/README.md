# Task History Integration Test Suite

This comprehensive test suite validates the complete task history workflow from frontend submission to backend processing and API retrieval.

## Test Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │   Backend API   │    │   Database      │
│                 │    │                 │    │                 │
│ TaskHistory     │───▶│ /api/v1/tasks/  │───▶│ WorkflowHistory │
│ Components      │    │ history/{repo}  │    │ Table           │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         │              │   Queue System  │              │
         │              │                 │              │
         └──────────────▶│ RabbitMQ +      │◀─────────────┘
                        │ AtomicService   │
                        └─────────────────┘
```

## Test Coverage

### 1. End-to-End Integration Tests (`/tests/integration/`)

**File: `task_history_integration_test.go`**

- **E2E Happy Path**: Submit task → Queue → Database → API retrieval → Frontend display
- **Atomic Transaction Testing**: Success and failure scenarios with rollback validation
- **API Pagination Testing**: Multiple pages with various limits and edge cases
- **Error Scenario Testing**: Database failures, invalid inputs, network issues
- **Performance Testing**: API response times under load (<200ms requirement)
- **Concurrent Operation Testing**: Multiple simultaneous queue operations

### 2. Complete Workflow Tests (`/tests/e2e/`)

**File: `complete_workflow_test.go`**

- **Complete Workflow**: Frontend submission through to API retrieval
- **Multiple Tasks Workflow**: Concurrent task handling and processing
- **Error Scenarios**: Queue failures, database errors, rollback validation
- **Performance Requirements**: Response time validation under realistic load

### 3. Performance Tests (`/tests/performance/`)

**File: `api_performance_test.go`**

- **Response Time Requirements**: <200ms API response validation
- **Concurrent Request Performance**: Load testing with multiple concurrent users
- **Memory Usage Performance**: Bounded memory usage verification
- **Database Index Performance**: Index effectiveness validation
- **Stress Testing**: Sustained load testing with high throughput

### 4. Frontend Integration Tests (`/frontend/src/__tests__/integration/`)

**File: `TaskHistoryIntegration.test.tsx`**

- **Component Integration**: TaskHistory, Pagination, StatusBadge working together
- **Real-time Polling**: Status updates and polling behavior
- **Error Handling**: API failures, timeout scenarios, retry mechanisms
- **Performance**: Large dataset handling, debounced interactions
- **Concurrent Operations**: Multiple component instances, repository switching

## Test Categories

### Critical Path Tests ✅
- Task submission through API
- Database atomic operations
- Queue message publishing
- API data retrieval
- Frontend component rendering

### Edge Case Tests ⚠️
- Empty repositories
- Invalid pagination parameters
- Database connection failures
- Queue publishing failures
- Network timeout scenarios

### Performance Tests ⚡
- <200ms API response requirement
- Concurrent request handling
- Large dataset pagination
- Memory usage optimization
- Database query performance

### Error Handling Tests ❌
- Database rollback scenarios
- API error responses
- Frontend error states
- Network failure recovery
- Invalid input validation

## Running Tests

### Quick Start
```bash
# Run all integration tests
./scripts/run_integration_tests.sh

# Verbose output
./scripts/run_integration_tests.sh --verbose

# Backend tests only
./scripts/run_integration_tests.sh --backend-only

# Frontend tests only
./scripts/run_integration_tests.sh --frontend-only

# Include load tests
./scripts/run_integration_tests.sh --load-tests
```

### Individual Test Suites

#### Backend Tests
```bash
# Integration tests
go test -v ./tests/integration

# E2E tests
go test -v ./tests/e2e

# Performance tests
go test -v ./tests/performance

# Specific test pattern
go test -v -run TestE2E_HappyPath ./tests/integration
```

#### Frontend Tests
```bash
cd frontend

# Integration tests
npm test -- --testPathPattern="integration"

# Specific test file
npm test -- TaskHistoryIntegration.test.tsx
```

## Test Data & Fixtures

### Mock Components

1. **MockQueuePublisher**: Simulates RabbitMQ message publishing
2. **MockQueueConsumer**: Simulates local-backend queue processing
3. **In-Memory Database**: SQLite for isolated test environments

### Test Data Generation

- **Small Dataset**: 50 records for basic functionality
- **Medium Dataset**: 200 records for pagination testing
- **Large Dataset**: 1000+ records for performance testing
- **Realistic Data**: Varied statuses, processing times, and metadata

## Performance Requirements

### API Response Times
- **Target**: <200ms for typical queries
- **Pagination**: <200ms regardless of page size
- **Large Datasets**: Linear scaling with page size
- **Concurrent Load**: Maintained performance under 20 concurrent requests

### Database Performance
- **Index Usage**: Compound indexes on (repository_name, created_at)
- **Query Optimization**: EXPLAIN plan validation
- **Memory Usage**: Bounded by page size, not total dataset size

### Frontend Performance
- **Rendering**: <1 second for large datasets
- **Debouncing**: Rapid interactions properly handled
- **Memory Leaks**: No retention of old repository data

## Test Environment Setup

### Prerequisites
- Go 1.19+
- Node.js 18+
- SQLite3
- npm dependencies installed

### Environment Variables
```bash
NODE_ENV=test
GO_ENV=test
```

### Test Database
- In-memory SQLite for isolation
- Auto-migration of schemas
- Clean state for each test

## Validation Checklist

### End-to-End Workflow ✅
- [ ] Task submission creates database record
- [ ] Queue message published atomically
- [ ] Processing updates database status
- [ ] API retrieval returns correct data
- [ ] Frontend displays all information

### Error Handling ✅
- [ ] Queue failures rollback database
- [ ] API errors return structured responses
- [ ] Frontend gracefully handles errors
- [ ] Retry mechanisms work correctly

### Performance ✅
- [ ] API responds within 200ms
- [ ] Concurrent requests handled efficiently
- [ ] Large datasets paginated properly
- [ ] Memory usage bounded

### Data Integrity ✅
- [ ] Atomic operations maintain consistency
- [ ] Pagination calculations correct
- [ ] Timestamps accurate and formatted
- [ ] Status transitions valid

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Ensure SQLite permissions
   - Check auto-migration status

2. **API Timeout Issues**
   - Verify database indexes
   - Check query optimization

3. **Frontend Test Failures**
   - Ensure mock services configured
   - Check async operation handling

4. **Performance Test Failures**
   - System resource constraints
   - Database optimization needed

### Debug Mode
```bash
# Enable verbose logging
VERBOSE=true ./scripts/run_integration_tests.sh

# Go test debugging
go test -v -race ./tests/integration

# Frontend debugging
npm test -- --verbose TaskHistoryIntegration.test.tsx
```

## Continuous Integration

### GitHub Actions Integration
```yaml
- name: Run Integration Tests
  run: |
    chmod +x ./scripts/run_integration_tests.sh
    ./scripts/run_integration_tests.sh --backend-only

- name: Run Frontend Tests
  run: |
    cd frontend
    npm test -- --testPathPattern="integration" --coverage
```

### Test Reports
- Go test output with timing
- Jest coverage reports
- Performance metrics logging
- Error rate monitoring

## Maintenance

### Adding New Tests
1. Follow existing test patterns
2. Include both positive and negative cases
3. Add performance validation
4. Update documentation

### Test Data Updates
- Keep datasets realistic
- Maintain test isolation
- Update fixtures for schema changes
- Preserve test consistency

### Performance Monitoring
- Track response time trends
- Monitor memory usage patterns
- Validate index effectiveness
- Update requirements as needed