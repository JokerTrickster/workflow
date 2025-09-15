---
name: task-queue-backend
description: Backend API system for task queue management with RabbitMQ integration and clean architecture
status: backlog
created: 2025-09-15T02:21:23Z
---

# PRD: task-queue-backend

## Executive Summary

Develop a backend API system that enables web applications to create, manage, and process tasks through a RabbitMQ queue system. The system will follow clean architecture principles using Go with Echo framework, providing RESTful APIs for task creation, deletion, and resumption while maintaining task state in MySQL database.

## Problem Statement

The current system lacks a robust task queue management solution that can:
- Handle task creation requests from web interfaces
- Reliably queue tasks for asynchronous processing
- Provide task lifecycle management (create, delete, resume)
- Maintain task state persistence and tracking
- Scale to handle multiple concurrent task operations

This backend system will serve as the critical middleware between web applications and task processing systems, ensuring reliable task delivery and state management.

## User Stories

### Primary Personas
- **Frontend Developers**: Need reliable APIs to submit and manage tasks from web applications
- **Task Processors**: Need guaranteed task delivery from the queue system
- **System Administrators**: Need visibility into task states and system health

### User Journeys

**Task Creation Flow**
1. Frontend sends task creation request with branch name, title, content, and repository
2. Backend validates and processes the request
3. Task is stored in MySQL database with pending status
4. Task is published to RabbitMQ claude-tasks queue
5. Response sent back to frontend with task ID and status

**Task Management Flow**
1. User requests task deletion through frontend
2. Backend marks task as cancelled in database
3. Task removal attempt from queue (if still pending)
4. Confirmation response sent to frontend

**Task Resume Flow**
1. User requests task resumption for failed/paused task
2. Backend validates task is eligible for resume
3. Task status updated to pending in database
4. Task re-queued to RabbitMQ
5. Confirmation response sent to frontend

## Requirements

### Functional Requirements

**Core API Endpoints**
- `POST /api/tasks` - Create new task
- `DELETE /api/tasks/{id}` - Delete/cancel task
- `PUT /api/tasks/{id}/resume` - Resume failed/paused task
- `GET /api/tasks/{id}` - Get task status
- `GET /api/tasks` - List tasks with filtering

**Task Data Structure**
```json
{
  "id": "uuid",
  "branch_name": "string",
  "title": "string", 
  "content": "string",
  "repository": "string",
  "user_id": "string",
  "status": "pending|processing|completed|failed|cancelled",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**RabbitMQ Integration**
- Connect to RabbitMQ at http://13.203.37.93:15672/
- Publish tasks to `claude-tasks` topic/queue
- Message format: JSON with task data
- Handle connection failures gracefully

**Database Operations**
- MySQL database for task persistence
- Task state tracking across lifecycle
- User session association
- Created/updated timestamp management

### Non-Functional Requirements

**Performance**
- Handle concurrent task creation requests
- Minimal latency for API responses (<200ms)
- Efficient database queries with proper indexing

**Reliability**
- Graceful handling of RabbitMQ connection failures
- Database transaction consistency
- Proper error handling and logging

**Scalability**
- Stateless API design for horizontal scaling
- Connection pooling for database and RabbitMQ
- Efficient queue message handling

**Security**
- Input validation for all API endpoints
- SQL injection prevention
- Proper error messages without sensitive data exposure

## Success Criteria

**Functional Success**
- Successfully create tasks via API and store in both DB and queue
- Task deletion removes from queue and updates DB status
- Task resume functionality re-queues eligible tasks
- All APIs return proper HTTP status codes and JSON responses

**Performance Metrics**
- API response time < 200ms for 95th percentile
- Support 100+ concurrent task operations
- Zero data loss for queued tasks

**Quality Metrics**
- Clean architecture implementation with proper layer separation
- Comprehensive error handling coverage
- Database transaction consistency maintained

## Constraints & Assumptions

**Technical Constraints**
- Must use Go with Echo framework
- MySQL database already running on target server
- RabbitMQ instance already configured and running
- No authentication/authorization required initially

**Assumptions**
- RabbitMQ server remains available and accessible
- MySQL database has sufficient storage capacity
- Network connectivity between services is reliable
- Task processing consumers will be implemented separately

## Out of Scope

**Explicitly NOT Building**
- Task processing/worker implementation
- User authentication and authorization system
- Web frontend interface
- Task scheduling or cron functionality
- Task result storage and retrieval
- Monitoring and alerting dashboards
- Load balancing or service discovery
- Database migration scripts

## Dependencies

**External Dependencies**
- RabbitMQ server at http://13.203.37.93:15672/
- MySQL database server on target environment
- Go Echo framework and related packages
- AMQP Go client library for RabbitMQ integration

**Internal Dependencies**
- Database schema creation for tasks table
- Environment configuration for database and RabbitMQ connections
- Error handling and logging framework setup

**Infrastructure Dependencies**
- Server environment with Go runtime
- Network connectivity to RabbitMQ and MySQL
- Proper firewall rules for service communication

## Technical Architecture

**Clean Architecture Layers**
1. **Presentation Layer**: Echo HTTP handlers and middleware
2. **Use Cases Layer**: Business logic for task operations
3. **Domain Layer**: Task entity and repository interfaces
4. **Infrastructure Layer**: MySQL and RabbitMQ implementations

**Dependency Injection**
- Constructor-based dependency injection
- Interface-based abstractions for testability
- Configuration injection for environment-specific settings

**Error Handling Strategy**
- Structured error responses with proper HTTP codes
- Comprehensive logging for debugging and monitoring
- Graceful degradation for external service failures