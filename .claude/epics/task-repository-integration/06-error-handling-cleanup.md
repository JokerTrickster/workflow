---
name: error-handling-cleanup
epic: task-repository-integration
type: infrastructure
priority: high
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 2
complexity: medium
dependencies: [branch-management, pr-automation]
---

# Task: Error Handling & Cleanup

## Overview
Implement comprehensive error handling and cleanup mechanisms for the Git workflow integration. This ensures system reliability by properly handling failures at any stage and maintaining clean repository states even when operations fail.

## Acceptance Criteria

### Comprehensive Error Recovery
- [ ] **GitHub API Failures**: Graceful handling when GitHub is unavailable or rate-limited
- [ ] **Git Operation Failures**: Recovery from failed branch operations, commits, or pushes
- [ ] **Network Issues**: Resilient behavior during connectivity problems
- [ ] **Permission Errors**: Clear error messages and recovery for access issues

### Cleanup Orchestration
- [ ] **Failed Task Cleanup**: Remove branches and issues for failed or cancelled tasks
- [ ] **Orphaned Resource Detection**: Identify and clean up abandoned branches and issues
- [ ] **Cleanup Status Tracking**: Monitor and report cleanup operation status
- [ ] **Manual Cleanup Support**: Tools for administrators to trigger cleanup operations

### System Resilience
- [ ] **Graceful Degradation**: Task execution continues even when Git workflow fails
- [ ] **State Consistency**: Prevent partial states that leave repositories inconsistent
- [ ] **Retry Logic**: Automatic retry for transient failures with backoff
- [ ] **Circuit Breaker**: Temporarily disable Git features during extended outages

### Monitoring & Alerting
- [ ] **Error Classification**: Categorize errors for appropriate response
- [ ] **Cleanup Metrics**: Track cleanup success rates and performance
- [ ] **Alert Generation**: Notify administrators of critical failures
- [ ] **Health Checks**: Regular validation of Git workflow system health

## Implementation Details

### Error Classification System
```go
type ErrorType int

const (
    ErrorTypeTransient ErrorType = iota // Retry-able errors
    ErrorTypePermanent                   // Non-retry-able errors
    ErrorTypeConfiguration               // Config or permission issues
    ErrorTypeNetwork                     // Network connectivity issues
)

type WorkflowError struct {
    Type      ErrorType
    Operation string
    Message   string
    Cause     error
    Retryable bool
    RequestID string
}

func (e *WorkflowError) Error() string {
    return fmt.Sprintf("%s failed: %s (type: %d, retryable: %t)",
        e.Operation, e.Message, e.Type, e.Retryable)
}

func ClassifyError(err error, operation string) *WorkflowError {
    if err == nil {
        return nil
    }

    workflowErr := &WorkflowError{
        Operation: operation,
        Cause:     err,
        Message:   err.Error(),
    }

    // GitHub API errors
    if isGitHubRateLimit(err) {
        workflowErr.Type = ErrorTypeTransient
        workflowErr.Retryable = true
        workflowErr.Message = "GitHub API rate limit exceeded"
    } else if isGitHubUnavailable(err) {
        workflowErr.Type = ErrorTypeNetwork
        workflowErr.Retryable = true
        workflowErr.Message = "GitHub API unavailable"
    } else if isGitHubPermission(err) {
        workflowErr.Type = ErrorTypeConfiguration
        workflowErr.Retryable = false
        workflowErr.Message = "GitHub permission denied"
    }

    // Git operation errors
    if isGitNetworkError(err) {
        workflowErr.Type = ErrorTypeNetwork
        workflowErr.Retryable = true
    } else if isGitPermissionError(err) {
        workflowErr.Type = ErrorTypeConfiguration
        workflowErr.Retryable = false
    } else if isGitConflictError(err) {
        workflowErr.Type = ErrorTypePermanent
        workflowErr.Retryable = false
    }

    return workflowErr
}
```

### Retry Logic with Exponential Backoff
```go
type RetryConfig struct {
    MaxAttempts int           `json:"max_attempts"`
    BaseDelay   time.Duration `json:"base_delay"`
    MaxDelay    time.Duration `json:"max_delay"`
    Multiplier  float64       `json:"multiplier"`
}

func (r *RetryConfig) Execute(ctx context.Context, operation string, fn func() error) error {
    var lastErr error

    for attempt := 1; attempt <= r.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        workflowErr := ClassifyError(err, operation)
        lastErr = workflowErr

        if !workflowErr.Retryable {
            log.Printf("Non-retryable error in %s: %v", operation, err)
            return workflowErr
        }

        if attempt == r.MaxAttempts {
            log.Printf("Max retry attempts reached for %s: %v", operation, err)
            break
        }

        // Calculate delay with exponential backoff
        delay := time.Duration(float64(r.BaseDelay) * math.Pow(r.Multiplier, float64(attempt-1)))
        if delay > r.MaxDelay {
            delay = r.MaxDelay
        }

        log.Printf("Retrying %s in %v (attempt %d/%d): %v",
            operation, delay, attempt+1, r.MaxAttempts, err)

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return lastErr
}
```

### Cleanup Orchestrator
```go
type CleanupOrchestrator struct {
    GitHubService    *github.Service
    BranchManager    *BranchManager
    DB              *sql.DB
    RetryConfig     *RetryConfig
}

func (co *CleanupOrchestrator) CleanupFailedTask(ctx context.Context, requestID string) error {
    // Get task information
    taskInfo, err := co.getTaskInfo(requestID)
    if err != nil {
        return fmt.Errorf("failed to get task info: %w", err)
    }

    log.Printf("Starting cleanup for failed task %s", requestID)

    // Update cleanup status
    co.updateCleanupStatus(requestID, "in_progress")

    var cleanupErrors []error

    // 1. Clean up branch if it exists
    if taskInfo.BranchName != "" {
        err := co.RetryConfig.Execute(ctx, "branch_cleanup", func() error {
            return co.BranchManager.CleanupBranch(taskInfo.BranchName)
        })
        if err != nil {
            cleanupErrors = append(cleanupErrors, fmt.Errorf("branch cleanup failed: %w", err))
        }
    }

    // 2. Close GitHub issue if it exists
    if taskInfo.GitHubIssueURL != "" {
        issueNumber := extractIssueNumber(taskInfo.GitHubIssueURL)
        if issueNumber > 0 {
            err := co.RetryConfig.Execute(ctx, "issue_cleanup", func() error {
                return co.GitHubService.CloseIssue("JokerTrickster", taskInfo.RepositoryName, issueNumber)
            })
            if err != nil {
                cleanupErrors = append(cleanupErrors, fmt.Errorf("issue cleanup failed: %w", err))
            }
        }
    }

    // 3. Close PR if it exists and wasn't merged
    if taskInfo.GitHubPRURL != "" {
        prNumber := extractPRNumber(taskInfo.GitHubPRURL)
        if prNumber > 0 {
            err := co.RetryConfig.Execute(ctx, "pr_cleanup", func() error {
                return co.GitHubService.ClosePR("JokerTrickster", taskInfo.RepositoryName, prNumber)
            })
            if err != nil {
                cleanupErrors = append(cleanupErrors, fmt.Errorf("PR cleanup failed: %w", err))
            }
        }
    }

    // Update final status
    if len(cleanupErrors) == 0 {
        co.updateCleanupStatus(requestID, "completed")
        log.Printf("Cleanup completed successfully for task %s", requestID)
    } else {
        co.updateCleanupStatus(requestID, "failed")
        log.Printf("Cleanup completed with errors for task %s: %v", requestID, cleanupErrors)
        return fmt.Errorf("cleanup had %d errors: %v", len(cleanupErrors), cleanupErrors)
    }

    return nil
}
```

### Orphaned Resource Detection
```go
func (co *CleanupOrchestrator) DetectOrphanedResources(ctx context.Context) (*OrphanReport, error) {
    report := &OrphanReport{
        Timestamp: time.Now(),
    }

    // Find orphaned branches (older than 7 days with no recent activity)
    orphanedBranches, err := co.findOrphanedBranches(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to find orphaned branches: %w", err)
    }
    report.OrphanedBranches = orphanedBranches

    // Find orphaned issues (no corresponding workflow record)
    orphanedIssues, err := co.findOrphanedIssues(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to find orphaned issues: %w", err)
    }
    report.OrphanedIssues = orphanedIssues

    // Find stalled cleanups (cleanup_status = 'in_progress' for > 1 hour)
    stalledCleanups, err := co.findStalledCleanups()
    if err != nil {
        return nil, fmt.Errorf("failed to find stalled cleanups: %w", err)
    }
    report.StalledCleanups = stalledCleanups

    return report, nil
}

func (co *CleanupOrchestrator) findOrphanedBranches(ctx context.Context) ([]OrphanedBranch, error) {
    // Get all repositories
    repos := []string{} // Load from configuration
    var orphaned []OrphanedBranch

    for _, repo := range repos {
        repoPath := utils.GetRepositoryPath(repo)

        // Get all branches with task- prefix
        branches, err := co.getTaskBranches(repoPath)
        if err != nil {
            log.Printf("Failed to get branches for %s: %v", repo, err)
            continue
        }

        for _, branch := range branches {
            // Check if branch has corresponding workflow record
            exists, err := co.hasWorkflowRecord(branch.Name)
            if err != nil {
                log.Printf("Failed to check workflow record for branch %s: %v", branch.Name, err)
                continue
            }

            // Consider orphaned if no workflow record and older than threshold
            if !exists && time.Since(branch.LastCommit) > 7*24*time.Hour {
                orphaned = append(orphaned, OrphanedBranch{
                    Repository: repo,
                    Name:       branch.Name,
                    Age:        time.Since(branch.LastCommit),
                    LastCommit: branch.LastCommit,
                })
            }
        }
    }

    return orphaned, nil
}
```

### Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    mu          sync.RWMutex
    state       CircuitState
    failures    int
    lastFailure time.Time
    threshold   int
    timeout     time.Duration
}

type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)

func (cb *CircuitBreaker) Execute(operation func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()

    switch state {
    case StateOpen:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.mu.Unlock()
        } else {
            return fmt.Errorf("circuit breaker open")
        }
    case StateHalfOpen:
        // Allow one attempt
    case StateClosed:
        // Normal operation
    }

    err := operation()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.failures >= cb.threshold {
            cb.state = StateOpen
        }
        return err
    }

    // Success - reset or close circuit
    cb.failures = 0
    cb.state = StateClosed
    return nil
}
```

### Health Check System
```go
type HealthChecker struct {
    GitHubService *github.Service
    DB           *sql.DB
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthReport {
    report := &HealthReport{
        Timestamp: time.Now(),
        Checks:    make(map[string]HealthCheck),
    }

    // Check GitHub API connectivity
    report.Checks["github_api"] = hc.checkGitHubAPI(ctx)

    // Check database connectivity
    report.Checks["database"] = hc.checkDatabase(ctx)

    // Check Git CLI availability
    report.Checks["git_cli"] = hc.checkGitCLI(ctx)

    // Check repository access
    report.Checks["repository_access"] = hc.checkRepositoryAccess(ctx)

    // Check cleanup queue health
    report.Checks["cleanup_queue"] = hc.checkCleanupQueue(ctx)

    // Determine overall health
    report.Overall = hc.calculateOverallHealth(report.Checks)

    return report
}

type HealthCheck struct {
    Status  HealthStatus `json:"status"`
    Message string       `json:"message"`
    Latency time.Duration `json:"latency"`
    Error   string       `json:"error,omitempty"`
}

type HealthStatus string

const (
    HealthStatusHealthy   HealthStatus = "healthy"
    HealthStatusDegraded  HealthStatus = "degraded"
    HealthStatusUnhealthy HealthStatus = "unhealthy"
)
```

## Testing Requirements

### Unit Tests
- [ ] Error classification for various failure types
- [ ] Retry logic with different backoff configurations
- [ ] Cleanup orchestration for failed tasks
- [ ] Circuit breaker state transitions
- [ ] Health check implementations

### Integration Tests
- [ ] End-to-end error handling scenarios
- [ ] Orphaned resource detection and cleanup
- [ ] System resilience during GitHub outages
- [ ] Recovery from various failure modes

### Chaos Testing
- [ ] Network partition scenarios
- [ ] GitHub API rate limiting
- [ ] Database connectivity issues
- [ ] Git operation failures

## Definition of Done
- [ ] Comprehensive error handling covers all failure modes
- [ ] Cleanup operations maintain repository consistency
- [ ] System gracefully degrades during outages
- [ ] Orphaned resources are automatically detected and cleaned
- [ ] Health monitoring provides operational visibility
- [ ] Retry logic reduces transient failure impact
- [ ] Circuit breaker prevents cascade failures
- [ ] All error scenarios tested and documented

## Dependencies
- Branch Management (04) - for cleanup operations
- PR Automation (05) - for PR-related error handling
- Database schema for tracking cleanup status
- Monitoring infrastructure for alerts

## Notes
Error handling and cleanup are critical for system reliability. Focus on preventing inconsistent states and providing clear operational visibility. The system should be resilient to external service failures while maintaining core functionality.