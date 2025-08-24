# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached if go.mod/go.sum don't change)
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o main ./cmd/server

# Development stage with Air
FROM golang:1.24-alpine AS development

WORKDIR /app

# Install git, ca-certificates, Air, sql-migrate, PostgreSQL client, and bash
RUN apk add --no-cache git ca-certificates postgresql-client bash && \
    go install github.com/air-verse/air@latest && \
    go install github.com/rubenv/sql-migrate/...@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy Air configuration
COPY .air.toml ./

# Copy migration scripts and make them executable
COPY scripts/migrate.sh /usr/local/bin/migrate.sh
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/migrate.sh /usr/local/bin/entrypoint.sh

# Expose port
EXPOSE 8888

# Set environment variable for Air
ENV AIR_WD=/app

# Use entrypoint script
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Production stage
FROM alpine:latest AS production

# Install ca-certificates, PostgreSQL client, and bash
RUN apk --no-cache add ca-certificates postgresql-client bash

# Create non-root user for security
RUN addgroup -g 1001 -S hashpost && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G hashpost -g hashpost hashpost

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# OpenAPI specification will be fetched from running service during CI/CD

# Copy migration configuration and scripts
COPY dbconfig.yml ./
COPY scripts/migrate.sh /usr/local/bin/migrate.sh
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/migrate.sh /usr/local/bin/entrypoint.sh

# Copy migrations directory
COPY internal/database/migrations ./internal/database/migrations

# Install sql-migrate in production (using a more efficient approach)
RUN apk add --no-cache --virtual .build-deps go git && \
    GOBIN=/usr/local/bin go install github.com/rubenv/sql-migrate/...@latest && \
    apk del .build-deps

# Change ownership to non-root user
RUN chown -R hashpost:hashpost /app

# Switch to non-root user
USER hashpost

# Expose port
EXPOSE 8888

# Use entrypoint script
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"] 