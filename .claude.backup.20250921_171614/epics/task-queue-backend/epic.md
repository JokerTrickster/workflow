---
name: task-queue-backend
status: backlog
created: 2025-09-15T02:23:51Z
progress: 0%
prd: .claude/prds/task-queue-backend.md
github: https://github.com/JokerTrickster/workflow/issues/47
---

# Epic: task-queue-backend

## Overview
Implement a Go-based backend API using Echo framework with clean architecture for task queue management. The system provides RESTful endpoints for task lifecycle operations (create, delete, resume) while maintaining dual persistence in MySQL database and RabbitMQ message queue for reliable task processing.

## Architecture Decisions
- **Framework**: Go with Echo for lightweight, high-performance HTTP server
- **Architecture Pattern**: Clean Architecture with dependency injection for testability and maintainability
- **Database**: MySQL for task persistence with proper indexing for performance
- **Message Queue**: RabbitMQ integration using AMQP protocol for reliable task delivery
- **JSON Communication**: Standard REST API with JSON request/response payloads
- **Error Handling**: Structured error responses with proper HTTP status codes
- **No Authentication**: Simplified initial implementation without auth layer

## Technical Approach

### Backend Services
**API Layer (Presentation)**
- Echo HTTP server with middleware for logging and error handling
- REST endpoints: POST /api/tasks, DELETE /api/tasks/{id}, PUT /api/tasks/{id}/resume, GET /api/tasks/{id}, GET /api/tasks
- Input validation and sanitization for all endpoints
- Structured JSON responses with consistent error format

**Business Logic (Use Cases)**
- Task creation service: validate input → save to DB → publish to queue
- Task deletion service: update DB status → attempt queue removal
- Task resume service: validate eligibility → update status → re-queue
- Task query service: fetch from DB with filtering support

**Domain Layer**
- Task entity with complete lifecycle states (pending, processing, completed, failed, cancelled)
- Repository interfaces for database and queue operations
- Domain-specific validation rules and business logic

**Infrastructure Layer**
- MySQL repository implementation with connection pooling
- RabbitMQ publisher implementation with connection recovery
- Configuration management for environment-specific settings

### Database Schema
```sql
CREATE TABLE tasks (
  id VARCHAR(36) PRIMARY KEY,
  branch_name VARCHAR(255) NOT NULL,
  title VARCHAR(500) NOT NULL,
  content TEXT,
  repository VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user_status (user_id, status),
  INDEX idx_created_at (created_at)
);
```

### Message Queue Integration
- Connect to RabbitMQ at http://13.203.37.93:15672/
- Publish to `claude-tasks` queue with JSON message format
- Implement connection recovery and graceful failure handling
- Message structure matches task entity for consistency

## Implementation Strategy
**Phase 1: Core Infrastructure** - Set up project structure, database connection, basic Echo server
**Phase 2: Domain & Use Cases** - Implement business logic and repository interfaces  
**Phase 3: API Implementation** - Build REST endpoints with proper error handling
**Phase 4: Queue Integration** - Add RabbitMQ publishing with reliability features
**Phase 5: Testing & Validation** - Comprehensive testing and performance validation

**Risk Mitigation**
- Use database transactions to ensure data consistency
- Implement circuit breaker pattern for RabbitMQ connectivity
- Add comprehensive logging for debugging and monitoring
- Design for stateless operation to enable horizontal scaling

**Testing Approach**
- Unit tests for business logic with dependency injection
- Integration tests for database and queue operations
- API tests for endpoint validation and error scenarios
- Performance tests for concurrent operation handling

## Task Breakdown Preview
High-level task categories that will be created:
- [ ] Project Setup: Go module initialization, directory structure, dependencies
- [ ] Database Layer: MySQL connection, task repository, schema creation
- [ ] Domain Layer: Task entity, business logic, repository interfaces
- [ ] Use Cases: Task operations (create, delete, resume, query) with validation
- [ ] API Layer: Echo server setup, REST endpoints, middleware, error handling
- [ ] RabbitMQ Integration: Connection management, message publishing, error recovery
- [ ] Configuration: Environment settings, dependency injection container
- [ ] Testing: Unit tests, integration tests, API tests
- [ ] Documentation: API documentation, deployment guide

## Dependencies
**External Dependencies**
- RabbitMQ server (http://13.203.37.93:15672/) - must be accessible and claude-tasks queue created
- MySQL database server - must be running with appropriate user permissions
- Go Echo framework and AMQP library packages

**Internal Dependencies**
- Database schema creation and initial setup
- Environment configuration file for connection parameters
- Basic logging framework for error tracking and debugging

**Infrastructure Dependencies**
- Go runtime environment (1.19+)
- Network connectivity to both MySQL and RabbitMQ services
- Server environment with appropriate resource allocation

## Success Criteria (Technical)
**Performance Benchmarks**
- API response time <200ms for 95th percentile under normal load
- Support 100+ concurrent requests without degradation
- Database queries optimized with proper indexing

**Quality Gates**
- Clean architecture compliance with proper layer separation
- 90%+ test coverage for business logic components
- Zero memory leaks and proper resource cleanup
- Graceful error handling for all external dependencies

**Acceptance Criteria**
- All REST endpoints functional and properly documented
- Tasks successfully persisted in both MySQL and RabbitMQ
- Task lifecycle operations work correctly (create→delete→resume)
- System handles RabbitMQ and MySQL connection failures gracefully

## Tasks Created
- [ ] #48 - RabbitMQ Integration (parallel: false)
- [ ] #49 - Project Setup and Go Module Initialization (parallel: true)
- [ ] #50 - Database Layer and MySQL Integration (parallel: false)
- [ ] #51 - Configuration Management and Environment Setup (parallel: true)
- [ ] #52 - Testing Implementation (parallel: false)
- [ ] #53 - Domain Layer Implementation (parallel: false)
- [ ] #54 - Use Cases and Business Logic (parallel: false)
- [ ] #55 - Documentation and Deployment (parallel: true)
- [ ] #56 - API Layer and REST Endpoints (parallel: false)

Total tasks: 9
Parallel tasks: 3 (49, 51, 55)
Sequential tasks: 6 (48, 50, 52, 53, 54, 56)
Estimated total effort: 72-96 hours (2-3 weeks for 1 developer)
## Estimated Effort
**Overall Timeline**: 2-3 weeks for full implementation
**Resource Requirements**: 1 backend developer with Go experience
**Critical Path Items**: 
1. Database and RabbitMQ connectivity (foundation for all features)
2. Clean architecture setup (affects all subsequent development)
3. API endpoint implementation (user-facing functionality)
4. Integration testing (validation of complete system)
