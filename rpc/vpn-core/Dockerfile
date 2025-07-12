# Simple Dockerfile for pre-built vpn-core service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc /app/migrations

# Copy the pre-built binary
COPY rpc/vpn-core/bin/vpn-core /app/vpn-core

# Copy configuration files if they exist
COPY rpc/vpn-core/etc/config.env /app/etc/config.env

# Copy migrations if they exist
COPY rpc/vpn-core/internal/adapters/database/migrations /app/migrations

# Make binary executable
RUN chmod +x /app/vpn-core

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/vpn-core"]
