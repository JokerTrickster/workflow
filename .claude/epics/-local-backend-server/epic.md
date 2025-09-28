---
github: "https://github.com/JokerTrickster/workflow/issues/57"
last_sync: "2025-09-26T17:16:05.524426Z"
status: completed

---


# Epic: local-backend-server

## Overview

Implementation of an always-on Go service using clean architecture that continuously processes RabbitMQ messages for Claude API-based code analysis. The system focuses on simplicity and modularity with sequential message processing, SQLite persistence, and comprehensive testing coverage.

## Architecture Decisions

- **Language**: Go for performance, concurrency support, and simple deployment
- **Architecture Pattern**: Clean Architecture with 4-layer separation (Domain, Application, Infrastructure, Presentation)
- **Message Queue**: RabbitMQ client with automatic reconnection handling
- **Database**: SQLite for lightweight local persistence with GORM for ORM
- **API Client**: Official Anthropic Go SDK for Claude API integration
- **Configuration**: Environment variables with viper for settings management
- **Testing**: Standard Go testing with testify for assertions and table-driven tests

## Technical Approach

### Backend Services

**Core Domain Models:**
- `Message` entity with type, payload, metadata
- `Request` entity with status tracking (pending, processing, completed, failed, cancelled)
- `ProcessingContext` for maintaining Claude API conversation state

**Application Services:**
- `MessageProcessor` - orchestrates message handling workflow
- `ClaudeService` - manages API calls with context preservation
- `RequestService` - handles request lifecycle and status updates

**Infrastructure Components:**
- `RabbitMQConsumer` - message queue integration with reconnection logic
- `SQLiteRepository` - database operations with transaction support
- `ConfigManager` - environment variable and settings management

### Infrastructure

**Deployment Considerations:**
- Single binary deployment for laptop environment
- Configuration through environment variables
- SQLite database file for local persistence
- Log output to stdout (no complex logging framework needed)

**Scaling Requirements:**
- Sequential processing initially (single goroutine)
- Designed for future concurrent processing enhancement
- Minimal memory footprint for continuous operation

## Implementation Strategy

**Development Phases:**
1. Core domain models and business logic
2. Infrastructure layer (database, queue, API clients)
3. Application services and message processing workflow
4. Configuration management and error handling
5. Comprehensive testing suite

**Risk Mitigation:**
- Mock external dependencies for reliable testing
- Graceful error handling to prevent service crashes
- Database transactions to ensure data consistency

**Testing Approach:**
- Unit tests for all business logic (domain/application layers)
- Integration tests for infrastructure components
- End-to-end tests with test RabbitMQ and mock Claude API
- Table-driven tests for message type routing

## Task Breakdown Preview

High-level task categories that will be created:
- [ ] **Project Setup**: Go module initialization, dependency management, project structure
- [ ] **Domain Layer**: Core entities, business rules, and interfaces
- [ ] **Database Infrastructure**: SQLite setup, GORM models, repository pattern
- [ ] **RabbitMQ Integration**: Consumer implementation, connection management, message parsing
- [ ] **Claude API Integration**: Service implementation, context management, response handling
- [ ] **Application Services**: Message processing workflow, request lifecycle management
- [ ] **Configuration Management**: Environment variables, settings validation
- [ ] **Testing Suite**: Unit, integration, and end-to-end tests
- [ ] **Error Handling**: Graceful failure handling, logging, recovery mechanisms

## Dependencies

**External Service Dependencies:**
- Running RabbitMQ instance (local configuration)
- Claude API access with valid API key
- Go runtime environment (1.21+)

**Go Package Dependencies:**
- `github.com/streadway/amqp` - RabbitMQ client
- `gorm.io/gorm` and `gorm.io/driver/sqlite` - Database ORM
- `github.com/anthropics/anthropic-sdk-go` - Claude API client
- `github.com/spf13/viper` - Configuration management
- `github.com/stretchr/testify` - Testing framework

**Internal Dependencies:**
- Message schema definitions (JSON structures)
- Database migration scripts
- Test data fixtures and mock responses

## Success Criteria (Technical)

**Performance Benchmarks:**
- Message processing latency < 100ms (excluding Claude API calls)
- Database operations < 10ms for status updates
- Memory usage < 50MB during continuous operation
- 100% message processing success rate for valid JSON

**Quality Gates:**
- 90%+ code coverage across all layers
- Zero critical security vulnerabilities
- All tests passing in CI environment
- Clean architecture boundaries maintained (no circular dependencies)

**Acceptance Criteria:**
- Successfully processes work request and cancellation messages
- Maintains accurate request status in database
- Preserves Claude API context across related requests
- Handles RabbitMQ connection failures gracefully
- Comprehensive test coverage for all message types

## Estimated Effort

**Overall Timeline:** 5-7 development days
- Project setup and domain modeling: 1 day
- Infrastructure layer implementation: 2 days  
- Application services and workflow: 1-2 days
- Testing suite development: 1-2 days

**Resource Requirements:**
- Single developer with Go experience
- Access to RabbitMQ and Claude API for testing
- Local development environment setup

**Critical Path Items:**
1. Domain model design (affects all other layers)
2. RabbitMQ integration (core functionality dependency)
3. Claude API service (primary business value)
4. Request status management (user-facing feature)

## Stats

Total tasks: 8
Parallel tasks: 4 (can be worked on simultaneously)
Sequential tasks: 4 (have dependencies)
Estimated total effort: 116 hours (approximately 15 development days)
