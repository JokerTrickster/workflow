---
name: task-history-api
status: backlog
created: 2025-09-22T03:27:42Z
progress: 0%
prd: .claude/prds/task-history-api.md
github: https://github.com/JokerTrickster/workflow/issues/69
---

# Epic: Task History API

## Overview

Implement database persistence for workflow tasks during queue insertion and create a REST API for repository-specific task history retrieval. This leverages existing infrastructure (MySQL with GORM, RabbitMQ, Echo/Gin router) to add minimal, focused functionality without over-engineering.

## Architecture Decisions

- **Database Strategy**: Use existing `workflow_histories` table and GORM models without schema changes
- **Transaction Pattern**: Implement atomic queue + database operations using database transactions
- **API Framework**: Extend existing Echo/Gin router with single endpoint for task history
- **Error Handling**: Fail-fast approach - if database insertion fails, queue operation must fail completely
- **Performance**: Leverage existing database connection pooling and add targeted indexes if needed

## Technical Approach

### Backend Services
- **Queue Integration**: Modify existing queue insertion logic in `backend/` folder to include database persistence
- **Database Layer**: Extend existing GORM setup to handle `WorkflowHistories` operations
- **API Endpoint**: Single GET endpoint `/api/tasks/history/{repository_name}` with pagination
- **Transaction Management**: Use GORM's transaction capabilities for atomic operations

### Frontend Components
- **API Client**: Extend existing API client to consume task history endpoint
- **Repository Filter**: Leverage existing repository management to filter task history
- **Pagination**: Standard pagination controls for task list
- **Polling**: Use existing polling patterns for status updates

### Infrastructure
- **Database**: No schema changes - use existing `workflow_histories` table
- **Indexing**: Add index on `repository_name` and `created_at` for query performance
- **Monitoring**: Extend existing logging for database operations and API performance

## Implementation Strategy

1. **Database Integration First**: Implement atomic queue + database operations
2. **API Development**: Create single task history endpoint with pagination
3. **Integration Testing**: End-to-end validation of queue → database → API flow
4. **Performance Validation**: Ensure <200ms API response times

## Task Breakdown Preview

High-level task categories (≤8 tasks total):
- [ ] **Queue-Database Integration**: Modify backend queue logic to persist tasks atomically
- [ ] **Database Transaction Handler**: Implement atomic queue + database operations with rollback
- [ ] **Task History API Endpoint**: Create GET `/api/tasks/history/{repo}` with pagination
- [ ] **Database Indexing**: Add performance indexes for repository and date queries
- [ ] **Frontend API Integration**: Update frontend to consume task history endpoint
- [ ] **Error Handling**: Implement comprehensive error responses and logging
- [ ] **Integration Testing**: End-to-end testing of complete workflow
- [ ] **Performance Optimization**: Validate and optimize query performance

## Dependencies

### External Dependencies
- Existing MySQL database with `workflow_histories` table (✅ Available)
- RabbitMQ queue system (✅ Operational)
- GORM ORM library (✅ Integrated)

### Internal Dependencies
- Backend queue implementation in `backend/` folder
- Frontend repository management system
- Existing database connection and transaction handling

## Success Criteria (Technical)

### Performance Benchmarks
- API response time: <200ms for 95% of requests (20-50 tasks)
- Database insertion: <50ms additional overhead per queue operation
- Concurrent operations: Support 10+ simultaneous queue operations

### Quality Gates
- 100% database insertion success rate for queued tasks
- Zero data loss between queue and database operations
- <1% API error rate under normal load
- Complete transaction rollback on any failure

### Acceptance Criteria
- Atomic queue + database operations with proper rollback
- Repository-specific task filtering with pagination
- Proper HTTP status codes and error handling
- Integration with existing frontend repository system

## Estimated Effort

### Overall Timeline
- **Development**: 3-4 days (backend integration: 1.5 days, API: 1 day, testing: 1.5 days)
- **Critical Path**: Queue-database transaction implementation
- **Resource Requirements**: 1 backend developer, minimal frontend changes

### Risk Mitigation
- **Transaction Complexity**: Use existing GORM transaction patterns, extensive testing
- **Performance Impact**: Implement with existing connection pool, monitor query performance
- **Integration Risk**: Leverage existing patterns, avoid architectural changes

## Tasks Created
- [ ] 001.md - Queue-Database Atomic Integration (parallel: false, 12h)
- [ ] 002.md - Database Performance Indexing (parallel: false, 4h)  
- [ ] 003.md - Task History API Endpoint (parallel: true, 8h)
- [ ] 004.md - Comprehensive Error Handling (parallel: true, 6h)
- [ ] 005.md - Frontend API Integration (parallel: true, 10h)
- [ ] 006.md - Integration Testing (parallel: false, 14h)
- [ ] 007.md - Performance Optimization (parallel: true, 8h)

Total tasks: 7
Parallel tasks: 4 (003, 004, 005, 007)
Sequential tasks: 3 (001, 002, 006)
Estimated total effort: 62 hours