# Simple Dockerfile for pre-built server-manager service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc /app/migrations

# Copy the pre-built binary
COPY rpc/server-manager/bin/server-manager /app/server-manager

# Copy configuration files if they exist
COPY rpc/server-manager/etc/config.env /app/etc/config.env

# Copy migrations if they exist
COPY rpc/server-manager/internal/adapters/database/migrations /app/migrations

# Make binary executable
RUN chmod +x /app/server-manager

# Expose port
EXPOSE 8085

# Run the application
CMD ["/app/server-manager"]
