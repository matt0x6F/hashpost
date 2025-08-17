#!/bin/bash

# Database migration script for HashPost
# This script runs database migrations using sql-migrate

set -e

# Default values
MIGRATIONS_DIR="./internal/database/migrations"
DATABASE_URL="${DATABASE_URL:-postgres://hashpost:hashpost_dev@localhost:5432/hashpost?sslmode=disable}"
DRIVER="postgres"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[MIGRATION]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to wait for database to be ready
wait_for_database() {
    print_status "Waiting for database to be ready..."
    
    # Use environment variables directly instead of parsing DATABASE_URL
    # These should be set by Kubernetes from ConfigMap and Secrets
    if [[ -z "$DB_HOST" || -z "$DB_PORT" || -z "$DB_USER" || -z "$DB_PASSWORD" || -z "$DB_NAME" ]]; then
        print_error "Required database environment variables are not set"
        print_error "DB_HOST: $DB_HOST, DB_PORT: $DB_PORT, DB_USER: $DB_USER, DB_NAME: $DB_NAME"
        exit 1
    fi
    
    # Use the individual environment variables
    DB_PASS="$DB_PASSWORD"
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        print_warning "psql not found, skipping database readiness check"
        print_warning "Make sure the database is running and accessible"
        sleep 5  # Give a moment for the database to be ready
        return 0
    fi
    
    # Wait for database to be ready
    until PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
        print_status "Database is not ready yet. Waiting..."
        sleep 2
    done
    
    print_success "Database is ready!"
}

# Function to check if sql-migrate is installed
check_sql_migrate() {
    if ! command -v sql-migrate &> /dev/null; then
        print_error "sql-migrate is not installed and not available in PATH"
        print_error "This should be pre-installed in the container image"
        exit 1
    fi
    print_status "sql-migrate is available"
}

# Function to run migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Construct DATABASE_URL from individual environment variables
    # This is needed because Kubernetes doesn't expand $(VAR) syntax in env values
    if [[ -n "$DB_HOST" && -n "$DB_PORT" && -n "$DB_USER" && -n "$DB_PASSWORD" && -n "$DB_NAME" && -n "$DB_SSLMODE" ]]; then
        # URL-encode the password to handle special characters using shell
        # This handles common problematic characters in passwords
        DB_PASSWORD_ENCODED="$DB_PASSWORD"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//%/%25}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED// /%20}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//!/%21}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\"/%22}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//#/%23}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\$/%24}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//&/%26}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\'/%27}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//(/%28}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//)/%29}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\*/%2A}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//+/%2B}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//,/%2C}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//:/%3A}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//;/%3B}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//</%3C}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//=/%3D}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//>/%3E}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\?/%3F}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//@/%40}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\[/%5B}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\\/%5C}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\]/%5D}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\^/%5E}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\`/%60}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\{/%7B}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//|/%7C}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//\}/%7D}"
        DB_PASSWORD_ENCODED="${DB_PASSWORD_ENCODED//~/%7E}"
        export DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD_ENCODED}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
        print_status "Constructed DATABASE_URL from individual environment variables (password URL-encoded)"
    else
        print_status "Using existing DATABASE_URL environment variable"
    fi
    
    # Check current migration status
    print_status "Current migration status:"
    sql-migrate status -config=dbconfig.yml
    
    # Run pending migrations
    print_status "Applying pending migrations..."
    if sql-migrate up -config=dbconfig.yml; then
        print_success "Migrations completed successfully!"
        
        # Show final status
        print_status "Final migration status:"
        sql-migrate status -config=dbconfig.yml
    else
        print_error "Migration failed!"
        exit 1
    fi
}

# Function to create new migration
create_migration() {
    local name="$1"
    if [ -z "$name" ]; then
        print_error "Migration name is required"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi
    
    print_status "Creating new migration: $name"
    sql-migrate new "$name" -config=dbconfig.yml
    print_success "Migration file created!"
}

# Function to rollback migrations
rollback_migrations() {
    local steps="${1:-1}"
    print_status "Rolling back $steps migration(s)..."
    
    if sql-migrate down -config=dbconfig.yml -limit="$steps"; then
        print_success "Rollback completed successfully!"
    else
        print_error "Rollback failed!"
        exit 1
    fi
}

# Function to show help
show_help() {
    echo "HashPost Database Migration Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  up              Run pending migrations (default)"
    echo "  down [steps]    Rollback migrations (default: 1 step)"
    echo "  status          Show migration status"
    echo "  create <name>   Create a new migration file"
    echo "  help            Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DATABASE_URL    Database connection string"
    echo "  MIGRATIONS_DIR  Directory containing migration files"
    echo ""
    echo "Examples:"
    echo "  $0 up                    # Run all pending migrations"
    echo "  $0 down 2               # Rollback 2 migrations"
    echo "  $0 create add_users     # Create new migration"
    echo "  $0 status               # Show migration status"
}

# Main script logic
main() {
    local command="${1:-up}"
    
    case "$command" in
        "up")
            check_sql_migrate
            wait_for_database
            run_migrations
            ;;
        "down")
            check_sql_migrate
            wait_for_database
            rollback_migrations "$2"
            ;;
        "status")
            check_sql_migrate
            wait_for_database
            print_status "Migration status:"
            sql-migrate status -config=dbconfig.yml
            ;;
        "create")
            check_sql_migrate
            create_migration "$2"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 