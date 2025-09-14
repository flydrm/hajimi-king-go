# Multi-stage build for Hajimi King Go v2.0
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hajimi-king-v2 ./cmd/app

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh hajimi

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/hajimi-king-v2 .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy example configuration
COPY --from=builder /app/.env.example ./.env.example
COPY --from=builder /app/queries.txt ./queries.txt

# Create data directory
RUN mkdir -p data logs cache

# Change ownership to non-root user
RUN chown -R hajimi:hajimi /app

# Switch to non-root user
USER hajimi

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./hajimi-king-v2"]