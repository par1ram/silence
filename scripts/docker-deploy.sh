#!/bin/bash

# Docker Deployment Script for Silence Project
# This script handles building and deploying all services using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="silence"
COMPOSE_FILE="docker-compose.yml"
LOG_FILE="deploy.log"

# Functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    success "Docker is running"
}

# Check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose > /dev/null 2>&1; then
        error "Docker Compose is not installed. Please install Docker Compose and try again."
        exit 1
    fi
    success "Docker Compose is available"
}

# Clean up previous containers and images
cleanup() {
    log "Cleaning up previous containers and images..."

    # Stop and remove containers
    if docker-compose -f "$COMPOSE_FILE" ps -q > /dev/null 2>&1; then
        docker-compose -f "$COMPOSE_FILE" down --remove-orphans
    fi

    # Remove dangling images
    if docker images -f "dangling=true" -q | grep -q .; then
        docker rmi $(docker images -f "dangling=true" -q) 2>/dev/null || true
    fi

    success "Cleanup completed"
}

# Build services
build_services() {
    log "Building Docker images..."

    # Build with no cache to ensure fresh build
    docker-compose -f "$COMPOSE_FILE" build --no-cache --parallel

    success "All services built successfully"
}

# Start infrastructure services first
start_infrastructure() {
    log "Starting infrastructure services..."

    # Start databases and message queues first
    docker-compose -f "$COMPOSE_FILE" up -d postgres redis rabbitmq influxdb

    # Wait for health checks
    log "Waiting for infrastructure services to be healthy..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep -E "(postgres|redis|rabbitmq|influxdb)" | grep -q "healthy"; then
            success "Infrastructure services are healthy"
            return 0
        fi

        log "Attempt $attempt/$max_attempts: Waiting for services to be healthy..."
        sleep 10
        ((attempt++))
    done

    error "Infrastructure services failed to become healthy within timeout"
    docker-compose -f "$COMPOSE_FILE" logs
    exit 1
}

# Start application services
start_applications() {
    log "Starting application services..."

    # Start auth service first
    docker-compose -f "$COMPOSE_FILE" up -d auth
    sleep 15

    # Start other services
    docker-compose -f "$COMPOSE_FILE" up -d analytics server-manager dpi-bypass vpn-core notifications
    sleep 10

    # Start gateway last
    docker-compose -f "$COMPOSE_FILE" up -d gateway

    success "All application services started"
}

# Check service health
check_services() {
    log "Checking service health..."

    local services=("auth" "gateway" "analytics" "server-manager" "dpi-bypass" "vpn-core" "notifications")
    local failed_services=()

    for service in "${services[@]}"; do
        if ! docker-compose -f "$COMPOSE_FILE" ps "$service" | grep -q "Up"; then
            failed_services+=("$service")
        fi
    done

    if [ ${#failed_services[@]} -eq 0 ]; then
        success "All services are running"
        return 0
    else
        error "Failed services: ${failed_services[*]}"
        return 1
    fi
}

# Show service status
show_status() {
    log "Service Status:"
    docker-compose -f "$COMPOSE_FILE" ps

    echo
    log "Service URLs:"
    echo "  Gateway:        http://localhost:8080"
    echo "  Auth:           http://localhost:8081"
    echo "  Analytics:      http://localhost:8082"
    echo "  DPI Bypass:     http://localhost:8083"
    echo "  VPN Core:       http://localhost:8084"
    echo "  Server Manager: http://localhost:8085"
    echo "  Notifications:  http://localhost:8087"
    echo
    echo "  PostgreSQL:     localhost:5432"
    echo "  Redis:          localhost:6379"
    echo "  RabbitMQ:       http://localhost:15672 (admin/admin)"
    echo "  InfluxDB:       http://localhost:8086"
}

# Show logs
show_logs() {
    if [ -n "$1" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f "$1"
    else
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

# Main deployment function
deploy() {
    log "Starting deployment of $PROJECT_NAME..."

    check_docker
    check_docker_compose

    if [ "$1" = "--clean" ]; then
        cleanup
    fi

    build_services
    start_infrastructure
    start_applications

    # Wait a bit for services to fully start
    sleep 20

    if check_services; then
        success "Deployment completed successfully!"
        show_status
    else
        error "Deployment failed!"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi
}

# Help function
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo
    echo "Commands:"
    echo "  deploy [--clean]    Deploy all services (use --clean to cleanup first)"
    echo "  start              Start all services"
    echo "  stop               Stop all services"
    echo "  restart            Restart all services"
    echo "  status             Show service status"
    echo "  logs [SERVICE]     Show logs (optionally for specific service)"
    echo "  cleanup            Clean up containers and images"
    echo "  build              Build all services"
    echo "  help               Show this help message"
    echo
    echo "Examples:"
    echo "  $0 deploy --clean"
    echo "  $0 logs gateway"
    echo "  $0 status"
}

# Main script logic
case "$1" in
    "deploy")
        deploy "$2"
        ;;
    "start")
        log "Starting all services..."
        docker-compose -f "$COMPOSE_FILE" up -d
        show_status
        ;;
    "stop")
        log "Stopping all services..."
        docker-compose -f "$COMPOSE_FILE" down
        success "All services stopped"
        ;;
    "restart")
        log "Restarting all services..."
        docker-compose -f "$COMPOSE_FILE" restart
        show_status
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs "$2"
        ;;
    "cleanup")
        cleanup
        ;;
    "build")
        build_services
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        if [ -z "$1" ]; then
            deploy
        else
            error "Unknown command: $1"
            show_help
            exit 1
        fi
        ;;
esac
