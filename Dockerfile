FROM golang:1.23-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install jq for JSON processing and ca-certificates for HTTPS
RUN apk --no-cache add jq ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Copy the binary from builder stage
COPY --from=builder /app/build/suppress-checker /usr/local/bin/suppress-checker

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Make sure the binary and script are executable
RUN chmod +x /usr/local/bin/suppress-checker /entrypoint.sh

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /github/workspace

# Set the entrypoint
ENTRYPOINT ["/entrypoint.sh"] 