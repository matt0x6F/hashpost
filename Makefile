# HashPost Makefile
# Provides convenient commands for development and deployment

.PHONY: help build run test clean migrate migrate-up migrate-down migrate-status migrate-create docker-build docker-up docker-down docker-logs generate test-integration-local test-models test-models-setup ui-install ui-generate-api ui-dev ui-build test-coverage test-coverage-html test-coverage-ci test-coverage-report

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
	@echo "  test-coverage   Run unit tests with coverage report"
	@echo "  test-coverage-html Run tests and generate HTML coverage report"
	@echo "  test-coverage-ci Run tests with coverage for CI (fails if < 70%)"
	@echo "  test-coverage-report Generate detailed coverage report"
	@echo "  test-integration-local Run integration tests with clean DB (includes model tests)"
	@echo "                         Usage: make test-integration-local TEST_PATH=./internal/api/integration/auth_integration_test.go"
	@echo "  test-dao        Run DAO integration tests with test database"
	@echo "  test-dao-ci     Run DAO integration tests for CI environment"
	@echo "  test-dao-setup-only Setup test database only (for debugging)"
	@echo "  test-dao-cleanup-only Cleanup test database only (for debugging)"
	@echo "  test-dao-pattern Run DAO tests matching pattern (PATTERN=TestPostDAO)"
	@echo "  test-models     Run database model tests only"
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
	@echo "  setup-ibe-keys  Setup enhanced IBE keys with domain separation"
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

# Development commands
build:
	@echo "Building application..."
	go build -o bin/hashpost ./cmd/server

run:
	@echo "Running application locally..."
	go run ./cmd/server
 
test: test-unit

test-unit:
	@echo "Running unit tests..."
	PSQL_DSN=postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable go test ./... -v

# Code Coverage Targets
test-coverage:
	@echo "🧪 Running unit tests with coverage..."
	@PSQL_DSN=postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable TEST_DATABASE_URL=postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable go test ./... -v -coverprofile=coverage.out
	@echo "📊 Coverage Summary:"
	@go tool cover -func=coverage.out | tail -1
	@echo "📁 Coverage data saved to: coverage.out"

test-coverage-html: test-coverage
	@echo "🌐 Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📄 HTML coverage report generated: coverage.html"
	@echo "💡 Open coverage.html in your browser to view detailed coverage"

test-coverage-ci:
	@echo "🧪 Running tests with coverage for CI..."
	@PSQL_DSN=postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable TEST_DATABASE_URL=postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable go test ./... -v -coverprofile=coverage.out -covermode=atomic
	@echo "📊 Coverage Summary:"
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 70" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage is below 70% threshold ($$COVERAGE%)"; \
		exit 1; \
	else \
		echo "✅ Coverage meets threshold ($$COVERAGE% >= 70%)"; \
	fi

test-coverage-report:
	@echo "📋 Generating detailed coverage report..."
	@if [ ! -f coverage.out ]; then \
		echo "No coverage data found. Running tests first..."; \
		make test-coverage; \
	fi
	@echo "📊 Coverage by package:"
	@go tool cover -func=coverage.out
	@echo ""
	@echo "📈 Coverage summary:"
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "🔍 Files with low coverage (< 80%):"
	@go tool cover -func=coverage.out | grep -E "\.go:[0-9]+" | awk '$$3 < 80 {print $$1 ":" $$3 "%"}' | head -10
	@echo ""
	@echo "💡 Run 'make test-coverage-html' to generate visual report"

# DAO Integration Tests
test-dao: test-dao-setup test-dao-run test-dao-cleanup
	@echo "✅ DAO integration tests completed successfully!"

test-dao-setup:
	@echo "🔧 Setting up test database for DAO integration tests..."
	@echo "Starting test PostgreSQL container..."
	docker-compose --profile test up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@sleep 3
	@echo "Applying migrations to test database..."
	@DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" sql-migrate up -config=dbconfig.yml -env=test
	@echo "✅ Test database setup complete!"

test-dao-run:
	@echo "🧪 Running DAO integration tests..."
	@TEST_DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" go test ./internal/database/dao/ -v

test-dao-cleanup:
	@echo "🧹 Cleaning up test database..."
	@docker-compose --profile test down postgres-test
	@echo "✅ Test database cleanup complete!"

# CI-specific DAO tests (uses GitHub Actions PostgreSQL service)
test-dao-ci: test-dao-ci-setup test-dao-ci-run
	@echo "✅ DAO integration tests completed successfully!"

test-dao-ci-setup:
	@echo "🔧 Setting up test database for CI DAO integration tests..."
	@echo "Applying migrations to test database..."
	@sql-migrate up -config=dbconfig.yml -env=ci
	@echo "✅ CI test database setup complete!"

test-dao-ci-run:
	@echo "🧪 Running DAO integration tests in CI..."
	@TEST_DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5432/hashpost_test?sslmode=disable" go test ./internal/database/dao/ -v

# Individual DAO test targets for debugging
test-dao-setup-only:
	@echo "🔧 Setting up test database only (no cleanup)..."
	@echo "Starting test PostgreSQL container..."
	docker-compose --profile test up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@sleep 3
	@echo "Applying migrations to test database..."
	@DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" sql-migrate up -config=dbconfig.yml -env=test
	@echo "✅ Test database setup complete!"
	@echo "💡 Run 'make test-dao-cleanup-only' when done testing"

test-dao-cleanup-only:
	@echo "🧹 Cleaning up test database..."
	@docker-compose --profile test down postgres-test
	@echo "✅ Test database cleanup complete!"

# Run specific DAO test pattern
test-dao-pattern: test-dao-setup
	@if [ -z "$(PATTERN)" ]; then \
		echo "Usage: make test-dao-pattern PATTERN=TestPostDAO"; \
		exit 1; \
	fi
	@echo "🧪 Running DAO tests matching pattern: $(PATTERN)"
	@TEST_DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" go test ./internal/database/dao/ -run $(PATTERN) -v
	@make test-dao-cleanup-only


# Setup test database for VSCode test runner
setup-test-db:
	@echo "Setting up test database for VSCode test runner..."
	@echo "Starting test database..."
	@docker-compose --profile test up -d postgres-test
	@sleep 3
	@echo "Applying migrations..."
	@DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" sql-migrate up -config=dbconfig.yml -env=test
	@echo "Test database ready for tests"

# Legacy test-integration-local target (for backward compatibility)
test-integration-local:
	@echo "Running integration tests with clean DB..."
	@if [ -z "$(TESTS)" ]; then \
		echo "Usage: make test-integration-local TESTS=./internal/api/integration/auth_integration_test.go"; \
		exit 1; \
	fi
	@echo "Starting test database..."
	@docker-compose --profile test up -d postgres-test
	@sleep 3
	@echo "Applying migrations..."
	@DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" sql-migrate up -config=dbconfig.yml -env=test
	@echo "Running tests: $(TESTS)"
	@TEST_DATABASE_URL="postgres://hashpost:hashpost_test@localhost:5433/hashpost_test?sslmode=disable" go test $(TESTS) -v
	@echo "Cleaning up..."
	@docker-compose --profile test down postgres-test

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
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
	cd internal/database && go run github.com/stephenafamo/bob/gen/bobgen-psql@v0.38.0 -c ../../bobgen.yaml

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
	@echo "Setting up IBE keys..."
	@echo "Generating enhanced IBE keys with domain separation..."
	./bin/hashpost generate-ibe-keys --output-dir ./keys --generate-new --non-interactive
	@echo "✅ IBE keys generated successfully!"
	@echo "📁 Keys location: ./keys/"
	@echo "🔐 Master key: ./keys/master.key"
	@echo "📋 Configuration: ./keys/ibe_config.json" 

setup-roles:
	@echo "Setting up role keys for all roles..."
	docker-compose exec app ./tmp/main setup-roles 
