# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

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
COPY scripts/init-ibe-keys.sh /usr/local/bin/init-ibe-keys.sh
RUN chmod +x /usr/local/bin/migrate.sh /usr/local/bin/entrypoint.sh /usr/local/bin/init-ibe-keys.sh

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

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration configuration and scripts
COPY dbconfig.yml ./
COPY scripts/migrate.sh /usr/local/bin/migrate.sh
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY scripts/init-ibe-keys.sh /usr/local/bin/init-ibe-keys.sh
RUN chmod +x /usr/local/bin/migrate.sh /usr/local/bin/entrypoint.sh /usr/local/bin/init-ibe-keys.sh

# Copy migrations directory
COPY internal/database/migrations ./internal/database/migrations

# Install sql-migrate in production
RUN apk add --no-cache go && \
    go install github.com/rubenv/sql-migrate/...@latest && \
    apk del go

# Expose port
EXPOSE 8888

# Use entrypoint script
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"] 