# Task #70 Analysis: Queue-Database Atomic Integration

## Overview
Implement atomic queue + database operations for workflow task tracking. This analysis identifies the current architecture and provides a concrete implementation strategy.

## Current Backend Structure

### Architecture Conflict Identified
The backend directory contains two different server architectures:
1. **Simple Gin server**: `/backend/main.go` - Basic structure
2. **Clean architecture**: `/backend/cmd/server/main.go` - DI container with proper separation

**Recommendation**: Use the clean architecture approach for better maintainability.

### Key Integration Points
- **Queue Handler**: `/backend/internal/delivery/http/handlers/claude.go:86-91`
- **Current Flow**: HTTP Request → Handler → Queue Publisher → RabbitMQ
- **Required Flow**: HTTP Request → Handler → Atomic Service → [DB Transaction + Queue] → Response

### Model Location Issue
- **Current**: WorkflowHistories model in `/local-backend/utils/db/mysql/gormDB.go`
- **Problem**: Not available in backend/ directory structure
- **Solution**: Create unified model in `backend/internal/infrastructure/database/models/`

## Implementation Strategy

### 1. Model Unification (2 hours)
Create `backend/internal/infrastructure/database/models/workflow_history.go`:
```go
type WorkflowHistory struct {
    ID               uint64     `json:"id" gorm:"primaryKey;column:id"`
    RequestID        string     `json:"request_id" gorm:"column:request_id;uniqueIndex"`
    Status           string     `json:"status" gorm:"column:status;index;default:pending"`
    Tasks            string     `json:"tasks" gorm:"column:tasks;type:text;not null"`
    RepositoryName   string     `json:"repository_name" gorm:"column:repository_name;index;not null"`
    WorkingDir       *string    `json:"working_dir,omitempty" gorm:"column:working_dir"`
    ClaudeCmd        *string    `json:"claude_cmd,omitempty" gorm:"column:claude_cmd"`
    Interactive      bool       `json:"interactive" gorm:"column:interactive;default:false"`
    ContinueTask     bool       `json:"continue_task" gorm:"column:continue_task;default:false"`
    CreatedAt        time.Time  `json:"created_at" gorm:"column:created_at;index"`
    CompletedAt      *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
    ProcessingTimeMs *int64     `json:"processing_time_ms,omitempty" gorm:"column:processing_time_ms"`
    Result           *string    `json:"result,omitempty" gorm:"column:result;type:text"`
    Error            *string    `json:"error,omitempty" gorm:"column:error;type:text"`
}
```

### 2. Atomic Service Implementation (6 hours)
Create `backend/internal/application/services/atomic_queue_service.go`:
```go
type AtomicQueueService struct {
    db        *gorm.DB
    publisher *publisher.RabbitMQPublisher
}

func (s *AtomicQueueService) PublishWithHistory(ctx context.Context, msg publisher.WorkflowMessage) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. Create workflow history record
        history := &models.WorkflowHistory{
            RequestID:      uuid.New().String(),
            Status:         "pending",
            Tasks:          msg.Tasks,
            RepositoryName: msg.RepositoryName,
            // ... other fields
        }
        
        if err := tx.Create(history).Error; err != nil {
            return fmt.Errorf("failed to create history: %w", err)
        }
        
        // 2. Publish to queue (with rollback on failure)
        if err := s.publisher.PublishWorkflowMessage(msg); err != nil {
            return fmt.Errorf("failed to publish message: %w", err)
        }
        
        return nil
    })
}
```

### 3. Dependencies and Setup (2 hours)
Add to `backend/go.mod`:
```go
require (
    github.com/google/uuid v1.3.0
)
```

Update dependency injection in `backend/cmd/server/main.go`:
```go
// Add atomic service to container
atomicQueueService := services.NewAtomicQueueService(db, publisher)
```

### 4. Handler Integration (2 hours)
Update `backend/internal/delivery/http/handlers/claude.go`:
```go
func (h *claudeHandler) RunTasks(c *gin.Context) {
    // ... existing validation code ...
    
    // Use atomic service instead of direct publisher
    if err := h.atomicQueueService.PublishWithHistory(c.Request.Context(), message); err != nil {
        c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to queue task: %v", err)})
        return
    }
    
    c.JSON(200, gin.H{
        "message":    "Task queued successfully",
        "request_id": message.RequestID,
    })
}
```

## Risk Assessment

### High Risk
- **Transaction Boundary**: RabbitMQ operates outside GORM transactions
- **Rollback Complexity**: Failed queue operations need proper cleanup

### Medium Risk  
- **UUID Generation**: Must replace existing timestamp-based ID system
- **Model Synchronization**: Keep models consistent between backend/local-backend

### Mitigation Strategies
1. **Compensating Transactions**: Implement cleanup for failed queue operations
2. **Idempotency**: Use request IDs to prevent duplicate processing
3. **Comprehensive Testing**: Unit tests for success/failure scenarios
4. **Monitoring**: Add logging for transaction success/failure rates

## Files to Modify
1. `backend/internal/infrastructure/database/models/workflow_history.go` (NEW)
2. `backend/internal/application/services/atomic_queue_service.go` (NEW)
3. `backend/internal/delivery/http/handlers/claude.go` (MODIFY)
4. `backend/cmd/server/main.go` (MODIFY - DI setup)
5. `backend/go.mod` (MODIFY - add UUID dependency)

## Testing Strategy
1. **Unit Tests**: Atomic service with mocked dependencies
2. **Integration Tests**: Full HTTP → DB → Queue flow
3. **Failure Tests**: Database/queue failure scenarios
4. **Performance Tests**: Transaction overhead measurement

## Success Criteria
- [ ] 100% atomic operations (queue + database succeed or both fail)
- [ ] UUID v4 request IDs generated correctly
- [ ] All task metadata stored in database
- [ ] Proper error handling and rollback
- [ ] No breaking changes to existing queue functionality
- [ ] Performance impact <50ms additional overhead

## Next Steps
1. Set up database connection in backend/ (if missing)
2. Create WorkflowHistory model
3. Implement AtomicQueueService
4. Update handlers to use atomic service
5. Add comprehensive tests
6. Performance validation