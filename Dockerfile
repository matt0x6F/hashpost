# Development stage with Air for hot reloading
FROM golang:1.24-alpine AS dev

WORKDIR /app

# Install Air for hot reloading (use compatible version for Go 1.24)
RUN go install github.com/cosmtrek/air@v1.49.0

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Create Air configuration
COPY .air.toml ./

# Expose ports
EXPOSE 8080 8081

# Default command for development
CMD ["air", "-c", ".air.toml"]
