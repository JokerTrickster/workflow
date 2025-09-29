# API Documentation

## Overview

The Task Repository Integration System provides RESTful APIs for task management, GitHub integration, and workflow automation.

## Base URLs

- **Backend API**: `http://localhost:7000/api/v1`
- **Local-backend**: `http://localhost:8080` (internal)

## Authentication

Most endpoints require authentication via session cookies or API keys.

```http
Cookie: session_token=your_session_token
```

## Task Management API

### Create Task

Creates a new task for execution with automatic GitHub integration.

**Endpoint**: `POST /api/v1/tasks`

**Request Body**:
```json
{
  "tasks": "Add unit tests for authentication module",
  "repository_name": "JokerTrickster/my-app",
  "provider": "claude",
  "working_dir": "/optional/custom/path",
  "interactive": false,
  "cmd": "npm test",
  "continue_task": false
}
```

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "request_id": "req_1234567890abcdef",
  "status": "pending",
  "tasks": "Add unit tests for authentication module",
  "repository_name": "JokerTrickster/my-app",
  "provider": "claude",
  "interactive": false,
  "continue_task": false,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Execute Task

Executes a pending task.

**Endpoint**: `POST /api/v1/tasks/{task_id}/execute`

**Response**:
```json
{
  "message": "Task execution started",
  "request_id": "req_1234567890abcdef",
  "status": "processing"
}
```

### Get Task Status

Retrieves the current status of a task.

**Endpoint**: `GET /api/v1/tasks/{task_id}/status`

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "progress": 100,
  "github_issue_url": "https://github.com/JokerTrickster/my-app/issues/123",
  "github_pr_url": "https://github.com/JokerTrickster/my-app/pull/456",
  "branch_name": "task/add-unit-tests-0115-1030-a1b2c3d4",
  "completed_at": "2024-01-15T10:45:00Z",
  "processing_time_ms": 900000,
  "result": {
    "success": true,
    "files_modified": ["auth_test.go", "user_test.go"],
    "output": "Tests added successfully"
  }
}
```

### Get Task History

Retrieves paginated task history for a repository.

**Endpoint**: `GET /api/v1/workflow-histories`

**Query Parameters**:
- `repository_name` (required): Repository name (e.g., "JokerTrickster/my-app")
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20, max: 100)
- `status` (optional): Filter by status (pending, processing, completed, failed)
- `provider` (optional): Filter by AI provider (claude, codex, cursor)

**Response**:
```json
{
  "tasks": [
    {
      "id": 1,
      "request_id": "req_1234567890abcdef",
      "status": "completed",
      "tasks": "Add unit tests for authentication module",
      "repository_name": "JokerTrickster/my-app",
      "provider": "claude",
      "github_issue_url": "https://github.com/JokerTrickster/my-app/issues/123",
      "github_pr_url": "https://github.com/JokerTrickster/my-app/pull/456",
      "branch_name": "task/add-unit-tests-0115-1030-a1b2c3d4",
      "cleanup_status": "completed",
      "created_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:45:00Z",
      "processing_time_ms": 900000
    }
  ],
  "total_count": 45,
  "page": 1,
  "limit": 20,
  "has_more": true
}
```

### Update Task

Updates task properties.

**Endpoint**: `PUT /api/v1/tasks/{task_id}`

**Request Body**:
```json
{
  "status": "cancelled",
  "cleanup_status": "pending"
}
```

### Delete Task

Deletes a task (soft delete).

**Endpoint**: `DELETE /api/v1/tasks/{task_id}`

**Response**: `204 No Content`

## GitHub Integration API

### Get GitHub User

Retrieves authenticated GitHub user information.

**Endpoint**: `GET /api/v1/github/user`

**Headers**:
```http
Authorization: Bearer github_token_here
```

**Response**:
```json
{
  "login": "username",
  "id": 12345,
  "name": "Full Name",
  "email": "user@example.com",
  "avatar_url": "https://avatars.githubusercontent.com/u/12345?v=4"
}
```

### Get User Repositories

Retrieves repositories accessible to the authenticated user.

**Endpoint**: `GET /api/v1/github/repositories`

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 30, max: 100)

**Response**:
```json
{
  "repositories": [
    {
      "id": 123456789,
      "name": "my-app",
      "full_name": "JokerTrickster/my-app",
      "private": false,
      "html_url": "https://github.com/JokerTrickster/my-app",
      "description": "My awesome application",
      "default_branch": "main",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total_count": 15,
  "page": 1,
  "per_page": 30
}
```

### Sync Repositories

Synchronizes user repositories with local storage.

**Endpoint**: `POST /api/v1/github/sync-repositories`

**Response**:
```json
{
  "message": "Repository synchronization started",
  "synced_count": 15,
  "status": "completed"
}
```

## Repository Management API

### Get Repository Status

Gets the connection and sync status of a repository.

**Endpoint**: `GET /api/v1/repositories/{repo_id}/status`

**Response**:
```json
{
  "connected": true,
  "local_path": "/Users/mac/project/git-repository/JokerTrickster/my-app",
  "last_sync": "2024-01-15T09:00:00Z",
  "branch": "main",
  "commit_hash": "a1b2c3d4e5f6789",
  "has_uncommitted_changes": false
}
```

### Clone Repository

Clones a repository to local storage.

**Endpoint**: `POST /api/v1/repositories/{repo_id}/clone`

**Response**:
```json
{
  "message": "Repository cloning started",
  "local_path": "/Users/mac/project/git-repository/JokerTrickster/my-app",
  "status": "cloning"
}
```

## Workflow Status API

### Get Active Workflows

Retrieves currently active task workflows.

**Endpoint**: `GET /api/v1/workflows/active`

**Response**:
```json
{
  "workflows": [
    {
      "repository_name": "JokerTrickster/my-app",
      "branch_name": "task/add-unit-tests-0115-1030-a1b2c3d4",
      "task_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "processing",
      "started_at": "2024-01-15T10:30:00Z",
      "github_issue_url": "https://github.com/JokerTrickster/my-app/issues/123"
    }
  ]
}
```

### Get System Health

Retrieves system health status.

**Endpoint**: `GET /api/v1/health`

**Response**:
```json
{
  "status": "healthy",
  "components": {
    "database": {
      "status": "up",
      "response_time_ms": 15
    },
    "rabbitmq": {
      "status": "up",
      "queue_depth": 5
    },
    "github_api": {
      "status": "up",
      "rate_limit_remaining": 4500
    },
    "local_backend": {
      "status": "up",
      "active_tasks": 2
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request body is invalid",
    "details": "Field 'repository_name' is required",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_error_123456"
  }
}
```

### Error Codes

- `INVALID_REQUEST` (400): Request validation failed
- `UNAUTHORIZED` (401): Authentication required
- `FORBIDDEN` (403): Insufficient permissions
- `NOT_FOUND` (404): Resource not found
- `CONFLICT` (409): Resource conflict
- `RATE_LIMIT_EXCEEDED` (429): Too many requests
- `INTERNAL_ERROR` (500): Server error
- `SERVICE_UNAVAILABLE` (503): Service temporarily unavailable

## Rate Limiting

API endpoints are rate limited to ensure fair usage:

- **General endpoints**: 100 requests per minute
- **Task execution**: 10 requests per minute
- **GitHub integration**: 1000 requests per hour (shared with GitHub API limits)

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## Webhooks

### GitHub Webhook Events

The system can receive GitHub webhook events for automated task management.

**Endpoint**: `POST /api/v1/webhooks/github`

**Supported Events**:
- `issues.opened`: Automatically create task from issue
- `pull_request.closed`: Cleanup associated branch
- `push`: Trigger automated tasks based on changes

**Webhook Payload Example**:
```json
{
  "action": "opened",
  "issue": {
    "number": 123,
    "title": "Add authentication tests",
    "body": "We need comprehensive tests for the auth module",
    "html_url": "https://github.com/JokerTrickster/my-app/issues/123"
  },
  "repository": {
    "full_name": "JokerTrickster/my-app"
  }
}
```

## SDK Examples

### JavaScript/Node.js

```javascript
const WorkflowAPI = require('./workflow-sdk');

const client = new WorkflowAPI({
  baseURL: 'http://localhost:7000/api/v1',
  apiKey: 'your-api-key'
});

// Create and execute a task
async function createTask() {
  const task = await client.tasks.create({
    tasks: 'Add unit tests for authentication module',
    repository_name: 'JokerTrickster/my-app',
    provider: 'claude'
  });

  await client.tasks.execute(task.id);

  // Poll for completion
  const status = await client.tasks.waitForCompletion(task.id);
  console.log('GitHub PR:', status.github_pr_url);
}
```

### Go

```go
package main

import (
    "context"
    "github.com/JokerTrickster/workflow/client"
)

func main() {
    client := workflow.NewClient(&workflow.Config{
        BaseURL: "http://localhost:7000/api/v1",
        APIKey:  "your-api-key",
    })

    task, err := client.Tasks.Create(context.Background(), &workflow.CreateTaskRequest{
        Tasks:          "Add unit tests for authentication module",
        RepositoryName: "JokerTrickster/my-app",
        Provider:       "claude",
    })

    if err != nil {
        panic(err)
    }

    err = client.Tasks.Execute(context.Background(), task.ID)
    if err != nil {
        panic(err)
    }
}
```

### Python

```python
from workflow_client import WorkflowClient

client = WorkflowClient(
    base_url='http://localhost:7000/api/v1',
    api_key='your-api-key'
)

# Create and execute task
task = client.tasks.create(
    tasks='Add unit tests for authentication module',
    repository_name='JokerTrickster/my-app',
    provider='claude'
)

client.tasks.execute(task.id)

# Wait for completion
status = client.tasks.wait_for_completion(task.id, timeout=1800)
print(f"GitHub PR: {status.github_pr_url}")
```

## Postman Collection

A complete Postman collection is available at `/docs/postman/Workflow_API.postman_collection.json` with pre-configured requests and examples.