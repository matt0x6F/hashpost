# HashPost Makefile
# Provides convenient commands for development and deployment

.PHONY: help build run test clean migrate migrate-up migrate-down migrate-status migrate-create docker-build docker-up docker-down docker-logs generate test-integration docker-test-up docker-test-down test-integration-local test-integration-vscode ui-install ui-generate-api ui-dev ui-build

# Default target
help:
	@echo "HashPost Development Commands"
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-up      Run pending database migrations"
	@echo "  migrate-down    Rollback last migration"
	@echo "  migrate-status  Show migration status"
	@echo "  migrate-create  Create a new migration file"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build    Build Docker images"
	@echo "  docker-up       Start development environment"
	@echo "  docker-down     Stop development environment"
	@echo "  docker-logs     Show application logs"
	@echo "  docker-prod     Start production environment"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run unit tests"
	@echo "  test-integration Run integration tests (requires test database)"
	@echo "  test-integration-local Run integration tests with clean DB (defaults to all tests)"
	@echo "                         Usage: make test-integration-local TEST_PATH=./internal/api/integration/auth_integration_test.go"
	@echo "  docker-test-up  Start test environment"
	@echo "  docker-test-down Stop test environment"
	@echo ""
	@echo "UI Development:"
	@echo "  ui-install      Install UI dependencies"
	@echo "  ui-generate-api Generate TypeScript API client from OpenAPI schema"
	@echo "  ui-dev          Start UI development server"
	@echo "  ui-build        Build UI for production"
	@echo ""
	@echo "Development:"
	@echo "  build           Build the application"
	@echo "  run             Run the application locally"
	@echo "  clean           Clean build artifacts"
	@echo ""
	@echo "Setup:"
	@echo "  setup-ibe-keys  Setup IBE master keys"
	@echo "  setup-roles     Setup role keys for all roles"

# Database migration commands (run inside Docker Compose app container)
migrate-up:
	@echo "Running database migrations in Docker Compose app container..."
	docker-compose exec app ./scripts/migrate.sh up

migrate-down:
	@echo "Rolling back last migration in Docker Compose app container..."
	docker-compose exec app ./scripts/migrate.sh down

migrate-status:
	@echo "Migration status in Docker Compose app container:"
	docker-compose exec app ./scripts/migrate.sh status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name) in Docker Compose app container"
	docker-compose exec app ./scripts/migrate.sh create $(name)

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting development environment..."
	docker-compose up -d --build

docker-down:
	@echo "Stopping development environment..."
	docker-compose down

docker-logs:
	@echo "Showing application logs..."
	docker-compose logs -f app

docker-prod:
	@echo "Starting production environment..."
	docker-compose --profile production up -d

# Test environment commands
docker-test-up:
	@echo "Starting test environment..."
	docker-compose --profile test up -d --build

docker-test-down:
	@echo "Stopping test environment..."
	docker-compose --profile test down

# Development commands
build:
	@echo "Building application..."
	go build -o bin/hashpost ./cmd/server

run:
	@echo "Running application locally..."
	go run ./cmd/server
 
test: test-unit test-integration-local

test-unit:
	@echo "Running unit tests..."
	go test ./...

test-integration:
	@echo "Running integration tests..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL environment variable is required for integration tests"; \
		echo "Example: DATABASE_URL='postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable' make test-integration"; \
		echo "Or use: make docker-test-up to start test environment"; \
		exit 1; \
	fi
	go test -v -tags=integration ./...

test-integration-local:
	@echo "Setting up clean test database..."
	@echo "Starting test PostgreSQL container..."
	docker-compose --profile test up -d postgres-test
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Ensuring test database exists and is migrated..."
	@docker-compose --profile test exec -T postgres-test psql -U hashpost -d postgres -c "DROP DATABASE IF EXISTS hashpost_test;" || true
	@docker-compose --profile test exec -T postgres-test psql -U hashpost -d postgres -c "CREATE DATABASE hashpost_test;" || true
	@DATABASE_URL='postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable' ./scripts/migrate.sh up
	@echo "Running integration tests..."
	@LOG_LEVEL=$${LOG_LEVEL:-error} DATABASE_URL='postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable' go test -v -tags=integration $${TEST_PATH:-./...}

# For VSCode test runner compatibility (runs integration tests if DATABASE_URL is set)
test-integration-vscode:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL environment variable is required for integration tests"; \
		echo "Example: DATABASE_URL='postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable' make test-integration-vscode"; \
		exit 1; \
	fi
	go test -v -tags=integration ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Database commands (local)
db-create:
	@echo "Creating database in Docker Compose PostgreSQL..."
	docker-compose exec postgres createdb -U hashpost hashpost || true

db-drop:
	@echo "Dropping database in Docker Compose PostgreSQL..."
	@echo "Terminating active connections..."
	docker-compose exec postgres psql -U hashpost -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'hashpost' AND pid <> pg_backend_pid();" || true
	@echo "Dropping database..."
	docker-compose exec postgres dropdb -U hashpost hashpost || true

db-reset: db-drop db-create migrate-up
	@echo "Database reset complete"

# Utility commands
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/rubenv/sql-migrate/...@latest

setup-dev: install-tools
	@echo "Setting up development environment..."
	@if ! command -v docker &> /dev/null; then \
		echo "Docker is required but not installed"; \
		exit 1; \
	fi
	@if ! command -v docker-compose &> /dev/null; then \
		echo "Docker Compose is required but not installed"; \
		exit 1; \
	fi
	@echo "Development environment setup complete"

# Show help by default
.DEFAULT_GOAL := help

generate:
	cd internal/database && go run github.com/stephenafamo/bob/gen/bobgen-psql@latest -c ../../bobgen.yaml

# UI Development commands
ui-install:
	@echo "Installing UI dependencies..."
	cd ui && npm install

ui-generate-api:
	@echo "Generating TypeScript API client from OpenAPI schema..."
	@echo "Make sure the HashPost server is running (make dev)"
	cd ui && npm run generate-api

ui-dev:
	@echo "Starting UI development server..."
	cd ui && npm run dev

ui-build:
	@echo "Building UI for production..."
	cd ui && npm run build

# IBE Key Management
setup-ibe-keys:
	@echo "Setting up IBE master keys..."
	./scripts/setup-ibe-keys.sh 

setup-roles:
	@echo "Setting up role keys for all roles..."
	docker-compose exec app ./tmp/main setup-roles 
