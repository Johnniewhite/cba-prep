# CBA Lite Backend

A scalable, production-ready backend for the CBA Lite real-time team collaboration application built with Go, PostgreSQL, Redis, and WebSockets.

## Features

- **Real-time Communication**: WebSocket-based messaging system
- **Task Management**: Create, assign, and track tasks
- **Team Collaboration**: Teams, channels, and direct messaging
- **Authentication**: JWT-based authentication with refresh tokens
- **Caching**: Redis integration for performance optimization
- **Rate Limiting**: API rate limiting to prevent abuse
- **Logging**: Structured logging with request tracking
- **Database**: PostgreSQL with migration support
- **Docker**: Full Docker and Docker Compose support

## Architecture

```
backend/
├── cmd/api/           # Application entry points
├── internal/          # Private application code
│   ├── cache/        # Redis cache implementation
│   ├── config/       # Configuration management
│   ├── database/     # Database connections
│   ├── domain/       # Domain models and entities
│   ├── handlers/     # HTTP request handlers
│   ├── middleware/   # HTTP middleware
│   ├── repository/   # Data access layer
│   ├── services/     # Business logic
│   └── websocket/    # WebSocket implementation
├── pkg/              # Public packages
│   ├── errors/       # Error handling
│   ├── logger/       # Logging utilities
│   └── validation/   # Input validation
├── migrations/       # Database migrations
└── deployments/      # Deployment configurations
```

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)

## Quick Start

### 1. Clone and Setup

```bash
cd backend
cp .env.example .env
# Edit .env with your configuration
```

### 2. Install Dependencies

```bash
make deps
```

### 3. Run with Docker Compose

```bash
make docker-up
```

### 4. Run Locally

```bash
# Start PostgreSQL and Redis (if not using remote)
make dev

# Or run directly
make run
```

## Development

### Available Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make lint          # Run linter
make fmt           # Format code
make migrate-up    # Run database migrations
make docker-up     # Start with Docker Compose
make dev           # Start development environment
```

### API Endpoints

#### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout

#### Teams
- `GET /api/v1/teams` - List user's teams
- `POST /api/v1/teams` - Create new team
- `GET /api/v1/teams/{id}` - Get team details
- `PUT /api/v1/teams/{id}` - Update team
- `DELETE /api/v1/teams/{id}` - Delete team

#### Messages
- `POST /api/v1/channels/{id}/messages` - Send message
- `GET /api/v1/channels/{id}/messages` - Get messages
- `PUT /api/v1/messages/{id}` - Update message
- `DELETE /api/v1/messages/{id}` - Delete message

#### Tasks
- `POST /api/v1/teams/{id}/tasks` - Create task
- `GET /api/v1/teams/{id}/tasks` - List tasks
- `GET /api/v1/tasks/{id}` - Get task details
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task

#### WebSocket
- `WS /api/v1/ws` - WebSocket connection for real-time updates

## Environment Variables

Key environment variables (see `.env.example` for full list):

```bash
# Application
APP_PORT=8080
APP_ENV=development

# Database
DB_HOST=168.231.113.231
DB_PORT=5432
DB_USER=cba
DB_PASSWORD=cbaPREP2025
DB_NAME=cbalite

# Redis
REDIS_ADDR=redis-17759.c258.us-east-1-4.ec2.redns.redis-cloud.com:17759
REDIS_USERNAME=default
REDIS_PASSWORD=QaHL8XUAMlUau0n2zKUgs902KI1pQaDk

# JWT
JWT_SECRET_KEY=your-secret-key
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d
```

## Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create name=add_new_table
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/services/...
```

## Production Deployment

### Using Docker

```bash
# Build production image
docker build -t cbalite-backend:prod .

# Run container
docker run -d \
  --name cbalite-backend \
  -p 8080:8080 \
  --env-file .env \
  cbalite-backend:prod
```

### Health Check

The application provides a health check endpoint:

```bash
curl http://localhost:8080/api/v1/health
```

## Security Features

- JWT-based authentication with refresh tokens
- Password hashing with bcrypt
- Rate limiting per IP address
- CORS configuration
- TLS/SSL support
- Input validation and sanitization
- SQL injection prevention
- XSS protection

## Performance Optimizations

- Redis caching for frequently accessed data
- Connection pooling for database
- Efficient WebSocket message broadcasting
- Pagination for list endpoints
- Database query optimization with indexes
- Graceful shutdown handling

## Monitoring

- Structured logging with request IDs
- Error tracking and recovery
- Health check endpoints
- Metrics collection ready

## Contributing

1. Follow Go best practices and idioms
2. Write tests for new features
3. Run `make check` before committing
4. Update documentation as needed

## License

Private and Confidential