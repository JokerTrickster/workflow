---
name: pr-automation
epic: task-repository-integration
type: core
priority: high
status: backlog
created: 2025-09-29T02:57:03Z
estimated_days: 3
complexity: medium
dependencies: [github-service-extension, task-pipeline-enhancement, branch-management]
---

# Task: PR Automation

## Overview
Implement automated pull request creation after successful task completion, including proper linking to GitHub issues, reviewer assignment, and standardized PR templates. This completes the Git workflow automation by ensuring all code changes go through proper review process.

## Acceptance Criteria

### PR Creation Automation
- [ ] **Automatic Trigger**: Create PR immediately after successful task completion and branch push
- [ ] **Issue Linking**: Link PR to originating GitHub issue using "Closes #N" syntax
- [ ] **Standardized Templates**: Use consistent PR templates with task context and review checklist
- [ ] **Reviewer Assignment**: Automatically assign appropriate reviewers based on repository configuration

### PR Content & Metadata
- [ ] **Descriptive Titles**: Generate meaningful PR titles based on task description
- [ ] **Comprehensive Body**: Include task details, changes summary, and review checklist
- [ ] **Labels & Tags**: Apply appropriate labels (automated-task, review-required, etc.)
- [ ] **Branch References**: Clearly identify source and target branches

### Integration & Tracking
- [ ] **Database Storage**: Store PR URL in workflow_histories table for tracking
- [ ] **Status Monitoring**: Track PR status (open, merged, closed) for cleanup operations
- [ ] **Error Handling**: Graceful failure when PR creation fails (task still succeeds)
- [ ] **Audit Trail**: Log all PR operations for monitoring and debugging

### Cleanup Integration
- [ ] **Merge Detection**: Detect when PR is merged to trigger branch cleanup
- [ ] **Auto-close Issues**: Close associated GitHub issues when PR is merged
- [ ] **Branch Deletion**: Remove feature branches after successful merge
- [ ] **Failure Cleanup**: Handle cleanup for rejected or closed PRs

## Implementation Details

### PR Creation Service
```go
type PRAutomationService struct {
    GitHubService  *github.Service
    BranchManager  *BranchManager
    TemplateEngine *template.Template
    DB            *sql.DB
}

func (pr *PRAutomationService) CreatePullRequest(ctx context.Context, req *CreatePRRequest) (*PRResult, error) {
    // 1. Validate prerequisites
    if err := pr.validatePRPrerequisites(req); err != nil {
        return nil, fmt.Errorf("PR prerequisites not met: %w", err)
    }

    // 2. Generate PR content
    prContent := pr.generatePRContent(req)

    // 3. Create PR via GitHub API
    pullRequest, err := pr.GitHubService.CreatePullRequest(
        req.Owner,
        req.Repository,
        prContent.Title,
        prContent.Body,
        req.SourceBranch,
        req.TargetBranch,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create PR: %w", err)
    }

    // 4. Apply labels and assign reviewers
    if err := pr.enhancePullRequest(pullRequest, req); err != nil {
        log.Printf("Warning: failed to enhance PR: %v", err)
        // Don't fail for enhancement errors
    }

    // 5. Update database tracking
    if err := pr.updatePRTracking(req.RequestID, pullRequest); err != nil {
        log.Printf("Warning: failed to update PR tracking: %v", err)
    }

    return &PRResult{
        URL:    pullRequest.HTMLURL,
        Number: pullRequest.Number,
        ID:     pullRequest.ID,
    }, nil
}

type CreatePRRequest struct {
    RequestID      string
    Owner          string
    Repository     string
    SourceBranch   string
    TargetBranch   string
    TaskDescription string
    IssueURL       string
    IssueNumber    int
    Changes        []ChangeInfo
    ExecutionTime  time.Duration
}

type PRContent struct {
    Title string
    Body  string
}
```

### PR Template Generation
```go
func (pr *PRAutomationService) generatePRContent(req *CreatePRRequest) PRContent {
    title := pr.generatePRTitle(req.TaskDescription)
    body := pr.generatePRBody(req)

    return PRContent{
        Title: title,
        Body:  body,
    }
}

func (pr *PRAutomationService) generatePRTitle(taskDescription string) string {
    // Clean and format task description for PR title
    title := strings.TrimSpace(taskDescription)

    // Limit length and add prefix
    if len(title) > 72 {
        title = title[:69] + "..."
    }

    return fmt.Sprintf("feat: %s", title)
}

func (pr *PRAutomationService) generatePRBody(req *CreatePRRequest) string {
    tmpl := `# {{ .TaskDescription }}

{{ if .IssueNumber }}**Closes**: #{{ .IssueNumber }}{{ end }}

## Changes Made
This pull request implements the requested task through automated execution.

{{ if .Changes }}
### Specific Changes:
{{- range .Changes }}
- **{{ .Type }}**: {{ .Description }}
{{- end }}
{{ end }}

## Task Context
- **Original Task**: {{ .TaskDescription }}
- **Repository**: {{ .Repository }}
- **Source Branch**: {{ .SourceBranch }}
- **Target Branch**: {{ .TargetBranch }}
- **Execution Time**: {{ .ExecutionTime }}
- **Request ID**: {{ .RequestID }}

## Review Checklist
- [ ] Code follows project conventions and style guidelines
- [ ] Changes are well-tested and don't break existing functionality
- [ ] Documentation has been updated if necessary
- [ ] No sensitive information has been committed
- [ ] Branch is up to date with target branch
- [ ] All CI checks are passing

## Automated Task Information
This pull request was automatically created by the local-backend task execution system. The changes were generated through Claude Code task processing and represent the implementation of the specified task.

**Review Guidelines:**
- Verify the changes align with the task requirements
- Check for any unintended modifications
- Ensure code quality meets project standards
- Test the changes in your local environment if needed

---
*🤖 This PR was automatically generated by the task execution system*
*📋 Review the changes carefully before merging*`

    t := template.Must(template.New("pr").Parse(tmpl))
    var buf bytes.Buffer

    data := struct {
        *CreatePRRequest
        ExecutionTime string
    }{
        CreatePRRequest: req,
        ExecutionTime:   req.ExecutionTime.String(),
    }

    t.Execute(&buf, data)
    return buf.String()
}
```

### Reviewer Assignment Logic
```go
type ReviewerConfig struct {
    Repository    string              `json:"repository"`
    DefaultTeam   []string           `json:"default_team"`
    PathReviewers map[string][]string `json:"path_reviewers"`
    RequiredCount int                `json:"required_count"`
}

func (pr *PRAutomationService) assignReviewers(pullRequest *github.PullRequest, req *CreatePRRequest) error {
    config, err := pr.loadReviewerConfig(req.Repository)
    if err != nil {
        log.Printf("Warning: could not load reviewer config: %v", err)
        return nil // Don't fail PR creation for reviewer assignment
    }

    reviewers := pr.selectReviewers(config, req.Changes)

    if len(reviewers) == 0 {
        log.Printf("No reviewers configured for repository %s", req.Repository)
        return nil
    }

    return pr.GitHubService.AssignReviewers(
        req.Owner,
        req.Repository,
        pullRequest.Number,
        reviewers,
    )
}

func (pr *PRAutomationService) selectReviewers(config *ReviewerConfig, changes []ChangeInfo) []string {
    reviewerSet := make(map[string]bool)

    // Add path-specific reviewers based on changes
    for _, change := range changes {
        for pathPattern, reviewers := range config.PathReviewers {
            if matched, _ := filepath.Match(pathPattern, change.Path); matched {
                for _, reviewer := range reviewers {
                    reviewerSet[reviewer] = true
                }
            }
        }
    }

    // Add default team members if no specific reviewers found
    if len(reviewerSet) == 0 {
        for _, reviewer := range config.DefaultTeam {
            reviewerSet[reviewer] = true
        }
    }

    // Convert to slice and limit count
    reviewers := make([]string, 0, len(reviewerSet))
    for reviewer := range reviewerSet {
        reviewers = append(reviewers, reviewer)
        if len(reviewers) >= config.RequiredCount {
            break
        }
    }

    return reviewers
}
```

### PR Status Monitoring
```go
func (pr *PRAutomationService) MonitorPRStatus(ctx context.Context) error {
    // Get all PRs that need status monitoring
    prs, err := pr.getActivePRs()
    if err != nil {
        return err
    }

    for _, prInfo := range prs {
        status, err := pr.GitHubService.GetPRStatus(prInfo.Owner, prInfo.Repository, prInfo.Number)
        if err != nil {
            log.Printf("Failed to get PR status for %s: %v", prInfo.URL, err)
            continue
        }

        switch status.State {
        case "merged":
            if err := pr.handlePRMerged(prInfo); err != nil {
                log.Printf("Failed to handle merged PR %s: %v", prInfo.URL, err)
            }
        case "closed":
            if err := pr.handlePRClosed(prInfo); err != nil {
                log.Printf("Failed to handle closed PR %s: %v", prInfo.URL, err)
            }
        }
    }

    return nil
}

func (pr *PRAutomationService) handlePRMerged(prInfo *PRInfo) error {
    // 1. Clean up feature branch
    if err := pr.BranchManager.CleanupBranch(prInfo.BranchName); err != nil {
        log.Printf("Failed to cleanup branch %s: %v", prInfo.BranchName, err)
    }

    // 2. Close associated issue
    if prInfo.IssueNumber > 0 {
        if err := pr.GitHubService.CloseIssue(prInfo.Owner, prInfo.Repository, prInfo.IssueNumber); err != nil {
            log.Printf("Failed to close issue #%d: %v", prInfo.IssueNumber, err)
        }
    }

    // 3. Update database
    return pr.updateCleanupStatus(prInfo.RequestID, "completed")
}

func (pr *PRAutomationService) handlePRClosed(prInfo *PRInfo) error {
    // Clean up branch for rejected/closed PRs
    if err := pr.BranchManager.CleanupBranch(prInfo.BranchName); err != nil {
        log.Printf("Failed to cleanup branch %s: %v", prInfo.BranchName, err)
    }

    // Update status
    return pr.updateCleanupStatus(prInfo.RequestID, "completed")
}
```

### Integration with Task Pipeline
```go
// In runTasksClaudeUseCase.go
func (d *RunTasksClaudeUseCase) handleTaskSuccess(ctx context.Context, req *request.ReqRunTasksClaude, branchName, issueURL string) error {
    // ... existing commit and push logic ...

    // Create PR automatically
    prRequest := &CreatePRRequest{
        RequestID:       req.RequestID,
        Owner:          "JokerTrickster",
        Repository:     req.RepositoryName,
        SourceBranch:   branchName,
        TargetBranch:   d.getDefaultBranch(utils.GetRepositoryPath(req.RepositoryName)),
        TaskDescription: req.Tasks,
        IssueURL:       issueURL,
        IssueNumber:    d.extractIssueNumber(issueURL),
        Changes:        d.analyzeChanges(utils.GetRepositoryPath(req.RepositoryName)),
        ExecutionTime:  time.Since(startTime),
    }

    prResult, err := d.PRAutomationService.CreatePullRequest(ctx, prRequest)
    if err != nil {
        log.Printf("Failed to create PR: %v", err)
        // Don't fail the task for PR creation failure
        prResult = &PRResult{URL: ""}
    }

    // Update database with all URLs
    return d.updateWorkflowCompletion(req.RequestID, issueURL, prResult.URL, branchName)
}
```

## Testing Requirements

### Unit Tests
- [ ] PR content generation with various task types
- [ ] Reviewer assignment logic for different scenarios
- [ ] Template rendering with edge cases
- [ ] Error handling for GitHub API failures

### Integration Tests
- [ ] End-to-end PR creation flow
- [ ] Issue linking and closure workflow
- [ ] Reviewer assignment in real repositories
- [ ] Branch cleanup after PR merge

### Performance Tests
- [ ] PR creation time under load
- [ ] Status monitoring performance
- [ ] Template rendering performance
- [ ] Database update efficiency

## Definition of Done
- [ ] PRs automatically created after successful tasks
- [ ] Issues properly linked and closed
- [ ] Reviewers assigned based on configuration
- [ ] Templates provide comprehensive context
- [ ] Status monitoring enables automatic cleanup
- [ ] Error handling maintains system stability
- [ ] Integration tests pass with real GitHub
- [ ] Performance meets requirements

## Dependencies
- GitHub Service Extension (01) - for PR creation API
- Task Pipeline Enhancement (03) - for integration point
- Branch Management (04) - for branch cleanup
- Test repositories with review permissions

## Notes
PR automation completes the Git workflow by ensuring all changes go through proper review. Focus on robust error handling and ensure that PR creation failures don't prevent task completion. The reviewer assignment should be configurable per repository to accommodate different team structures.