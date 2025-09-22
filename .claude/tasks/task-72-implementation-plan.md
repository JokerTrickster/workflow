# Task #72: Task History API Endpoint Implementation Plan

## Overview
Implement GET `/api/tasks/history/{repository_name}` with pagination support using Gin router, leveraging existing WorkflowHistory model and comprehensive error handling.

## Implementation Steps

### 1. Database Service Layer
- Create repository pattern for WorkflowHistory queries
- Implement pagination logic with offset/limit
- Add repository name filtering and date sorting

### 2. API Handler
- Create handler function for task history endpoint
- Implement input validation for repository name and pagination params
- Add proper error handling using existing middleware

### 3. Router Integration
- Add new route to tasks group in main.go
- Integrate with existing auth middleware

### 4. Response Format
- Implement exact JSON response format from PRD
- Include data array and pagination metadata
- Handle empty results appropriately

### 5. Testing
- Unit tests for service layer
- Integration tests for API endpoint
- Performance validation

## Technical Details

### Database Query Pattern
```go
// Repository filtering + pagination + sorting
query := db.Where("repository_name = ?", repositoryName).
         Order("created_at DESC").
         Offset((page - 1) * limit).
         Limit(limit)
```

### Response Format (from PRD)
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

### Error Handling
- Use existing middleware.HandleError() for standardized error responses
- Validation errors: 400 Bad Request
- Repository not found: 404 Not Found
- Database errors: 500 Internal Server Error

## File Changes Required
1. Create: `backend/internal/services/task_history_service.go`
2. Create: `backend/internal/handlers/task_history.go`
3. Modify: `backend/main.go` (add route)
4. Create: Tests for service and handler

## Success Criteria
- [x] Endpoint functional: GET `/api/tasks/history/{repository_name}`
- [x] Pagination with page/limit parameters
- [x] Tasks sorted by created_at DESC
- [x] Pagination metadata in response
- [x] Input validation for all parameters
- [x] Proper HTTP status codes (200, 400, 404, 500)
- [x] Response format matches PRD exactly
- [x] Performance <200ms for typical queries
- [x] Unit and integration tests