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

# Function to URL-encode a string
url_encode() {
    local input="$1"
    # Use Python to robustly URL-encode the input
    python3 -c "import urllib.parse; print(urllib.parse.quote('$input'))" 2>/dev/null || \
    # Fallback to basic shell encoding for common characters
    echo "$input" | sed 's/%/%25/g; s/ /%20/g; s/!/%21/g; s/"/%22/g; s/#/%23/g; s/\$/%24/g; s/&/%26/g; s/'\''/%27/g; s/(/%28/g; s/)/%29/g; s/\*/%2A/g; s/+/%2B/g; s/,/%2C/g; s/-/%2D/g; s/\./%2E/g; s/\//%2F/g; s/:/%3A/g; s/;/%3B/g; s/</%3C/g; s/=/%3D/g; s/>/%3E/g; s/?/%3F/g; s/@/%40/g; s/\[/%5B/g; s/\\/%5C/g; s/\]/%5D/g; s/\^/%5E/g; s/_/%5F/g; s/`/%60/g; s/{/%7B/g; s/|/%7C/g; s/}/%7D/g; s/~/%7E/g'
}

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
    
    # Use individual environment variables if available (Kubernetes deployment)
    if [[ -n "$DB_HOST" && -n "$DB_PORT" && -n "$DB_USER" && -n "$DB_PASSWORD" && -n "$DB_NAME" ]]; then
        DB_PASS="$DB_PASSWORD"
        print_status "Using individual environment variables for database connection"
    else
        # Fallback to parsing DATABASE_URL (development mode)
        if [[ $DATABASE_URL =~ postgres://([^:]+):([^@]+)@([^:]+):([^/]+)/([^?]+) ]]; then
            DB_USER="${BASH_REMATCH[1]}"
            DB_PASS="${BASH_REMATCH[2]}"
            DB_HOST="${BASH_REMATCH[3]}"
            DB_PORT="${BASH_REMATCH[4]}"
            DB_NAME="${BASH_REMATCH[5]}"
            print_status "Using DATABASE_URL for database connection"
        else
            print_error "Invalid DATABASE_URL format and individual env vars not available"
            exit 1
        fi
    fi
    
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
        exec ./main server
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
    
    # Only run migrations in development mode (production uses init containers)
    if [ "$ENV" = "development" ]; then
        # Run database migrations
        run_migrations
    else
        print_status "Skipping migrations (handled by init container in production/testing)"
    fi
    
    # Construct DATABASE_URL from individual environment variables for the application
    # This is needed because Kubernetes doesn't expand $(VAR) syntax in env values
    if [[ -n "$DB_HOST" && -n "$DB_PORT" && -n "$DB_USER" && -n "$DB_PASSWORD" && -n "$DB_NAME" && -n "$DB_SSLMODE" ]]; then
        # URL-encode the password to handle special characters using shell
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
        print_status "Constructed DATABASE_URL from individual environment variables for application"
    fi
    
    # Initialize IBE keys
    initialize_ibe_keys
    
    # Start the application
    start_application
}

# Handle signals gracefully
trap 'print_status "Received signal, shutting down..."; exit 0' SIGTERM SIGINT

# Run main function
main "$@" 