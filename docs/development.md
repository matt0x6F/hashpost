# Development Setup Guide

## Overview

This guide covers setting up the HashPost development environment, including Docker setup, database management, and common development workflows.

## Prerequisites

### Required Software
- **Docker & Docker Compose**: For containerized development environment
- **Go 1.21+**: For backend development
- **Git**: For version control
- **Make**: For build automation (optional but recommended)

### Optional Software
- **PostgreSQL Client**: For direct database access
- **PlantUML**: For viewing database ERD diagrams
- **VS Code Extensions**: Go, Docker, PlantUML extensions

## Quick Start

### 1. Clone the Repository
```bash
git clone https://github.com/hashpost/hashpost.git
cd hashpost
```

### 2. Start Development Environment
```bash
# Start all services (PostgreSQL, Redis, API server)
make dev

# Or start services individually
docker-compose up -d postgres redis
make run
```

### 3. Verify Setup
```bash
# Check API health
curl http://localhost:8888/health

# Check database connection
make migrate-status
```

## Development Environment

### Docker Services

The development environment includes the following services:

#### PostgreSQL Database
- **Port**: 5432
- **Database**: hashpost
- **Username**: hashpost_user
- **Password**: hashpost_password
- **Auto-migration**: Enabled on startup

#### Redis Cache
- **Port**: 6379
- **Purpose**: Session storage and caching
- **Persistence**: Disabled in development

#### HashPost API Server
- **Port**: 8888
- **Environment**: Development mode
- **Auto-reload**: Enabled with air

### Environment Configuration

#### Required Environment Variables
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hashpost
DB_USER=hashpost_user
DB_PASSWORD=hashpost_password

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION=24h
JWT_DEVELOPMENT=true

# API Configuration
API_PORT=8888
API_HOST=0.0.0.0
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

#### Optional Environment Variables
```bash
# Security Configuration
SECURITY_ENABLE_MFA=false

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=console

# Development Configuration
ENVIRONMENT=development
DEBUG=true
```

### Docker Compose Configuration

The `docker-compose.yml` file defines the development environment:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: hashpost
      POSTGRES_USER: hashpost_user
      POSTGRES_PASSWORD: hashpost_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/entrypoint.sh:/docker-entrypoint-initdb.d/entrypoint.sh
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hashpost_user -d hashpost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly no

  api:
    build: .
    ports:
      - "8888:8888"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=hashpost
      - DB_USER=hashpost_user
      - DB_PASSWORD=hashpost_password
      - JWT_SECRET=your-super-secret-jwt-key-change-in-production
      - JWT_DEVELOPMENT=true
      - API_PORT=8888
      - API_HOST=0.0.0.0
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    volumes:
      - .:/app
    command: air -c .air.toml

volumes:
  postgres_data:
```

## Database Management

### Database Migrations

The project uses database migrations for schema management. Migrations are automatically applied when the development environment starts.

#### Migration Commands
```bash
# Create a new migration
make migrate-create name=descriptive_migration_name

# Apply migrations
make migrate-up

# Check migration status
make migrate-status

# Rollback migrations (if needed)
make migrate-down
```

#### Migration Workflow
1. **Create migration**: `make migrate-create name=your_migration_name`
2. **Edit migration file** in `internal/database/migrations/`
3. **Apply migration**: `make migrate-up`
4. **Verify status**: `make migrate-status`
5. **Generate models**: `make generate`
6. **Test compilation**: `go build ./...`

### Database Access

#### Direct Database Access
```bash
# Connect to PostgreSQL
psql -h localhost -p 5432 -U hashpost_user -d hashpost

# Or using Docker
docker-compose exec postgres psql -U hashpost_user -d hashpost
```

#### Database Management Tools
- **pgAdmin**: Web-based PostgreSQL administration
- **DBeaver**: Universal database tool
- **TablePlus**: Modern database client

### Model Generation

After applying database migrations, regenerate the Bob models:

```bash
# Generate models
make generate

# Clean generated files (if needed)
make clean
```

## Development Workflow

### Common Commands

#### Build and Run
```bash
# Build the application
make build

# Run the application
make run

# Run with hot reload (using air)
make dev

# Stop all services
make stop
```

#### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test file
go test ./internal/api/handlers/ -v
```

#### Code Quality
```bash
# Format code
make fmt

# Lint code
make lint

# Run security checks
make security
```

### Development Tips

#### Hot Reload
The development environment uses [Air](https://github.com/cosmtrek/air) for hot reloading:

```bash
# Start with hot reload
make dev

# Or manually
air -c .air.toml
```

#### Debugging
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run with debug information
go run -race cmd/server/main.go
```

#### Database Reset
```bash
# Reset database (WARNING: destroys all data)
make db-reset

# Reset and apply migrations
make db-reset && make migrate-up
```

## API Development

### API Endpoints

The API is available at `http://localhost:8888` with the following endpoints:

#### Health Check
```bash
curl http://localhost:8888/health
```

#### Authentication
```bash
# Register user
curl -X POST http://localhost:8888/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password","display_name":"testuser"}'

# Login
curl -X POST http://localhost:8888/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

#### Content Management
```bash
# Get posts (requires authentication)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/subforums/golang/posts
```

### API Documentation

- **Interactive API Docs**: Available at `http://localhost:8888/docs` (when implemented)
- **OpenAPI Spec**: Available at `http://localhost:8888/openapi.json` (when implemented)

## Testing

### Test Structure

```
internal/
├── api/
│   ├── handlers/
│   │   ├── auth_test.go
│   │   ├── content_test.go
│   │   └── ...
│   └── middleware/
│       ├── auth_test.go
│       └── ...
├── database/
│   ├── dao/
│   │   ├── users_test.go
│   │   └── ...
│   └── models/
│       └── ...
└── ibe/
    └── ibe_test.go
```

### Running Tests

#### Unit Tests
```bash
# Run all tests
make test

# Run specific package
go test ./internal/api/handlers/

# Run with verbose output
go test -v ./internal/api/handlers/

# Run with race detection
go test -race ./internal/api/handlers/
```

#### Integration Tests
```bash
# Run integration tests (requires database)
make test-integration

# Run with test database
DB_NAME=hashpost_test make test-integration
```

#### Test Coverage
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out
```

### Writing Tests

#### Example Test Structure
```go
func TestUserRegistration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    handler := NewAuthHandler(db)
    
    // Test cases
    tests := []struct {
        name    string
        input   *models.UserRegistrationInput
        wantErr bool
    }{
        {
            name: "valid registration",
            input: &models.UserRegistrationInput{
                Email:       "test@example.com",
                Password:    "password",
                DisplayName: "testuser",
            },
            wantErr: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := handler.RegisterUser(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Add assertions
        })
    }
}
```

## Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Reset database
make db-reset
```

#### Migration Issues
```bash
# Check migration status
make migrate-status

# Reset migrations
make migrate-down
make migrate-up
```

#### Build Issues
```bash
# Clean build artifacts
make clean

# Rebuild from scratch
docker-compose build --no-cache
```

#### Port Conflicts
```bash
# Check what's using port 8888
lsof -i :8888

# Use different port
API_PORT=8889 make dev
```

### Debugging

#### Enable Debug Logging
```bash
export LOG_LEVEL=debug
export DEBUG=true
make dev
```

#### Database Debugging
```bash
# Connect to database
docker-compose exec postgres psql -U hashpost_user -d hashpost

# Check tables
\dt

# Check recent logs
SELECT * FROM system_events ORDER BY timestamp DESC LIMIT 10;
```

#### API Debugging
```bash
# Check API logs
docker-compose logs api

# Test API endpoints
curl -v http://localhost:8888/health
```

## Performance Monitoring

### Development Metrics

#### Database Performance
```sql
-- Check slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

#### Application Metrics
```bash
# Check memory usage
docker stats

# Check API response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8888/health
```

## Deployment Preparation

### Production Checklist

Before deploying to production:

- [ ] Update environment variables for production
- [ ] Set secure JWT secret
- [ ] Configure HTTPS
- [ ] Set up proper logging
- [ ] Configure database backups
- [ ] Set up monitoring and alerting
- [ ] Review security settings
- [ ] Test all migrations
- [ ] Verify API endpoints

### Environment-Specific Configuration

#### Development
```bash
JWT_DEVELOPMENT=true
DEBUG=true
LOG_LEVEL=debug
```

#### Production
```bash
JWT_DEVELOPMENT=false
DEBUG=false
LOG_LEVEL=info
SECURITY_ENABLE_MFA=true
```

## Additional Resources

### Documentation
- [API Documentation](api-documentation.md)
- [Database Schema](database-schema.md)
- [Authentication Guide](authentication.md)
- [Identity-Based Encryption](identity-based-encryption.md)

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Huma Framework](https://huma.rocks/)

### Community
- [GitHub Issues](https://github.com/hashpost/hashpost/issues)
- [Discussions](https://github.com/hashpost/hashpost/discussions)
- [Security Reporting](mailto:security@hashpost.com) 