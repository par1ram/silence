# Simple Dockerfile for pre-built dpi-bypass service
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create directories
RUN mkdir -p /app/etc

# Copy the pre-built binary
COPY rpc/dpi-bypass/bin/dpi-bypass /app/dpi-bypass

# Copy configuration files if they exist
COPY rpc/dpi-bypass/etc/config.env /app/etc/config.env

# Make binary executable
RUN chmod +x /app/dpi-bypass

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/dpi-bypass"]
