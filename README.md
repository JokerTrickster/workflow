# Task Repository Integration System

A comprehensive task execution system with automated Git workflow management and GitHub integration.

## 🚀 Overview

This system transforms the local-backend task processing into a comprehensive Git workflow management platform that automatically creates GitHub issues, manages repository branching, executes tasks in isolated environments, and creates pull requests for review.

## ✨ Features

### Core Functionality
- **Automated GitHub Issue Creation**: Tasks automatically create corresponding GitHub issues for tracking
- **Branch Management**: Isolated feature branches with conflict prevention and cleanup
- **Pull Request Automation**: Automatic PR creation with issue linking and review templates
- **Error Handling & Recovery**: Comprehensive error classification and recovery strategies
- **Frontend Integration**: Enhanced UI displaying GitHub links, branch info, and workflow status

### GitHub Integration
- **Issue Tracking**: Each task creates a GitHub issue with standardized templates
- **Branch Isolation**: Unique branch names with conflict prevention
- **PR Management**: Automatic PR creation with API and CLI fallback
- **Cleanup Automation**: Orphaned branch detection and removal

### Error Management
- **Error Classification**: Intelligent error type detection (GitHub, Git, Provider, Network, etc.)
- **Recovery Strategies**: Automatic retry and fallback mechanisms
- **Comprehensive Logging**: Detailed error tracking and audit trails

## 🏗️ Architecture

### System Components

```
Frontend (Next.js) ←→ Backend (Go) ←→ RabbitMQ ←→ Local-backend (Go)
                                                        ↓
                                             GitHub API Integration
                                                        ↓
                                               Git Repository Management
```

### Key Services

1. **ConcurrentTaskWorker**: Main task processing engine with GitHub integration
2. **BranchManager**: Handles branch lifecycle and conflict prevention
3. **ErrorHandler**: Comprehensive error handling and recovery
4. **PRCreator**: GitHub Pull Request automation with API/CLI fallback
5. **GitHubService**: Direct GitHub API integration

## 📋 Installation & Setup

### Prerequisites
- Go 1.19+
- Node.js 18+
- RabbitMQ
- MySQL/PostgreSQL
- Git CLI
- GitHub CLI (optional, for fallback)

### Environment Variables
```bash
# GitHub Integration
GITHUB_TOKEN=your_github_personal_access_token

# Database
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=workflow
MYSQL_USER=root
MYSQL_PASSWORD=your_password

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_QUEUE_NAME=workflow_queue
```

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/JokerTrickster/workflow.git
   cd workflow
   ```

2. **Setup Backend**
   ```bash
   cd backend
   go mod tidy
   go build
   ./backend
   ```

3. **Setup Local-backend**
   ```bash
   cd local-backend
   go mod tidy
   go build
   ./local-backend
   ```

4. **Setup Frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## 🔧 Configuration

### GitHub Integration Setup

1. **Create GitHub Personal Access Token**
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Create token with `repo`, `issues`, and `pull_requests` scopes
   - Set as `GITHUB_TOKEN` environment variable

2. **Configure Repository Access**
   - Ensure repositories are accessible with the token
   - Set up proper permissions for issue and PR creation

### Database Schema

The system uses the `workflow_histories` table with GitHub integration fields:

```sql
CREATE TABLE workflow_histories (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  request_id VARCHAR(255) UNIQUE NOT NULL,
  status ENUM('pending', 'processing', 'completed', 'failed') NOT NULL,
  tasks TEXT NOT NULL,
  repository_name VARCHAR(255) NOT NULL,
  provider ENUM('claude', 'codex', 'cursor') NOT NULL,

  -- GitHub Integration Fields
  github_issue_url VARCHAR(500),
  github_pr_url VARCHAR(500),
  branch_name VARCHAR(255),
  cleanup_status VARCHAR(50),

  -- Timing Fields
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP NULL,
  processing_time_ms BIGINT,

  -- Results
  result TEXT,
  error TEXT
);
```

## 🚀 Usage

### Basic Task Execution

1. **Submit a Task**
   ```bash
   curl -X POST http://localhost:7000/api/v1/tasks \
     -H "Content-Type: application/json" \
     -d '{
       "tasks": "Add unit tests for authentication module",
       "repository_name": "JokerTrickster/my-app",
       "provider": "claude"
     }'
   ```

2. **Task Workflow**
   - GitHub issue automatically created
   - Feature branch created from main/master
   - Task executed in isolated environment
   - Changes committed and pushed
   - Pull request created with issue linking
   - Branch marked for cleanup after PR merge

### Frontend Usage

1. **Access the Web Interface**
   - Open http://localhost:3000
   - Navigate to Tasks tab
   - Select repository

2. **Create and Execute Tasks**
   - Click "Create Task"
   - Fill in task description
   - Select AI provider
   - Submit for execution

3. **Monitor Progress**
   - View task status in real-time
   - Click GitHub issue/PR links
   - Track branch progress

## 🔍 API Reference

### Task Management

#### Create Task
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "tasks": "Task description",
  "repository_name": "owner/repo",
  "provider": "claude",
  "working_dir": "/optional/path",
  "interactive": false,
  "cmd": "optional command"
}
```

#### Get Task Status
```http
GET /api/v1/tasks/{task_id}/status
```

#### Get Task History
```http
GET /api/v1/workflow-histories?repository_name=owner/repo&page=1&limit=20
```

### GitHub Integration

The system automatically handles:
- Issue creation with standardized templates
- Branch creation with unique naming
- PR creation with issue linking
- Cleanup after completion

## 🧪 Testing

### Run Unit Tests
```bash
cd local-backend
go test ./utils/... -v
```

### Run Integration Tests
```bash
cd local-backend
go test ./utils/ -tags=integration -v
```

### Frontend Tests
```bash
cd frontend
npm test
```

## 🔧 Advanced Configuration

### Branch Naming Strategy
```go
// Format: task/{sanitized-description}-{timestamp}-{hash}
// Example: task/add-unit-tests-1029-1504-a1b2c3d4
```

### Error Recovery Configuration
```go
// Configurable retry strategies
const (
    MaxRetries = 5
    BackoffDuration = 30 * time.Second
    RateLimitWait = 5 * time.Minute
)
```

### Cleanup Settings
```go
// Automatic cleanup of old branches
CleanupThreshold = 24 * time.Hour
OrphanedBranchCleanup = true
```

## 🐛 Troubleshooting

### Common Issues

1. **GitHub API Rate Limits**
   ```bash
   # Check rate limit status
   curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/rate_limit
   ```

2. **Branch Creation Failures**
   - Verify Git configuration
   - Check repository permissions
   - Ensure working directory exists

3. **PR Creation Issues**
   - Verify GitHub token permissions
   - Check if branch has changes
   - Ensure target repository exists

### Debug Logging

Enable detailed logging:
```go
log.SetLevel(log.DebugLevel)
```

## 📚 Development

### Adding New Features

1. **Extend GitHub Integration**
   - Add new endpoints in `GitHubService`
   - Update error handling in `ErrorHandler`
   - Add UI components in frontend

2. **Enhance Error Recovery**
   - Add new error types in `ErrorHandler`
   - Implement recovery strategies
   - Update classification logic

3. **Improve Branch Management**
   - Enhance naming strategies
   - Add conflict resolution
   - Implement advanced cleanup

### Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Support

For issues and questions:
- Create GitHub issue
- Check existing documentation
- Review error logs for debugging

## 🚧 Roadmap

- [ ] Multi-repository atomic operations
- [ ] Advanced merge conflict resolution
- [ ] Custom review workflows
- [ ] GitLab integration
- [ ] CI/CD pipeline integration
- [ ] Advanced branch protection rules