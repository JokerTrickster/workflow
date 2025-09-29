---
name: database-schema-update
epic: task-repository-integration
type: infrastructure
priority: high
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 1
complexity: low
dependencies: []
---

# Task: Database Schema Update

## Overview
Add Git workflow tracking fields to the existing `workflow_histories` table to support GitHub issue links, PR references, branch tracking, and cleanup status. This enables full audit trail and status tracking for the automated Git workflow without requiring new tables or complex schema changes.

## Acceptance Criteria

### Schema Changes
- [ ] **GitHub Issue URL**: Add `github_issue_url` VARCHAR(500) field for issue tracking
- [ ] **GitHub PR URL**: Add `github_pr_url` VARCHAR(500) field for PR tracking
- [ ] **Branch Name**: Add `branch_name` VARCHAR(255) field for Git branch tracking
- [ ] **Cleanup Status**: Add `cleanup_status` ENUM field for tracking cleanup completion
- [ ] **Git Operations Log**: Add `git_operations` JSON field for detailed Git operation logging

### Migration Safety
- [ ] **Backward Compatibility**: Existing queries continue to work unchanged
- [ ] **Default Values**: All new fields have appropriate default values
- [ ] **Rollback Plan**: Migration can be safely rolled back if needed
- [ ] **Data Validation**: Constraints ensure data integrity

### GORM Model Updates
- [ ] **Model Extension**: Update `WorkflowHistories` struct with new fields
- [ ] **JSON Tags**: Proper JSON serialization for API responses
- [ ] **Validation Tags**: Input validation for new fields
- [ ] **Index Creation**: Appropriate indexes for query performance

## Implementation Details

### Database Migration
```sql
-- Add new columns to workflow_histories table
ALTER TABLE workflow_histories
ADD COLUMN github_issue_url VARCHAR(500) DEFAULT NULL,
ADD COLUMN github_pr_url VARCHAR(500) DEFAULT NULL,
ADD COLUMN branch_name VARCHAR(255) DEFAULT NULL,
ADD COLUMN cleanup_status ENUM('pending', 'in_progress', 'completed', 'failed') DEFAULT 'pending',
ADD COLUMN git_operations JSON DEFAULT NULL;

-- Add indexes for performance
CREATE INDEX idx_workflow_histories_github_issue ON workflow_histories(github_issue_url);
CREATE INDEX idx_workflow_histories_branch_name ON workflow_histories(branch_name);
CREATE INDEX idx_workflow_histories_cleanup_status ON workflow_histories(cleanup_status);
```

### GORM Model Update
```go
type WorkflowHistories struct {
    RequestID        string    `json:"request_id" gorm:"primaryKey;size:255"`
    Status           string    `json:"status" gorm:"size:50"`
    Tasks            string    `json:"tasks" gorm:"type:text"`
    RepositoryName   string    `json:"repository_name" gorm:"size:255"`
    WorkingDir       string    `json:"working_dir" gorm:"size:500"`
    ClaudeCmd        string    `json:"claude_cmd" gorm:"size:500"`
    Interactive      bool      `json:"interactive"`
    ContinueTask     bool      `json:"continue_task"`
    CreatedAt        time.Time `json:"created_at"`
    CompletedAt      *time.Time `json:"completed_at"`
    ProcessingTimeMs *int64    `json:"processing_time_ms"`
    Result           string    `json:"result" gorm:"type:json"`
    Error            string    `json:"error" gorm:"type:text"`

    // New Git workflow fields
    GitHubIssueURL   *string   `json:"github_issue_url" gorm:"size:500"`
    GitHubPRURL      *string   `json:"github_pr_url" gorm:"size:500"`
    BranchName       *string   `json:"branch_name" gorm:"size:255"`
    CleanupStatus    string    `json:"cleanup_status" gorm:"type:enum('pending','in_progress','completed','failed');default:'pending'"`
    GitOperations    *string   `json:"git_operations" gorm:"type:json"`
}
```

### Migration Script
Create migration file: `local-backend/migrations/add_git_workflow_fields.sql`

```sql
-- Migration: Add Git workflow tracking fields
-- Version: 20250929_git_workflow_integration
-- Description: Add GitHub integration and Git workflow tracking to workflow_histories

-- Add new columns with appropriate constraints
ALTER TABLE workflow_histories
ADD COLUMN github_issue_url VARCHAR(500) DEFAULT NULL
    COMMENT 'GitHub issue URL for task tracking',
ADD COLUMN github_pr_url VARCHAR(500) DEFAULT NULL
    COMMENT 'GitHub pull request URL for code review',
ADD COLUMN branch_name VARCHAR(255) DEFAULT NULL
    COMMENT 'Git branch name used for task execution',
ADD COLUMN cleanup_status ENUM('pending', 'in_progress', 'completed', 'failed') DEFAULT 'pending'
    COMMENT 'Status of cleanup operations (branch deletion, issue closure)',
ADD COLUMN git_operations JSON DEFAULT NULL
    COMMENT 'Detailed log of Git operations performed during task execution';

-- Add performance indexes
CREATE INDEX idx_workflow_histories_github_issue ON workflow_histories(github_issue_url);
CREATE INDEX idx_workflow_histories_github_pr ON workflow_histories(github_pr_url);
CREATE INDEX idx_workflow_histories_branch_name ON workflow_histories(branch_name);
CREATE INDEX idx_workflow_histories_cleanup_status ON workflow_histories(cleanup_status);

-- Add constraints for data validation
ALTER TABLE workflow_histories
ADD CONSTRAINT chk_github_issue_url_format
    CHECK (github_issue_url IS NULL OR github_issue_url REGEXP '^https://github.com/[^/]+/[^/]+/issues/[0-9]+$'),
ADD CONSTRAINT chk_github_pr_url_format
    CHECK (github_pr_url IS NULL OR github_pr_url REGEXP '^https://github.com/[^/]+/[^/]+/pull/[0-9]+$'),
ADD CONSTRAINT chk_branch_name_format
    CHECK (branch_name IS NULL OR LENGTH(branch_name) >= 3);
```

### Rollback Script
Create rollback file: `local-backend/migrations/rollback_git_workflow_fields.sql`

```sql
-- Rollback: Remove Git workflow tracking fields
-- Version: 20250929_git_workflow_integration_rollback

-- Drop indexes first
DROP INDEX IF EXISTS idx_workflow_histories_cleanup_status ON workflow_histories;
DROP INDEX IF EXISTS idx_workflow_histories_branch_name ON workflow_histories;
DROP INDEX IF EXISTS idx_workflow_histories_github_pr ON workflow_histories;
DROP INDEX IF EXISTS idx_workflow_histories_github_issue ON workflow_histories;

-- Drop constraints
ALTER TABLE workflow_histories
DROP CONSTRAINT IF EXISTS chk_branch_name_format,
DROP CONSTRAINT IF EXISTS chk_github_pr_url_format,
DROP CONSTRAINT IF EXISTS chk_github_issue_url_format;

-- Remove columns
ALTER TABLE workflow_histories
DROP COLUMN IF EXISTS git_operations,
DROP COLUMN IF EXISTS cleanup_status,
DROP COLUMN IF EXISTS branch_name,
DROP COLUMN IF EXISTS github_pr_url,
DROP COLUMN IF EXISTS github_issue_url;
```

## Testing Requirements

### Migration Testing
- [ ] **Forward Migration**: Test migration on copy of production data
- [ ] **Rollback Testing**: Verify rollback works cleanly
- [ ] **Data Integrity**: Existing data remains unchanged
- [ ] **Performance Impact**: Migration completes within acceptable time
- [ ] **Index Performance**: New indexes improve query performance

### GORM Model Testing
- [ ] **Model Validation**: All fields properly serialize/deserialize
- [ ] **Database Operations**: CRUD operations work with new fields
- [ ] **Null Handling**: Nullable fields handle NULL values correctly
- [ ] **JSON Serialization**: API responses include new fields properly

### Integration Testing
- [ ] **Existing Functionality**: All current features continue working
- [ ] **New Field Population**: New fields can be populated during task execution
- [ ] **Query Performance**: Database queries perform acceptably with new indexes

## Definition of Done
- [ ] Migration scripts created and tested
- [ ] GORM model updated with new fields
- [ ] All existing tests continue to pass
- [ ] New fields can be populated and queried
- [ ] Performance impact is minimal
- [ ] Rollback procedure verified
- [ ] Documentation updated

## Dependencies
- Database access for migration execution
- Test database for migration testing
- GORM model definitions in local-backend

## Notes
- Keep migration simple and focused on adding fields only
- Ensure all new fields are optional to maintain backward compatibility
- Use appropriate data types to prevent performance issues
- Plan for future Git workflow features in field sizing