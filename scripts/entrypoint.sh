#!/bin/bash

# HashPost Container Entrypoint Script
# This script runs database migrations and then starts the application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[ENTRYPOINT]${NC} $1"
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
    
    # Extract connection details from DATABASE_URL
    if [[ $DATABASE_URL =~ postgres://([^:]+):([^@]+)@([^:]+):([^/]+)/([^?]+) ]]; then
        DB_USER="${BASH_REMATCH[1]}"
        DB_PASS="${BASH_REMATCH[2]}"
        DB_HOST="${BASH_REMATCH[3]}"
        DB_PORT="${BASH_REMATCH[4]}"
        DB_NAME="${BASH_REMATCH[5]}"
    else
        print_error "Invalid DATABASE_URL format"
        exit 1
    fi
    
    # Wait for database to be ready
    until PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
        print_status "Database is not ready yet. Waiting..."
        sleep 2
    done
    
    print_success "Database is ready!"
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Check if sql-migrate is available
    if command -v sql-migrate &> /dev/null; then
        # Run migrations
        if sql-migrate up -config=dbconfig.yml; then
            print_success "Database migrations completed successfully!"
        else
            print_error "Database migrations failed!"
            exit 1
        fi
    else
        print_warning "sql-migrate not found, skipping migrations"
    fi
}

# Function to initialize IBE keys
initialize_ibe_keys() {
    print_status "Checking IBE keys..."
    
    # Check if keys directory exists and has content
    if [ -d "/app/keys" ] && [ "$(ls -A /app/keys 2>/dev/null)" ]; then
        print_success "IBE keys found in /app/keys, skipping generation"
        return 0
    fi
    
    print_status "No IBE keys found, generating new keys..."
    
    # Check if the application binary is available and can generate IBE keys
    if command -v ./main &> /dev/null; then
        if ./main generate-ibe-keys --output-dir /app/keys --generate-new --non-interactive; then
            print_success "IBE key generation completed successfully!"
        else
            print_warning "IBE key generation failed, continuing without IBE keys"
        fi
    else
        print_warning "Application binary not found, skipping IBE key generation"
    fi
}

# Function to start the application
start_application() {
    print_status "Starting HashPost application..."
    
    # Check if we're in development mode (using Air)
    if [ "$ENV" = "development" ] && command -v air &> /dev/null; then
        print_status "Starting in development mode with Air..."
        exec air -c .air.toml
    else
        print_status "Starting in production mode..."
        exec ./main
    fi
}

# Main entrypoint logic
main() {
    print_status "HashPost container starting..."
    
    # Set default environment if not provided
    export ENV="${ENV:-production}"
    export DATABASE_URL="${DATABASE_URL:-postgres://hashpost:hashpost_dev@postgres:5432/hashpost?sslmode=disable}"
    
    print_status "Environment: $ENV"
    print_status "Database URL: $DATABASE_URL"
    
    # Wait for database to be ready
    wait_for_database
    
    # Run database migrations
    run_migrations
    
    # Initialize IBE keys
    initialize_ibe_keys
    
    # Start the application
    start_application
}

# Handle signals gracefully
trap 'print_status "Received signal, shutting down..."; exit 0' SIGTERM SIGINT

# Run main function
main "$@" 