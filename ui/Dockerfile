# Use Node.js 18 Alpine as base image
FROM node:18-alpine

# Install Java (required for OpenAPI generator)
RUN apk add --no-cache openjdk11-jre

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install all dependencies (including dev dependencies for build)
RUN npm ci

# Copy OpenAPI spec and generate API client
COPY openapi.yaml ./
COPY openapi-generator-config.json ./
RUN npm run generate-api

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Remove dev dependencies to reduce image size
RUN npm prune --production

# Expose port
EXPOSE 3000

# Start the application
# Use dev mode if NODE_ENV is development, otherwise use production
CMD ["sh", "-c", "if [ \"$NODE_ENV\" = \"development\" ]; then npm run dev; else npm start; fi"]