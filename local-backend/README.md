# Local Backend Server

Always-on Go service that processes RabbitMQ messages and sends code analysis requests to Claude API.

## Architecture

This service follows Clean Architecture principles with the following layers:

- **Domain Layer** (`internal/domain/`): Core business entities and rules
- **Use Case Layer** (`internal/usecase/`): Application business logic  
- **Repository Layer** (`internal/repository/`): Data access abstractions
- **Infrastructure Layer** (`internal/infrastructure/`): External concerns (DB, RabbitMQ, Claude API)

## Features

- **RabbitMQ Consumer**: Continuously processes messages from queue
- **Claude API Integration**: Sends code analysis requests to Anthropic's Claude
- **SQLite Database**: Local persistence for request tracking and status
- **Context Management**: Maintains conversation context across related requests
- **Configuration Management**: Environment-based configuration with Viper

## Setup

### Prerequisites

- Go 1.21 or higher
- RabbitMQ server running locally
- Claude API key from Anthropic

### Installation

1. Clone and navigate to the local-backend directory:
   ```bash
   cd local-backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Copy and configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Configure RabbitMQ and Claude API settings in `configs/config.yaml` or via environment variables.

### Configuration

Key configuration options:

- `RABBITMQ_URL`: RabbitMQ connection string (default: amqp://localhost:5672/)
- `RABBITMQ_QUEUE`: Queue name to consume from (default: task_queue)
- `CLAUDE_API_KEY`: Your Claude API key (required)
- `DATABASE_DSN`: SQLite database file path (default: ./local_backend.db)

See `.env.example` for all available options.

### Running

```bash
# Build the application
go build -o bin/server ./cmd/server

# Run the server
./bin/server

# Or run directly
go run ./cmd/server
```

### Message Format

The service expects JSON messages in the following format:

```json
{
  "type": "work_request",
  "id": "unique-request-id", 
  "payload": {
    "code": "// Code to analyze",
    "task": "Analyze this Go function for potential bugs"
  },
  "context_id": "optional-context-id"
}
```

Supported message types:
- `work_request`: Code analysis request
- `work_cancellation`: Cancel pending request

## Development

### Project Structure

```
local-backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Domain layer (entities, interfaces)
│   ├── usecase/         # Use cases/application layer  
│   ├── repository/      # Data access layer
│   └── infrastructure/  # External concerns (DB, config, clients)
├── pkg/                 # Public packages
├── configs/             # Configuration files
└── README.md
```

### Building

```bash
# Build for current platform
go build ./cmd/server

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain
```

## Status

This is the foundation setup for the local backend server. Core functionality will be implemented in subsequent tasks:

- [ ] Domain models and entities
- [ ] Database infrastructure  
- [ ] RabbitMQ consumer integration
- [ ] Claude API service implementation
- [ ] Application services and workflow
- [ ] Comprehensive testing suite