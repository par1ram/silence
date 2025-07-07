# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app/api/gateway

# Install git for Go modules
RUN apk add --no-cache git

# Copy shared library first
COPY shared /app/shared

# Copy go.mod and go.sum files
COPY api/gateway/go.mod api/gateway/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY api/gateway ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc

# Copy the binary from builder stage
COPY --from=builder /app/api/gateway/gateway /app/gateway

# Copy configuration files
COPY api/gateway/etc/config.env /app/etc/config.env

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/gateway"]
