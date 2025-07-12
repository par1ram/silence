# Simple Dockerfile for pre-built analytics service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc

# Copy the pre-built binary
COPY rpc/analytics/bin/analytics /app/analytics

# Copy configuration files if they exist
COPY rpc/analytics/etc/config.env /app/etc/config.env

# Make binary executable
RUN chmod +x /app/analytics

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/analytics"]
