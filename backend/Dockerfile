# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
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

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Copy binary from builder stage
COPY --from=builder /app/bin/node /usr/local/bin/node

# Change ownership
RUN chown appuser:appuser /usr/local/bin/node

# Switch to non-root user
USER appuser

# Expose port (libp2p will use random ports, but we expose a default)
EXPOSE 8080

# Run the application
CMD ["node"]
