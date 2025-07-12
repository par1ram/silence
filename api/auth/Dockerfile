# Simple Dockerfile for pre-built auth service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc /app/migrations

# Copy the pre-built binary
COPY api/auth/bin/auth /app/auth

# Copy configuration files if they exist
COPY api/auth/etc/config.env /app/etc/config.env

# Copy migrations if they exist
COPY api/auth/internal/adapters/database/migrations /app/migrations

# Make binary executable
RUN chmod +x /app/auth

# Expose port
EXPOSE 8080 9080

# Run the application
CMD ["/app/auth"]
