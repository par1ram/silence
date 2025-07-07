#!/bin/bash

# Dev startup script for Silence project
# This script starts all services in development mode with hot reload

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Check if required tools are installed
check_dependencies() {
    print_header "Checking Dependencies"

    # Check if air is installed
    if ! command -v air &> /dev/null; then
        print_error "air (hot reload tool) is not installed"
        echo "Install it with: go install github.com/cosmtrek/air@latest"
        exit 1
    fi

    # Check if task is installed
    if ! command -v task &> /dev/null; then
        print_error "task runner is not installed"
        echo "Install it with: go install github.com/go-task/task/v3/cmd/task@latest"
        exit 1
    fi

    # Check if docker-compose is available
    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose is not installed"
        exit 1
    fi

    print_status "All dependencies are available"
}

# Load environment variables
load_env() {
    print_header "Loading Environment Variables"

    if [ -f .env.local ]; then
        print_status "Loading .env.local"
        export $(cat .env.local | grep -v '^#' | xargs)
    else
        print_warning ".env.local not found, using defaults"
    fi

    # Set default migration paths
    export AUTH_MIGRATIONS_DIR="./internal/adapters/database/migrations"
    export SERVER_MANAGER_MIGRATIONS_DIR="./internal/adapters/database/migrations"

    print_status "Environment variables loaded"
}

# Start infrastructure services
start_infrastructure() {
    print_header "Starting Infrastructure Services"

    print_status "Starting PostgreSQL, Redis, RabbitMQ, and InfluxDB..."
    docker-compose up -d postgres redis rabbitmq influxdb

    print_status "Waiting for services to be healthy..."
    sleep 10

    # Wait for PostgreSQL
    until docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; do
        print_status "Waiting for PostgreSQL..."
        sleep 2
    done

    # Wait for Redis
    until docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
        print_status "Waiting for Redis..."
        sleep 2
    done

    print_status "Infrastructure services are ready"
}

# Clean up function
cleanup() {
    print_header "Cleaning Up"

    # Kill background processes
    if [ ! -z "$PIDS" ]; then
        print_status "Stopping all services..."
        for pid in $PIDS; do
            kill $pid 2>/dev/null || true
        done
    fi

    # Wait a bit for graceful shutdown
    sleep 2

    print_status "Cleanup completed"
}

# Trap cleanup on exit
trap cleanup EXIT

# Start a single service in background
start_service() {
    local service=$1
    local dir=$2
    local port=$3

    print_status "Starting $service on port $port..."

    cd $dir
    air > ../dev-$service.log 2>&1 &
    local pid=$!
    cd - > /dev/null

    PIDS="$PIDS $pid"

    # Wait a bit and check if service started
    sleep 3
    if ! kill -0 $pid 2>/dev/null; then
        print_error "$service failed to start"
        return 1
    fi

    print_status "$service started (PID: $pid)"
    return 0
}

# Wait for service to be ready
wait_for_service() {
    local service=$1
    local url=$2
    local max_attempts=30
    local attempt=0

    print_status "Waiting for $service to be ready..."

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|404"; then
            print_status "$service is ready"
            return 0
        fi

        attempt=$((attempt + 1))
        sleep 2
    done

    print_warning "$service may not be ready (timeout)"
    return 1
}

# Start all application services
start_services() {
    print_header "Starting Application Services"

    PIDS=""

    # Start Auth service
    start_service "Auth" "api/auth" "9999"
    wait_for_service "Auth" "http://localhost:9999/health"

    # Start Analytics service
    start_service "Analytics" "rpc/analytics" "8001"
    wait_for_service "Analytics" "http://localhost:8001/health"

    # Start Server Manager service
    start_service "Server Manager" "rpc/server-manager" "8002"
    wait_for_service "Server Manager" "http://localhost:8002/health"

    # Start DPI Bypass service
    start_service "DPI Bypass" "rpc/dpi-bypass" "8003"
    wait_for_service "DPI Bypass" "http://localhost:8003/health"

    # Start VPN Core service
    start_service "VPN Core" "rpc/vpn-core" "8004"
    wait_for_service "VPN Core" "http://localhost:8004/health"

    # Start Notifications service
    start_service "Notifications" "rpc/notifications" "8005"
    wait_for_service "Notifications" "http://localhost:8005/healthz"

    # Start Gateway service last (depends on other services)
    start_service "Gateway" "api/gateway" "8000"
    wait_for_service "Gateway" "http://localhost:8000/health"

    print_status "All services started successfully"
}

# Show running services
show_services() {
    print_header "Running Services"

    echo "Infrastructure:"
    echo "  PostgreSQL: localhost:5432"
    echo "  Redis: localhost:6379"
    echo "  RabbitMQ: localhost:5672 (Management: localhost:15672)"
    echo "  InfluxDB: localhost:8086"
    echo ""
    echo "Application Services:"
    echo "  Gateway: http://localhost:8000"
    echo "  Auth: http://localhost:9999"
    echo "  Analytics: http://localhost:8001"
    echo "  Server Manager: http://localhost:8002"
    echo "  DPI Bypass: http://localhost:8003"
    echo "  VPN Core: http://localhost:8004"
    echo "  Notifications: http://localhost:8005"
    echo ""
    echo "Logs are available in:"
    echo "  dev-auth.log, dev-analytics.log, dev-server-manager.log, etc."
}

# Main function
main() {
    print_header "Silence Development Environment"

    check_dependencies
    load_env
    start_infrastructure
    start_services
    show_services

    print_header "All Services Running"
    print_status "Press Ctrl+C to stop all services"

    # Wait for interrupt
    while true; do
        sleep 1
    done
}

# Run main function
main
