---
name: task-history-api
description: Database integration for task queuing and REST API for repository-specific task history management
status: backlog
created: 2025-09-22T03:21:57Z
---

# PRD: Task History API

## Executive Summary

The Task History API system integrates database persistence with the existing workflow queue system and provides a REST API for frontend consumption. This feature enables task tracking from queue insertion through completion, with repository-specific filtering for frontend display.

## Problem Statement

Currently, the workflow system lacks persistent task tracking and historical data access:
- Tasks are queued via RabbitMQ but not stored in database for tracking
- Frontend has no way to display task history for specific repositories
- No visibility into task status progression (pending → processing → completed/failed)
- Users cannot track workflow execution history or debug failed tasks

## User Stories

### Primary User: Frontend Application
**As a frontend application**
- I want to display task history for a specific repository
- I want to show current task status and progression
- I want to paginate through task history efficiently
- So that users can track their workflow executions

### Primary User: System Administrator
**As a system administrator**
- I want all queued tasks to be persisted in database immediately
- I want atomic operations (queue + database or complete failure)
- I want to query task status and history for troubleshooting
- So that I can maintain system reliability and provide user support

## Requirements

### Functional Requirements

#### FR1: Database Integration During Queue Operations
- **FR1.1**: Insert task record to `workflow_histories` table immediately upon successful queue insertion
- **FR1.2**: Use transaction pattern: if database insertion fails, queue operation must fail completely
- **FR1.3**: Generate unique `request_id` for each task
- **FR1.4**: Store all task metadata matching existing database schema

#### FR2: Task Status Management
- **FR2.1**: Initialize all tasks with `pending` status
- **FR2.2**: Support status transitions: `pending` → `processing` → `completed`/`failed`/`cancelled`
- **FR2.3**: Update `completed_at` and `processing_time_ms` fields upon completion
- **FR2.4**: Store error messages in `error` field for failed tasks

#### FR3: Task History REST API
- **FR3.1**: Provide GET endpoint for repository-specific task listing
- **FR3.2**: Support pagination with page/limit parameters
- **FR3.3**: Return tasks sorted by creation date (newest first)
- **FR3.4**: Include all relevant task fields (id, request_id, status, tasks, repository_name, created_at, completed_at, processing_time_ms, error)

### Non-Functional Requirements

#### NFR1: Performance
- **NFR1.1**: API response time < 200ms for typical queries (20-50 tasks)
- **NFR1.2**: Database insertion must not significantly slow queue operations
- **NFR1.3**: Support concurrent queue operations without blocking

#### NFR2: Reliability
- **NFR2.1**: Atomic transactions for queue + database operations
- **NFR2.2**: Graceful error handling with appropriate HTTP status codes
- **NFR2.3**: Database connection pooling for efficient resource usage

#### NFR3: Security
- **NFR3.1**: Input validation for all API parameters
- **NFR3.2**: SQL injection prevention through parameterized queries
- **NFR3.3**: Rate limiting considerations for API endpoints

## API Specification

### GET /api/tasks/history/{repository_name}

**Parameters:**
- `repository_name` (path): Exact repository name for filtering
- `page` (query, optional): Page number (default: 1)
- `limit` (query, optional): Items per page (default: 20, max: 100)

**Response Format:**
```json
{
  "data": [
    {
      "id": 123,
      "request_id": "uuid-v4-string",
      "status": "completed",
      "tasks": "Task description",
      "repository_name": "my-repo",
      "working_dir": "/path/to/repo",
      "claude_cmd": "command string",
      "interactive": false,
      "continue_task": false,
      "created_at": "2025-09-22T03:21:57Z",
      "completed_at": "2025-09-22T03:25:30Z",
      "processing_time_ms": 213000,
      "result": "Success details",
      "error": null
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

**HTTP Status Codes:**
- `200`: Success
- `400`: Invalid parameters
- `404`: Repository not found
- `500`: Server error

## Success Criteria

### Measurable Outcomes
1. **Database Integration**: 100% of queued tasks are persisted in database
2. **API Performance**: Average response time < 200ms for 95% of requests
3. **Data Consistency**: Zero data loss between queue and database operations
4. **Frontend Integration**: Frontend can successfully display task history for any repository

### Key Metrics
- Database insertion success rate: 100%
- API error rate: < 1%
- Task status accuracy: 100% match between actual and stored status
- Frontend polling efficiency: Successful data retrieval in < 3 polling attempts

## Technical Implementation Details

### Database Schema (Existing)
Uses existing `workflow_histories` table with all required fields already defined.

### Backend Integration Points
1. **Queue Handler**: Modify existing queue insertion logic in `backend/` folder
2. **Database Layer**: Utilize existing GORM models and MySQL connection
3. **API Router**: Add new REST endpoint to existing Echo/Gin router

### Error Handling Strategy
- **Queue + DB Transaction**: Use database transactions to ensure atomicity
- **API Errors**: Return appropriate HTTP status codes with error details
- **Logging**: Comprehensive logging for debugging and monitoring

## Constraints & Assumptions

### Technical Constraints
- Must use existing MySQL database and GORM models
- Must integrate with current RabbitMQ queue system
- Must not break existing queue processing workflow

### Timeline Constraints
- Backend integration: Priority task for immediate implementation
- API development: Can be done in parallel with backend work

### Resource Constraints
- Use existing database connection pool
- Minimal additional infrastructure requirements

## Out of Scope

### Explicitly NOT Building
- Real-time WebSocket updates (using polling approach)
- Task modification/deletion capabilities
- Advanced filtering (by status, date range, etc.)
- User authentication/authorization
- Task retry mechanisms
- Bulk operations
- Export functionality
- Task scheduling features

## Dependencies

### External Dependencies
- Existing MySQL database with `workflow_histories` table
- RabbitMQ queue system currently operational
- GORM ORM library already integrated

### Internal Dependencies
- Backend team: Database integration implementation
- Frontend team: API consumption and UI updates
- DevOps: Database migration if schema changes needed

## Risk Assessment

### High Risk
- **Database Transaction Complexity**: Queue + DB atomicity requires careful implementation
- **Performance Impact**: Additional database operations may slow queue processing

### Medium Risk
- **Data Volume Growth**: Task history accumulation may require archiving strategy
- **API Load**: Frontend polling may create database load

### Mitigation Strategies
- Implement proper database indexing on `repository_name` and `created_at`
- Use connection pooling and prepared statements
- Monitor query performance and optimize as needed
- Consider background cleanup for old task records

## Acceptance Criteria

### Backend Integration
- [ ] Task queuing triggers database insertion
- [ ] Failed database insertion prevents queue operation
- [ ] All task metadata is accurately stored
- [ ] Transaction rollback works correctly on failures

### API Implementation
- [ ] GET endpoint returns correct data for valid repository names
- [ ] Pagination works correctly with page/limit parameters
- [ ] Response format matches specification exactly
- [ ] Appropriate HTTP status codes for all scenarios
- [ ] Input validation prevents invalid queries

### Integration Testing
- [ ] End-to-end flow: Queue → Database → API → Frontend
- [ ] Error scenarios handled gracefully
- [ ] Performance requirements met under normal load
- [ ] No data corruption or loss during concurrent operations