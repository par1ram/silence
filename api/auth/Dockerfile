# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy go.mod and go.sum files for the auth service
COPY api/auth/go.mod api/auth/go.sum ./api/auth/

# Copy shared module
COPY shared ./shared

# Change to auth service directory
WORKDIR /app/api/auth

# Download dependencies
RUN go mod download

# Copy auth service source code
COPY api/auth ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth ./cmd/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc /app/migrations

# Copy the binary from builder stage
COPY --from=builder /app/api/auth/auth /app/auth

# Copy configuration files
COPY api/auth/etc/config.env /app/etc/config.env

# Copy migrations
COPY api/auth/internal/adapters/database/migrations /app/migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/auth"]
