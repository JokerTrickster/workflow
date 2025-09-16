---
name: local-backend-server
description: Always-on local Go server that processes RabbitMQ messages and sends code analysis requests to Claude API
status: backlog
created: 2025-09-16T13:19:31Z
---

# PRD: Local Backend Server

## Executive Summary

The Local Backend Server is a continuously running Go service designed to process messages from RabbitMQ queues and execute code analysis tasks through the Claude API. Built with clean architecture principles, this server will handle multiple message types (work requests, cancellations) while maintaining simplicity and modularity. The system focuses on core functionality without complex monitoring or authentication layers.

## Problem Statement

Currently, there's no automated way to process code analysis requests in a queue-based system. Manual processing is inefficient and doesn't scale well for batch operations. We need a reliable, always-on service that can:

- Continuously monitor message queues for new tasks
- Process different types of work requests systematically  
- Integrate seamlessly with Claude API for code analysis
- Maintain context across related requests
- Provide visibility into processing status through database tracking

**Why now?** The increasing volume of code analysis tasks requires automation, and having a dedicated local server ensures consistent processing without external dependencies.

## User Stories

### Primary User Persona: Developer/System Administrator
A developer who needs automated code analysis and wants to queue multiple requests for processing.

**User Journey:**
1. Developer submits code analysis request to RabbitMQ queue
2. Local backend server picks up the message automatically
3. Server processes request through Claude API with proper context management
4. Results are logged and status updated in database
5. Developer can check processing status and results

### Key User Stories:

**Story 1: Queue Work Requests**
- As a developer, I want to submit code analysis requests to a queue
- So that I can batch multiple requests without manual intervention
- Acceptance Criteria: JSON messages are successfully queued and processed sequentially

**Story 2: Cancel Pending Work**  
- As a developer, I want to cancel pending work requests
- So that I can stop unnecessary processing when requirements change
- Acceptance Criteria: Cancellation messages properly remove pending tasks from processing

**Story 3: Track Processing Status**
- As a developer, I want to see the status of my requests
- So that I know when analysis is complete and can retrieve results
- Acceptance Criteria: Database maintains accurate status for all requests with timestamps

## Requirements

### Functional Requirements

**Core Processing Engine**
- Process JSON messages from RabbitMQ queues
- Support multiple message types: work requests, work cancellations
- Sequential processing (one message at a time initially)
- Context management for related Claude API calls
- Database persistence for request tracking and status updates

**Message Handling**  
- JSON message parsing and validation
- Message type routing (work request vs. cancellation)
- Error handling for malformed messages
- Graceful handling of unsupported message types

**Claude API Integration**
- Direct Claude API calls using default model
- Code analysis request processing
- Context preservation across related requests
- Response logging for debugging purposes

**Data Persistence**
- Request status tracking (pending, processing, completed, failed, cancelled)
- Message metadata storage
- Processing timestamps and duration tracking
- Basic query capabilities for status checks

### Non-Functional Requirements

**Performance**
- Single-threaded processing initially (can be enhanced later)
- Minimal latency for message pickup from queue
- Efficient memory usage for context management

**Reliability**
- Automatic reconnection to RabbitMQ on connection loss
- Proper error handling without server crashes
- Data consistency in status updates

**Maintainability**
- Clean architecture with clear separation of concerns
- Modular design for easy feature additions
- Comprehensive unit and integration tests
- Configuration management for queue settings and API keys

## Success Criteria

**Measurable Outcomes:**
- Successfully processes 100% of valid JSON messages from RabbitMQ
- Maintains 99%+ uptime when running continuously
- Context is properly managed across related Claude API calls
- Database accurately reflects processing status for all requests
- Zero data loss during normal operation

**Key Metrics:**
- Message processing success rate
- Average processing time per message
- Claude API response time and success rate
- Database query response time
- Memory usage during continuous operation

## Constraints & Assumptions

**Technical Constraints:**
- Go language implementation required
- RabbitMQ integration (local instance initially, configurable later)
- SQLite or similar lightweight database for local storage
- Claude API rate limits and token constraints

**Resource Constraints:**
- Single developer implementation
- Local laptop deployment initially
- Minimal external dependencies

**Assumptions:**
- RabbitMQ instance is already running and accessible
- Claude API keys are available and valid
- Local storage is sufficient for request tracking
- Network connectivity is stable for API calls

## Out of Scope

**Explicitly NOT building:**
- Advanced monitoring dashboards or metrics collection
- Authentication/authorization systems
- Rate limiting or throttling mechanisms
- Health check endpoints or status APIs
- Complex retry logic or dead letter queue handling
- Multi-node deployment or clustering
- Real-time notifications or webhooks
- Advanced logging frameworks or log aggregation
- Performance optimization beyond basic efficiency
- UI or web interface for status checking

## Dependencies

**External Dependencies:**
- RabbitMQ server (assumed running)
- Claude API availability and access
- Go runtime environment
- Database driver (SQLite or similar)

**Internal Dependencies:**
- Configuration system for queue connection details
- Message schema definitions for JSON parsing
- Database schema for request tracking
- Test data sets for validation

## Technical Architecture Overview

**Clean Architecture Layers:**
1. **Presentation Layer**: Message consumers and API clients
2. **Application Layer**: Business logic for message processing and workflow
3. **Domain Layer**: Core entities (Message, Request, Status) and business rules
4. **Infrastructure Layer**: RabbitMQ client, Claude API client, database operations

**Key Components:**
- Message Queue Consumer (RabbitMQ integration)
- Message Router (handles different message types)
- Claude Service (API integration with context management)
- Request Repository (database operations)
- Configuration Manager (settings and environment variables)

**Testing Strategy:**
- Unit tests for all business logic components
- Integration tests for queue and API interactions
- Mock implementations for external dependencies
- End-to-end tests with test messages and responses