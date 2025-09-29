---
name: task-pipeline-enhancement
epic: task-repository-integration
type: core
priority: critical
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 4
complexity: high
dependencies: [github-service-extension, database-schema-update]
---

# Task: Task Pipeline Enhancement

## Overview
Modify the existing task execution pipeline in `runTasksClaudeUseCase.go` to integrate Git workflow automation. This includes GitHub issue creation before task execution, branch management during execution, and PR creation after successful completion, while maintaining full backward compatibility with existing functionality.

## Acceptance Criteria

### Pre-Execution Phase
- [ ] **GitHub Issue Creation**: Automatically create GitHub issue before task execution
- [ ] **Issue Link Storage**: Store GitHub issue URL in workflow_histories table
- [ ] **Graceful Degradation**: Continue task execution even if GitHub issue creation fails
- [ ] **Issue Templates**: Use standardized templates for consistent issue format

### Execution Phase Enhancement
- [ ] **Branch Creation**: Create feature branch before task execution
- [ ] **Branch Tracking**: Store branch name in database for cleanup tracking
- [ ] **Isolation**: Ensure task executes in isolated branch environment
- [ ] **Existing Functionality**: All current task execution features continue working unchanged

### Post-Execution Phase
- [ ] **Automated Commit**: Commit changes with descriptive message including issue reference
- [ ] **PR Creation**: Create pull request automatically after successful task completion
- [ ] **PR Linking**: Link PR to original GitHub issue for full traceability
- [ ] **Status Updates**: Update workflow_histories with PR URL and completion status

### Error Handling & Cleanup
- [ ] **Failure Cleanup**: Clean up branches and issues for failed tasks
- [ ] **Rollback Capability**: Ability to revert changes if Git operations fail
- [ ] **Audit Logging**: Comprehensive logging of all Git operations
- [ ] **Status Tracking**: Track cleanup status for monitoring and debugging

## Implementation Details

### Core Workflow Integration
Enhance `RunTasks` method in `runTasksClaudeUseCase.go`:

```go
func (d *RunTasksClaudeUseCase) RunTasks(c context.Context, req *request.ReqRunTasksClaude) error {
    // 1. Pre-execution: Create GitHub issue
    issueURL, err := d.createGitHubIssue(ctx, req)
    if err != nil {
        log.Printf("Failed to create GitHub issue: %v", err)
        // Continue without issue - graceful degradation
    }

    // 2. Create and switch to feature branch
    branchName, err := d.createFeatureBranch(ctx, req)
    if err != nil {
        return d.handleBranchCreationFailure(err, issueURL)
    }

    // 3. Execute task (existing logic enhanced)
    err = d.executeTaskWithBranchTracking(ctx, req, branchName)

    // 4. Post-execution: Handle results
    if err != nil {
        return d.handleTaskFailure(ctx, req, branchName, issueURL, err)
    }

    // 5. Create PR and cleanup
    return d.handleTaskSuccess(ctx, req, branchName, issueURL)
}
```

### GitHub Issue Integration
```go
func (d *RunTasksClaudeUseCase) createGitHubIssue(ctx context.Context, req *request.ReqRunTasksClaude) (string, error) {
    // Extract repository owner/name from req.RepositoryName
    owner := "JokerTrickster" // From configuration or request
    repo := req.RepositoryName

    // Generate issue title and body from task
    title := fmt.Sprintf("Task: %s", d.generateIssueTitle(req.Tasks))
    body := d.generateIssueBody(req)
    labels := []string{"automated-task", "local-backend"}

    // Create issue via GitHub service
    issue, err := d.GitHubService.CreateIssue(owner, repo, title, body, labels)
    if err != nil {
        return "", err
    }

    return issue.HTMLURL, nil
}
```

### Branch Management
```go
func (d *RunTasksClaudeUseCase) createFeatureBranch(ctx context.Context, req *request.ReqRunTasksClaude) (string, error) {
    repoPath := utils.GetRepositoryPath(req.RepositoryName)

    // Generate unique branch name
    timestamp := time.Now().Format("20060102-150405")
    taskHash := d.generateTaskHash(req.Tasks)
    branchName := fmt.Sprintf("task-%s-%s", timestamp, taskHash)

    // Create and switch to branch
    if err := d.gitCreateBranch(repoPath, branchName); err != nil {
        return "", fmt.Errorf("failed to create branch %s: %w", branchName, err)
    }

    // Update database with branch info
    d.updateWorkflowBranch(req.RequestID, branchName)

    return branchName, nil
}

func (d *RunTasksClaudeUseCase) gitCreateBranch(repoPath, branchName string) error {
    // Ensure we're on default branch
    if err := d.gitCheckoutDefault(repoPath); err != nil {
        return err
    }

    // Pull latest changes
    if err := d.gitPull(repoPath); err != nil {
        log.Printf("Warning: failed to pull latest changes: %v", err)
    }

    // Create and checkout new branch
    cmd := exec.Command("git", "checkout", "-b", branchName)
    cmd.Dir = repoPath
    return cmd.Run()
}
```

### PR Creation and Linking
```go
func (d *RunTasksClaudeUseCase) createPullRequest(ctx context.Context, req *request.ReqRunTasksClaude, branchName, issueURL string) (string, error) {
    owner := "JokerTrickster"
    repo := req.RepositoryName

    // Generate PR title and body
    title := fmt.Sprintf("Task Implementation: %s", d.generatePRTitle(req.Tasks))
    body := d.generatePRBody(req, branchName, issueURL)

    // Get default branch
    defaultBranch := d.getDefaultBranch(utils.GetRepositoryPath(req.RepositoryName))

    // Create PR
    pr, err := d.GitHubService.CreatePullRequest(owner, repo, title, body, branchName, defaultBranch)
    if err != nil {
        return "", err
    }

    return pr.HTMLURL, nil
}
```

### Enhanced Error Handling
```go
func (d *RunTasksClaudeUseCase) handleTaskFailure(ctx context.Context, req *request.ReqRunTasksClaude, branchName, issueURL string, taskErr error) error {
    // Log the failure
    log.Printf("Task execution failed: %v", taskErr)

    // Cleanup branch
    if branchName != "" {
        if err := d.cleanupBranch(req.RepositoryName, branchName); err != nil {
            log.Printf("Failed to cleanup branch %s: %v", branchName, err)
        }
    }

    // Update issue status (optional)
    if issueURL != "" {
        d.updateIssueWithFailure(issueURL, taskErr)
    }

    // Update database
    d.updateWorkflowStatus(req.RequestID, "failed", taskErr.Error())

    return taskErr
}

func (d *RunTasksClaudeUseCase) handleTaskSuccess(ctx context.Context, req *request.ReqRunTasksClaude, branchName, issueURL string) error {
    repoPath := utils.GetRepositoryPath(req.RepositoryName)

    // Commit changes
    if err := d.commitChanges(repoPath, req.Tasks, issueURL); err != nil {
        return fmt.Errorf("failed to commit changes: %w", err)
    }

    // Push branch
    if err := d.pushBranch(repoPath, branchName); err != nil {
        return fmt.Errorf("failed to push branch: %w", err)
    }

    // Create PR
    prURL, err := d.createPullRequest(ctx, req, branchName, issueURL)
    if err != nil {
        log.Printf("Failed to create PR: %v", err)
        // Don't fail the task for PR creation failure
    }

    // Update database
    d.updateWorkflowCompletion(req.RequestID, issueURL, prURL, branchName)

    return nil
}
```

### Template Generation
```go
func (d *RunTasksClaudeUseCase) generateIssueBody(req *request.ReqRunTasksClaude) string {
    return fmt.Sprintf(`# Task: %s

**Repository**: %s
**Executed by**: Local Backend Task System
**Created**: %s

## Task Details
%s

## Execution Context
- **Branch**: Will be created automatically
- **Working Directory**: %s
- **Task ID**: %s

---
*This issue was automatically created by the task execution system.*`,
        req.Tasks,
        req.RepositoryName,
        time.Now().Format("2006-01-02 15:04:05"),
        req.Tasks,
        req.WorkingDir,
        // Extract request ID from context or generate
    )
}

func (d *RunTasksClaudeUseCase) generatePRBody(req *request.ReqRunTasksClaude, branchName, issueURL string) string {
    // Extract issue number from URL for closing syntax
    issueNumber := d.extractIssueNumber(issueURL)

    return fmt.Sprintf(`# Task Implementation: %s

**Closes**: #%s

## Changes Made
%s

## Task Context
- **Original Task**: %s
- **Repository**: %s
- **Branch**: %s
- **Execution Time**: %s

## Review Checklist
- [ ] Code follows project conventions
- [ ] Changes are well-tested
- [ ] Documentation updated if needed
- [ ] No breaking changes introduced

---
*This PR was automatically created by the task execution system.*`,
        req.Tasks,
        issueNumber,
        "Automated implementation via Claude Code task execution",
        req.Tasks,
        req.RepositoryName,
        branchName,
        time.Now().Format("2006-01-02 15:04:05"),
    )
}
```

## Testing Requirements

### Unit Tests
- [ ] Issue creation with various task types
- [ ] Branch name generation and uniqueness
- [ ] PR creation and linking logic
- [ ] Error handling for each workflow step
- [ ] Template generation with different inputs

### Integration Tests
- [ ] End-to-end workflow with real repositories
- [ ] GitHub API integration testing
- [ ] Database updates during workflow
- [ ] Cleanup operations for failed tasks
- [ ] Backward compatibility with existing functionality

### Performance Tests
- [ ] Task execution time impact measurement
- [ ] Concurrent task handling with Git operations
- [ ] Database query performance with new fields
- [ ] GitHub API rate limiting behavior

## Definition of Done
- [ ] All workflow phases implemented and tested
- [ ] GitHub integration working end-to-end
- [ ] Branch management prevents task conflicts
- [ ] Error handling covers all failure scenarios
- [ ] Database tracking enables full audit trail
- [ ] Existing functionality continues unchanged
- [ ] Performance impact < 10% of baseline
- [ ] Code review completed and approved

## Dependencies
- GitHub Service Extension (01) - for issue/PR creation
- Database Schema Update (02) - for workflow tracking
- Test repositories with GitHub integration
- GitHub API tokens with appropriate permissions

## Notes
This is the core integration task that brings together all Git workflow components. Focus on robust error handling and graceful degradation to ensure task execution continues even when Git operations fail. Maintain full backward compatibility with existing task execution patterns.