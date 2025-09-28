# CLAUDE.md

> Think carefully and implement the most concise solution that changes as little code as possible.

## Project Information
- **Repository**: JokerTrickster/workflow
- **GitHub URL**: https://github.com/JokerTrickster/workflow
- **Project**: workflow

## CCPM Integration (Claude Code PM)
This project uses CCPM for structured project management with AI agents. 

### Quick Commands:
- `/pm:prd-new feature-name` - Create new Product Requirements Document
- `/pm:prd-parse feature-name` - Convert PRD to implementation plan
- `/pm:epic-oneshot feature-name` - Decompose and sync to GitHub issues
- `/pm:issue-start 1234` - Start working on GitHub issue with specialized agent
- `/pm:next` - Get next priority task with context
- `/pm:status` - Overall project dashboard

## USE SUB-AGENTS FOR CONTEXT OPTIMIZATION

### 1. Always use the file-analyzer sub-agent when asked to read files.
The file-analyzer agent is an expert in extracting and summarizing critical information from files, particularly log files and verbose outputs. It provides concise, actionable summaries that preserve essential information while dramatically reducing context usage.

### 2. Always use the code-analyzer sub-agent when asked to search code, analyze code, research bugs, or trace logic flow.

The code-analyzer agent is an expert in code analysis, logic tracing, and vulnerability detection. It provides concise, actionable summaries that preserve essential information while dramatically reducing context usage.

### 3. Always use the test-runner sub-agent to run tests and analyze the test results.

Using the test-runner agent ensures:

- Full test output is captured for debugging
- Main conversation stays clean and focused
- Context usage is optimized
- All issues are properly surfaced
- No approval dialogs interrupt the workflow

## Philosophy

### Error Handling

- **Fail fast** for critical configuration (missing text model)
- **Log and continue** for optional features (extraction model)
- **Graceful degradation** when external services unavailable
- **User-friendly messages** through resilience layer

### Testing

- Always use the test-runner agent to execute tests.
- Do not use mock services for anything ever.
- Do not move on to the next test until the current test is complete.
- If the test fails, consider checking if the test is structured correctly before deciding we need to refactor the codebase.
- Tests to be verbose so we can use them for debugging.

## Tone and Behavior

- Criticism is welcome. Please tell me when I am wrong or mistaken, or even when you think I might be wrong or mistaken.
- Please tell me if there is a better approach than the one I am taking.
- Please tell me if there is a relevant standard or convention that I appear to be unaware of.
- Be skeptical.
- Be concise.
- Short summaries are OK, but don't give an extended breakdown unless we are working through the details of a plan.
- Do not flatter, and do not give compliments unless I am specifically asking for your judgement.
- Occasional pleasantries are fine.
- Feel free to ask many questions. If you are in doubt of my intent, don't guess. Ask.

## ABSOLUTE RULES:

- NO PARTIAL IMPLEMENTATION
- NO SIMPLIFICATION : no "//This is simplified stuff for now, complete implementation would blablabla"
- NO CODE DUPLICATION : check existing codebase to reuse functions and constants Read files before writing new functions. Use common sense function name to find them easily.
- NO DEAD CODE : either use or delete from codebase completely
- IMPLEMENT TEST FOR EVERY FUNCTIONS
- NO CHEATER TESTS : test must be accurate, reflect real usage and be designed to reveal flaws. No useless tests! Design tests to be verbose so we can use them for debuging.
- NO INCONSISTENT NAMING - read existing codebase naming patterns.
- NO OVER-ENGINEERING - Don't add unnecessary abstractions, factory patterns, or middleware when simple functions would work. Don't think "enterprise" when you need "working"
- NO MIXED CONCERNS - Don't put validation logic inside API handlers, database queries inside UI components, etc. instead of proper separation
- NO RESOURCE LEAKS - Don't forget to close database connections, clear timeouts, remove event listeners, or clean up file handles
- ALWAYS COMMIT AND PUSH - After completing any implementation work, always commit changes and push to GitHub
- ALWAYS LOG TASKS - Every task execution must be logged to .claude/tasks folder with detailed context, status, and results for continuity and web display

## Workflow History System

### Database Flow
The project implements a workflow history system that tracks Claude task executions from frontend to local-backend:

**1. Frontend → Backend → RabbitMQ**
- User submits task via `ClaudeTaskRunner` component
- Backend receives request and publishes to RabbitMQ queue
- Backend returns `request_id` to frontend for tracking

**2. Local-backend → Database**
- Local-backend consumes RabbitMQ messages
- Creates `workflow_histories` record with status 'pending'
- Updates status to 'processing' when task starts
- Updates status to 'completed'/'failed' with results when task finishes

**3. Database Schema (Simplified)**
```sql
workflow_histories:
- request_id (unique identifier for frontend)
- status (pending → processing → completed/failed)
- tasks, repository_name, working_dir, claude_cmd
- interactive, continue_task (frontend input values)
- created_at, completed_at, processing_time_ms
- result (JSON), error (text)
```

**4. GORM Model**
```go
type WorkflowHistories struct {
    RequestID        string    `json:"request_id"`
    Status           string    `json:"status"`
    Tasks            string    `json:"tasks"`
    RepositoryName   string    `json:"repository_name"`
    // ... (matches frontend display data exactly)
}
```

**5. Implementation Flow**
- RabbitMQ message received → Create DB record (status: pending)
- Task execution starts → Update status to 'processing'
- Task completes → Update completed_at, processing_time_ms, result/error, status

This system provides full traceability of task executions while keeping the database schema simple and focused on frontend display requirements.

## Repository Management and Task Execution

### Local Repository Structure
All repositories are managed under the following directory structure:
```
/Users/mac/project/git-repository/JokerTrickster/
├── gallery_ios/          # iOS gallery app repository
├── workflow/              # Main workflow repository
├── board_game_app/        # Board game application
├── eatplay_app/          # Food recommendation app
└── ...                   # Other repositories
```

### Task Execution Workflow

When a task is submitted with a `repository_name` (e.g., "gallery_ios"), the local-backend:

1. **Repository Resolution**:
   - Looks for repository in `/Users/mac/project/git-repository/JokerTrickster/{repository_name}`
   - Validates that the directory exists and is a Git repository

2. **Branch Management**:
   - Creates a new feature branch: `claude-task-{timestamp}`
   - Switches to the new branch for task execution
   - Example: `claude-task-1727692800`

3. **Task Execution**:
   - Executes Claude AI tasks within the repository directory
   - Files are created/modified in the repository context
   - All changes happen in the feature branch

4. **Git Operations**:
   - Automatically stages all changes: `git add .`
   - Creates commit with descriptive message
   - Pushes branch to GitHub: `git push --set-upstream origin {branch}`

5. **Pull Request Creation**:
   - Automatically creates GitHub PR from feature branch to main
   - PR title includes task description
   - PR body includes task details and AI-generated metadata

### Repository Requirements

For a repository to be eligible for task execution:
- Must exist in `/Users/mac/project/git-repository/JokerTrickster/`
- Must be a valid Git repository (contains `.git` directory)
- Must have a configured remote origin pointing to GitHub
- Must be up-to-date with the remote main branch

### Example Task Flow

```bash
# Task: "Add current date to README.md"
# Repository: gallery_ios

1. Repository resolution: /Users/mac/project/git-repository/JokerTrickster/gallery_ios
2. Branch creation: git checkout -b claude-task-1727692800
3. Claude execution: Modifies README.md with current date
4. Git operations:
   - git add .
   - git commit -m "feat: Add current date to README.md"
   - git push --set-upstream origin claude-task-1727692800
5. PR creation: Creates PR #123 with changes
```

This ensures all AI-driven changes are properly versioned, reviewed, and integrated through standard Git workflows.
