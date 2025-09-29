---
name: branch-management
epic: task-repository-integration
type: core
priority: high
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 3
complexity: medium
dependencies: [database-schema-update]
---

# Task: Branch Management

## Overview
Implement robust Git branch management for task isolation, including branch creation with unique naming, automatic cleanup, and conflict prevention. This ensures that concurrent tasks don't interfere with each other and provides a clean Git history for code review.

## Acceptance Criteria

### Branch Creation & Naming
- [ ] **Unique Branch Names**: Generate collision-resistant branch names using timestamp and task hash
- [ ] **Naming Convention**: Consistent format: `task-{timestamp}-{hash}` for easy identification
- [ ] **Default Branch Detection**: Automatically detect and use repository's default branch (main/master)
- [ ] **Branch Validation**: Ensure branch names comply with Git naming rules

### Branch Isolation
- [ ] **Clean Working State**: Ensure working directory is clean before branch creation
- [ ] **Latest Changes**: Pull latest changes from default branch before creating feature branch
- [ ] **Workspace Isolation**: Each task executes in its own branch without affecting others
- [ ] **Concurrent Support**: Multiple tasks can work on same repository simultaneously

### Branch Lifecycle Management
- [ ] **Automatic Cleanup**: Delete feature branches after successful PR merge
- [ ] **Failed Task Cleanup**: Clean up branches for failed or cancelled tasks
- [ ] **Orphan Detection**: Identify and clean up abandoned branches
- [ ] **Cleanup Status Tracking**: Track cleanup completion in database

### Git Operations Robustness
- [ ] **Authentication Handling**: Work with existing Git authentication setup
- [ ] **Network Resilience**: Handle network failures during Git operations
- [ ] **Permission Validation**: Verify Git repository permissions before operations
- [ ] **State Recovery**: Ability to recover from interrupted Git operations

## Implementation Details

### Branch Naming Strategy
```go
type BranchManager struct {
    RepositoryPath string
    GitOperations  *GitOperations
}

func (bm *BranchManager) GenerateBranchName(taskDescription string) string {
    // Use timestamp for uniqueness
    timestamp := time.Now().Format("20060102-150405")

    // Generate short hash from task description
    hasher := sha256.New()
    hasher.Write([]byte(taskDescription))
    hash := hex.EncodeToString(hasher.Sum(nil))[:8]

    // Sanitize for Git branch name compliance
    branchName := fmt.Sprintf("task-%s-%s", timestamp, hash)
    return sanitizeBranchName(branchName)
}

func sanitizeBranchName(name string) string {
    // Remove invalid characters and ensure Git compliance
    reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
    sanitized := reg.ReplaceAllString(name, "-")

    // Ensure doesn't start/end with special characters
    sanitized = strings.Trim(sanitized, "-_")

    // Limit length to reasonable size
    if len(sanitized) > 50 {
        sanitized = sanitized[:50]
    }

    return sanitized
}
```

### Branch Creation Workflow
```go
func (bm *BranchManager) CreateFeatureBranch(branchName string) error {
    // 1. Ensure clean working state
    if err := bm.ensureCleanWorkingState(); err != nil {
        return fmt.Errorf("working directory not clean: %w", err)
    }

    // 2. Switch to default branch
    defaultBranch, err := bm.getDefaultBranch()
    if err != nil {
        return fmt.Errorf("failed to detect default branch: %w", err)
    }

    if err := bm.checkoutBranch(defaultBranch); err != nil {
        return fmt.Errorf("failed to checkout default branch: %w", err)
    }

    // 3. Pull latest changes
    if err := bm.pullLatestChanges(); err != nil {
        log.Printf("Warning: failed to pull latest changes: %v", err)
        // Continue anyway - might be offline or other non-critical issue
    }

    // 4. Create and checkout feature branch
    if err := bm.createAndCheckoutBranch(branchName); err != nil {
        return fmt.Errorf("failed to create branch %s: %w", branchName, err)
    }

    log.Printf("Successfully created and switched to branch: %s", branchName)
    return nil
}

func (bm *BranchManager) ensureCleanWorkingState() error {
    cmd := exec.Command("git", "status", "--porcelain")
    cmd.Dir = bm.RepositoryPath
    output, err := cmd.Output()
    if err != nil {
        return err
    }

    if len(strings.TrimSpace(string(output))) > 0 {
        return fmt.Errorf("working directory has uncommitted changes")
    }

    return nil
}

func (bm *BranchManager) getDefaultBranch() (string, error) {
    // Try to get default branch from remote
    cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
    cmd.Dir = bm.RepositoryPath
    output, err := cmd.Output()
    if err == nil {
        branch := strings.TrimSpace(string(output))
        if strings.HasPrefix(branch, "refs/remotes/origin/") {
            return strings.TrimPrefix(branch, "refs/remotes/origin/"), nil
        }
    }

    // Fallback: check common default branch names
    for _, branch := range []string{"main", "master", "develop"} {
        if bm.branchExists(branch) {
            return branch, nil
        }
    }

    return "", fmt.Errorf("could not determine default branch")
}

func (bm *BranchManager) createAndCheckoutBranch(branchName string) error {
    cmd := exec.Command("git", "checkout", "-b", branchName)
    cmd.Dir = bm.RepositoryPath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("git checkout -b failed: %w, output: %s", err, string(output))
    }
    return nil
}
```

### Branch Cleanup Operations
```go
func (bm *BranchManager) CleanupBranch(branchName string) error {
    // 1. Switch back to default branch
    defaultBranch, err := bm.getDefaultBranch()
    if err != nil {
        return err
    }

    if err := bm.checkoutBranch(defaultBranch); err != nil {
        return fmt.Errorf("failed to checkout default branch for cleanup: %w", err)
    }

    // 2. Delete local branch
    if err := bm.deleteLocalBranch(branchName); err != nil {
        log.Printf("Warning: failed to delete local branch %s: %v", branchName, err)
    }

    // 3. Delete remote branch (if it exists)
    if err := bm.deleteRemoteBranch(branchName); err != nil {
        log.Printf("Warning: failed to delete remote branch %s: %v", branchName, err)
    }

    log.Printf("Successfully cleaned up branch: %s", branchName)
    return nil
}

func (bm *BranchManager) deleteLocalBranch(branchName string) error {
    cmd := exec.Command("git", "branch", "-D", branchName)
    cmd.Dir = bm.RepositoryPath
    return cmd.Run()
}

func (bm *BranchManager) deleteRemoteBranch(branchName string) error {
    // Check if remote branch exists first
    cmd := exec.Command("git", "ls-remote", "--heads", "origin", branchName)
    cmd.Dir = bm.RepositoryPath
    output, err := cmd.Output()
    if err != nil || len(strings.TrimSpace(string(output))) == 0 {
        // Remote branch doesn't exist
        return nil
    }

    // Delete remote branch
    cmd = exec.Command("git", "push", "origin", "--delete", branchName)
    cmd.Dir = bm.RepositoryPath
    return cmd.Run()
}
```

### Concurrent Task Support
```go
type BranchRegistry struct {
    mu     sync.RWMutex
    active map[string]map[string]bool // repo -> branch -> active
}

func (br *BranchRegistry) RegisterBranch(repository, branch string) error {
    br.mu.Lock()
    defer br.mu.Unlock()

    if br.active == nil {
        br.active = make(map[string]map[string]bool)
    }

    if br.active[repository] == nil {
        br.active[repository] = make(map[string]bool)
    }

    if br.active[repository][branch] {
        return fmt.Errorf("branch %s already active in repository %s", branch, repository)
    }

    br.active[repository][branch] = true
    return nil
}

func (br *BranchRegistry) UnregisterBranch(repository, branch string) {
    br.mu.Lock()
    defer br.mu.Unlock()

    if br.active[repository] != nil {
        delete(br.active[repository], branch)
    }
}

func (br *BranchRegistry) GetActiveBranches(repository string) []string {
    br.mu.RLock()
    defer br.mu.RUnlock()

    var branches []string
    if br.active[repository] != nil {
        for branch := range br.active[repository] {
            branches = append(branches, branch)
        }
    }
    return branches
}
```

### Integration with Database Tracking
```go
func (bm *BranchManager) TrackBranchCreation(requestID, branchName string) error {
    // Update workflow_histories table with branch information
    query := `UPDATE workflow_histories
              SET branch_name = ?, cleanup_status = 'pending'
              WHERE request_id = ?`

    _, err := bm.DB.Exec(query, branchName, requestID)
    if err != nil {
        return fmt.Errorf("failed to track branch creation: %w", err)
    }

    return nil
}

func (bm *BranchManager) UpdateCleanupStatus(requestID, status string) error {
    query := `UPDATE workflow_histories
              SET cleanup_status = ?
              WHERE request_id = ?`

    _, err := bm.DB.Exec(query, status, requestID)
    return err
}
```

## Testing Requirements

### Unit Tests
- [ ] Branch name generation uniqueness and compliance
- [ ] Clean working state validation
- [ ] Default branch detection across different repositories
- [ ] Branch creation and checkout operations
- [ ] Cleanup operations for various scenarios

### Integration Tests
- [ ] Concurrent task execution with branch isolation
- [ ] Branch lifecycle from creation to cleanup
- [ ] Network failure handling during Git operations
- [ ] Repository permission validation
- [ ] Database tracking integration

### Performance Tests
- [ ] Branch creation time under load
- [ ] Cleanup operation performance
- [ ] Memory usage for branch registry
- [ ] Git operation timeout handling

## Definition of Done
- [ ] Branch management prevents task conflicts
- [ ] Unique branch naming ensures no collisions
- [ ] Automatic cleanup maintains repository cleanliness
- [ ] Concurrent tasks work independently
- [ ] Error handling covers all Git operation failures
- [ ] Database tracking enables audit and monitoring
- [ ] Performance impact is minimal
- [ ] Integration tests pass with real repositories

## Dependencies
- Database Schema Update (02) - for branch tracking
- Git CLI available in execution environment
- Repository access permissions
- Database connection for tracking

## Notes
Branch management is critical for task isolation and preventing conflicts. Focus on robust error handling and ensure that failed operations don't leave repositories in inconsistent states. The branch registry should be thread-safe to support concurrent task execution.