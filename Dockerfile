# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN go build -o bin/hashpost-pds ./cmd/pds
RUN go build -o bin/hashpost-appview ./cmd/appview

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binaries
COPY --from=builder /app/bin/ /app/bin/

# Copy config files
COPY --from=builder /app/config/ /app/config/

# Make binaries executable
RUN chmod +x /app/bin/*

# Default command
CMD ["/app/bin/hashpost-pds"]
