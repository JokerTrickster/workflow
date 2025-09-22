# Backend and Frontend Integration Test Plan

## Overview
This document outlines the integration between the backend API and frontend for Claude task execution.

## Backend API Endpoint
- **URL**: `POST /api/v1/claude/run-tasks`
- **Request Structure**:
```json
{
  "tasks": "string (required)",
  "repository_name": "string (required)", 
  "working_dir": "string (optional)",
  "interactive": "boolean (optional)",
  "claude_cmd": "string (optional)",
  "continue_task": "boolean (optional)"
}
```
- **Response Structure**:
```json
{
  "request_id": "string",
  "status": "pending",
  "message": "string",
  "created_at": "ISO timestamp"
}
```

## Frontend Integration

### 1. Service Layer
- `claudeService.ts` - API client wrapper
- `useClaude.ts` - React hook for state management
- `ClaudeTaskRunner.tsx` - UI component example

### 2. Usage Example
```typescript
import { useClaude } from '../hooks/useClaude';

function MyComponent() {
  const claude = useClaude();
  
  const handleRunTasks = async () => {
    try {
      const result = await claude.runTasksAndWait({
        tasks: "Implement user authentication",
        repository_name: "my-project",
        working_dir: "/path/to/project",
        interactive: false
      });
      
      console.log('Task completed:', result);
    } catch (error) {
      console.error('Task failed:', error);
    }
  };
  
  return (
    <button onClick={handleRunTasks} disabled={claude.isLoading}>
      {claude.isLoading ? 'Processing...' : 'Run Tasks'}
    </button>
  );
}
```

## Testing Steps

### 1. Backend Test
```bash
# Start the backend server
cd /Users/luxrobo/project/workflow/backend
go run main.go

# Test the endpoint with curl
curl -X POST http://localhost:8080/api/v1/claude/run-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": "Test task",
    "repository_name": "test-repo"
  }'
```

### 2. Frontend Test
```bash
# Start the frontend development server
cd /Users/luxrobo/project/workflow/frontend
npm run dev

# Navigate to a page that uses the ClaudeTaskRunner component
# Fill out the form and submit a test task
```

### 3. Integration Test
1. Start both backend and frontend servers
2. Use the frontend UI to submit a Claude task
3. Verify the request reaches the backend
4. Check that the backend logs the task submission
5. Verify the frontend receives the proper response

## Local Backend Integration

The backend is designed to match the `ReqRunTasksClaude` structure used by local-backend:

```go
type ReqRunTasksClaude struct {
    Tasks          string `json:"tasks" validate:"required"`
    RepositoryName string `json:"repository_name" validate:"required"`
    WorkingDir     string `json:"working_dir,omitempty"`
    Interactive    bool   `json:"interactive,omitempty"`
    ClaudeCmd      string `json:"claude_cmd,omitempty"`
    ContinueTask   bool   `json:"continue_task,omitempty"`
}
```

This ensures compatibility when the queue system forwards messages to local-backend for processing.

## Queue Integration (Future Enhancement)

Currently, the backend API returns a mock response. For full integration:

1. The backend should publish messages to RabbitMQ queue
2. Local-backend should consume messages from the queue
3. Results should be stored and made available via status endpoints
4. Frontend should poll for task completion status

## Configuration

### Backend Environment Variables
```env
PORT=8080
GIN_MODE=release
```

### Frontend Environment Variables
```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```

## Security Considerations

1. **Authentication**: The API currently has placeholder auth middleware
2. **Validation**: Input validation is implemented on both frontend and backend
3. **CORS**: Configured to allow frontend requests during development

## Error Handling

- Frontend provides user-friendly error messages
- Backend returns structured error responses
- Network timeouts and offline scenarios are handled
- Retry logic for failed requests

## Performance

- Frontend uses React hooks for efficient state management
- Backend handles requests asynchronously
- Queue system (when implemented) provides scalability

## Success Criteria

✅ Backend API endpoint accepts requests in ReqRunTasksClaude format
✅ Frontend service can successfully communicate with backend
✅ Request/response structures match between frontend and backend
✅ Error handling works for invalid requests
✅ Code compiles and builds without errors

## Next Steps

1. Implement full queue integration with local-backend
2. Add task status polling endpoints
3. Implement proper authentication
4. Add comprehensive error recovery
5. Set up monitoring and logging