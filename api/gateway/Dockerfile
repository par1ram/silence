# Simple Dockerfile for pre-built gateway service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc

# Copy the pre-built binary
COPY api/gateway/bin/gateway /app/gateway

# Copy configuration files if they exist
COPY api/gateway/etc/config.env /app/etc/config.env

# Make binary executable
RUN chmod +x /app/gateway

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/gateway"]
