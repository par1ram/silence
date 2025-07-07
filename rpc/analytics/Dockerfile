# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy go work files first for better caching
COPY go.work go.work.sum ./

# Copy all go.mod files referenced in go.work
COPY api/auth/go.mod api/auth/go.sum ./api/auth/
COPY api/gateway/go.mod api/gateway/go.sum ./api/gateway/
COPY rpc/analytics/go.mod rpc/analytics/go.sum ./rpc/analytics/
COPY rpc/dpi-bypass/go.mod rpc/dpi-bypass/go.sum ./rpc/dpi-bypass/
COPY rpc/notifications/go.mod rpc/notifications/go.sum ./rpc/notifications/
COPY rpc/server-manager/go.mod rpc/server-manager/go.sum ./rpc/server-manager/
COPY rpc/vpn-core/go.mod rpc/vpn-core/go.sum ./rpc/vpn-core/
COPY shared ./shared

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o analytics rpc/analytics/cmd/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc

# Copy the binary from builder stage
COPY --from=builder /app/analytics /app/analytics

# Copy configuration files
COPY rpc/analytics/etc/config.env /app/etc/config.env
COPY rpc/analytics/etc/analytics.yaml /app/etc/analytics.yaml

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/analytics"]
