#!/bin/bash

# Simple test script for single service
# Usage: ./test-service.sh [service_name]

set -e

SERVICE=${1:-auth}
PORT=${2:-9999}

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Set local environment
export DB_HOST=localhost
export REDIS_HOST=localhost
export INFLUXDB_URL=http://localhost:8086
export NOTIFICATIONS_RABBITMQ_URL=amqp://admin:admin@localhost:5672/
export AUTH_MIGRATIONS_DIR=./internal/adapters/database/migrations
export SERVER_MANAGER_MIGRATIONS_DIR=./internal/adapters/database/migrations
export MIGRATIONS_DIR=./internal/adapters/database/migrations
export HTTP_PORT=$PORT

# Start infrastructure if not running
if ! docker-compose ps | grep -q "Up\|healthy"; then
    print_info "Starting infrastructure..."
    docker-compose up -d postgres redis rabbitmq influxdb
    sleep 10
fi

# Navigate to service directory
case $SERVICE in
    "auth")
        cd api/auth
        ;;
    "gateway")
        cd api/gateway
        ;;
    "analytics")
        cd rpc/analytics
        ;;
    "server-manager")
        cd rpc/server-manager
        ;;
    "dpi-bypass")
        cd rpc/dpi-bypass
        ;;
    "vpn-core")
        cd rpc/vpn-core
        ;;
    "notifications")
        cd rpc/notifications
        ;;
    *)
        print_error "Unknown service: $SERVICE"
        exit 1
        ;;
esac

print_info "Starting $SERVICE on port $PORT..."
print_info "Directory: $(pwd)"
print_info "Environment: HTTP_PORT=$HTTP_PORT"

# Start service
air
