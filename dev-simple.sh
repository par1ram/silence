#!/bin/bash

# Simple development script for Silence project
# Starts infrastructure services in Docker and application services locally

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_deps() {
    print_info "Checking dependencies..."

    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose not found"
        exit 1
    fi

    if ! command -v air &> /dev/null; then
        print_error "air not found. Install with: go install github.com/cosmtrek/air@latest"
        exit 1
    fi

    print_info "Dependencies OK"
}

# Start infrastructure
start_infra() {
    print_info "Starting infrastructure services..."

    # Start only infrastructure services
    docker-compose up -d postgres redis rabbitmq influxdb

    # Wait for services to be ready
    print_info "Waiting for services to be ready..."
    sleep 5

    # Simple health check
    local retries=30
    while [ $retries -gt 0 ]; do
        if docker-compose ps | grep -q "healthy\|Up"; then
            print_info "Infrastructure services are ready"
            return 0
        fi
        print_info "Waiting for services... ($retries retries left)"
        sleep 2
        retries=$((retries - 1))
    done

    print_warn "Services may still be starting up"
}

# Set local environment
set_local_env() {
    print_info "Setting local environment variables..."

    # Override Docker settings for local development
    export DB_HOST=localhost
    export REDIS_HOST=localhost
    export INFLUXDB_URL=http://localhost:8086
    export NOTIFICATIONS_RABBITMQ_URL=amqp://admin:admin@localhost:5672/
    export AUTH_MIGRATIONS_DIR=./internal/adapters/database/migrations
    export SERVER_MANAGER_MIGRATIONS_DIR=./internal/adapters/database/migrations

    # Local service ports
    export AUTH_PORT=9999
    export GATEWAY_PORT=8000
    export ANALYTICS_PORT=8001
    export SERVER_MANAGER_PORT=8002
    export DPI_BYPASS_PORT=8003
    export VPN_CORE_PORT=8004
    export NOTIFICATIONS_PORT=8005

    # Service URLs for local development
    export GATEWAY_AUTH_SERVICE_URL=http://localhost:9999
    export GATEWAY_ANALYTICS_SERVICE_URL=http://localhost:8001
    export GATEWAY_SERVER_MANAGER_SERVICE_URL=http://localhost:8002
    export GATEWAY_DPI_BYPASS_SERVICE_URL=http://localhost:8003
    export GATEWAY_VPN_CORE_SERVICE_URL=http://localhost:8004
    export NOTIFICATIONS_ANALYTICS_URL=http://localhost:8001
}

# Start single service
start_service() {
    local service_name=$1
    local service_dir=$2
    local port=$3

    print_info "Starting $service_name on port $port..."

    cd "$service_dir"

    # Set HTTP_PORT for this service
    export HTTP_PORT=$port

    # Start with air in background
    air > "../dev-$service_name.log" 2>&1 &
    local pid=$!

    cd - > /dev/null

    echo $pid >> .dev-pids

    # Quick check if service started
    sleep 2
    if ! kill -0 $pid 2>/dev/null; then
        print_error "$service_name failed to start"
        return 1
    fi

    print_info "$service_name started (PID: $pid)"
}

# Start all services
start_services() {
    print_info "Starting application services..."

    # Remove old PID file
    rm -f .dev-pids

    # Start services in order
    start_service "auth" "api/auth" $AUTH_PORT
    sleep 2

    start_service "analytics" "rpc/analytics" $ANALYTICS_PORT
    sleep 2

    start_service "server-manager" "rpc/server-manager" $SERVER_MANAGER_PORT
    sleep 2

    start_service "dpi-bypass" "rpc/dpi-bypass" $DPI_BYPASS_PORT
    sleep 2

    start_service "vpn-core" "rpc/vpn-core" $VPN_CORE_PORT
    sleep 2

    start_service "notifications" "rpc/notifications" $NOTIFICATIONS_PORT
    sleep 2

    # Gateway last (depends on others)
    start_service "gateway" "api/gateway" $GATEWAY_PORT

    print_info "All services started"
}

# Show status
show_status() {
    echo ""
    echo "=== Silence Development Environment ==="
    echo ""
    echo "Infrastructure (Docker):"
    echo "  PostgreSQL: localhost:5432"
    echo "  Redis: localhost:6379"
    echo "  RabbitMQ: localhost:5672 (Management: localhost:15672)"
    echo "  InfluxDB: localhost:8086"
    echo ""
    echo "Application Services (Local):"
    echo "  Gateway: http://localhost:$GATEWAY_PORT"
    echo "  Auth: http://localhost:$AUTH_PORT"
    echo "  Analytics: http://localhost:$ANALYTICS_PORT"
    echo "  Server Manager: http://localhost:$SERVER_MANAGER_PORT"
    echo "  DPI Bypass: http://localhost:$DPI_BYPASS_PORT"
    echo "  VPN Core: http://localhost:$VPN_CORE_PORT"
    echo "  Notifications: http://localhost:$NOTIFICATIONS_PORT"
    echo ""
    echo "Logs: dev-*.log files"
    echo "Stop: Ctrl+C or run ./dev-simple.sh stop"
}

# Stop all services
stop_services() {
    print_info "Stopping services..."

    # Stop application services
    if [ -f .dev-pids ]; then
        while read -r pid; do
            if kill -0 $pid 2>/dev/null; then
                print_info "Stopping service (PID: $pid)"
                kill $pid
            fi
        done < .dev-pids
        rm -f .dev-pids
    fi

    # Stop infrastructure
    docker-compose down

    print_info "All services stopped"
}

# Cleanup on exit
cleanup() {
    print_info "Cleaning up..."
    stop_services
}

# Main function
main() {
    case "${1:-start}" in
        "start")
            trap cleanup EXIT
            check_deps
            set_local_env
            start_infra
            start_services
            show_status

            print_info "Press Ctrl+C to stop all services"
            # Wait for interrupt
            while true; do
                sleep 1
            done
            ;;
        "stop")
            stop_services
            ;;
        "status")
            show_status
            ;;
        *)
            echo "Usage: $0 [start|stop|status]"
            exit 1
            ;;
    esac
}

main "$@"
