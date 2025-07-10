#!/bin/bash

# Silence VPN Observability Stack Management Script
# This script helps manage the OpenTelemetry observability stack

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="silence-observability"
SERVICES=(
    "redis"
    "otel-collector"
    "prometheus"
    "grafana"
    "jaeger"
    "zipkin"
    "loki"
    "promtail"
    "tempo"
    "alertmanager"
    "node-exporter"
    "cadvisor"
)

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if docker and docker-compose are installed
check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    log_success "Dependencies check passed"
}

# Create necessary directories
create_directories() {
    log_info "Creating necessary directories..."

    # Create log directories
    mkdir -p logs/{analytics,gateway,vpn-core,dpi-bypass,server-manager,notifications}
    chmod 755 logs
    chmod 755 logs/*

    # Create data directories
    mkdir -p data/{prometheus,grafana,loki,tempo,jaeger,redis,alertmanager}

    log_success "Directories created"
}

# Validate configuration files
validate_config() {
    log_info "Validating configuration files..."

    local config_files=(
        "otel-collector-config.yaml"
        "prometheus.yml"
        "loki-config.yaml"
        "tempo-config.yaml"
        "alertmanager.yml"
        "promtail-config.yaml"
        "alert_rules.yml"
        "recording_rules.yml"
    )

    for file in "${config_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            log_error "Configuration file $file is missing"
            exit 1
        fi
    done

    log_success "Configuration validation passed"
}

# Start the observability stack
start_stack() {
    log_info "Starting observability stack..."

    check_dependencies
    create_directories
    validate_config

    # Start services in order
    log_info "Starting core services..."
    docker-compose -p $PROJECT_NAME up -d redis
    sleep 5

    log_info "Starting observability services..."
    docker-compose -p $PROJECT_NAME up -d prometheus grafana loki tempo jaeger zipkin
    sleep 10

    log_info "Starting collectors and processors..."
    docker-compose -p $PROJECT_NAME up -d otel-collector promtail alertmanager
    sleep 5

    log_info "Starting monitoring services..."
    docker-compose -p $PROJECT_NAME up -d node-exporter cadvisor
    sleep 5

    log_info "Starting health check service..."
    docker-compose -p $PROJECT_NAME up -d health-check

    log_success "Observability stack started successfully"
    show_status
}

# Stop the observability stack
stop_stack() {
    log_info "Stopping observability stack..."
    docker-compose -p $PROJECT_NAME down
    log_success "Observability stack stopped"
}

# Restart the observability stack
restart_stack() {
    log_info "Restarting observability stack..."
    stop_stack
    sleep 3
    start_stack
}

# Show status of all services
show_status() {
    log_info "Checking service status..."

    echo ""
    echo "=== Service Status ==="
    docker-compose -p $PROJECT_NAME ps

    echo ""
    echo "=== Service URLs ==="
    echo "Grafana Dashboard:    http://localhost:3000 (admin/admin)"
    echo "Prometheus:           http://localhost:9090"
    echo "Jaeger UI:            http://localhost:16686"
    echo "Zipkin UI:            http://localhost:9411"
    echo "AlertManager:         http://localhost:9093"
    echo "Tempo:                http://localhost:3200"
    echo "Loki:                 http://localhost:3100"
    echo "Node Exporter:        http://localhost:9100"
    echo "cAdvisor:             http://localhost:8080"
    echo "OTel Collector:       http://localhost:8888"
    echo ""
}

# Show logs for a specific service
show_logs() {
    local service=$1
    if [[ -z "$service" ]]; then
        log_error "Please specify a service name"
        echo "Available services: ${SERVICES[*]}"
        exit 1
    fi

    log_info "Showing logs for $service..."
    docker-compose -p $PROJECT_NAME logs -f "$service"
}

# Health check for all services
health_check() {
    log_info "Performing health check..."

    local failed_services=()

    # Check Prometheus
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/-/healthy | grep -q "200"; then
        failed_services+=("prometheus")
    fi

    # Check Grafana
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/health | grep -q "200"; then
        failed_services+=("grafana")
    fi

    # Check Jaeger
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:16686/ | grep -q "200"; then
        failed_services+=("jaeger")
    fi

    # Check Loki
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:3100/ready | grep -q "200"; then
        failed_services+=("loki")
    fi

    # Check OpenTelemetry Collector
    if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:13133/ | grep -q "200"; then
        failed_services+=("otel-collector")
    fi

    if [[ ${#failed_services[@]} -eq 0 ]]; then
        log_success "All services are healthy"
    else
        log_error "Failed services: ${failed_services[*]}"
        exit 1
    fi
}

# Clean up everything (remove containers, volumes, networks)
cleanup() {
    log_warning "This will remove all containers, volumes, and networks related to observability stack"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up observability stack..."
        docker-compose -p $PROJECT_NAME down -v --remove-orphans
        docker system prune -f
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Update stack (pull latest images and restart)
update_stack() {
    log_info "Updating observability stack..."
    docker-compose -p $PROJECT_NAME pull
    restart_stack
    log_success "Stack updated successfully"
}

# Backup configuration and data
backup() {
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    log_info "Creating backup in $backup_dir..."

    mkdir -p "$backup_dir"

    # Backup configuration files
    cp *.yml *.yaml "$backup_dir/"

    # Backup Grafana data
    docker run --rm -v silence-observability_grafana_data:/data -v "$(pwd)/$backup_dir":/backup alpine tar czf /backup/grafana_data.tar.gz -C /data .

    # Backup Prometheus data
    docker run --rm -v silence-observability_prometheus_data:/data -v "$(pwd)/$backup_dir":/backup alpine tar czf /backup/prometheus_data.tar.gz -C /data .

    log_success "Backup created in $backup_dir"
}

# Restore from backup
restore() {
    local backup_dir=$1
    if [[ -z "$backup_dir" ]]; then
        log_error "Please specify backup directory"
        exit 1
    fi

    if [[ ! -d "$backup_dir" ]]; then
        log_error "Backup directory $backup_dir does not exist"
        exit 1
    fi

    log_warning "This will restore configuration and data from $backup_dir"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Restoring from backup..."

        # Stop services
        stop_stack

        # Restore configuration files
        cp "$backup_dir"/*.yml "$backup_dir"/*.yaml . 2>/dev/null || true

        # Restore Grafana data
        if [[ -f "$backup_dir/grafana_data.tar.gz" ]]; then
            docker run --rm -v silence-observability_grafana_data:/data -v "$(pwd)/$backup_dir":/backup alpine tar xzf /backup/grafana_data.tar.gz -C /data
        fi

        # Restore Prometheus data
        if [[ -f "$backup_dir/prometheus_data.tar.gz" ]]; then
            docker run --rm -v silence-observability_prometheus_data:/data -v "$(pwd)/$backup_dir":/backup alpine tar xzf /backup/prometheus_data.tar.gz -C /data
        fi

        # Start services
        start_stack

        log_success "Restore completed"
    else
        log_info "Restore cancelled"
    fi
}

# Show help
show_help() {
    cat << EOF
Silence VPN Observability Stack Management Script

Usage: $0 [COMMAND]

Commands:
  start         Start the observability stack
  stop          Stop the observability stack
  restart       Restart the observability stack
  status        Show status of all services
  logs [SERVICE] Show logs for a specific service
  health        Perform health check on all services
  cleanup       Remove all containers, volumes, and networks
  update        Update all images and restart stack
  backup        Create a backup of configuration and data
  restore [DIR] Restore from backup directory
  help          Show this help message

Examples:
  $0 start                    # Start the stack
  $0 logs prometheus          # Show Prometheus logs
  $0 health                   # Check health of all services
  $0 backup                   # Create backup
  $0 restore backups/20240115_120000  # Restore from backup

Available services for logs:
  ${SERVICES[*]}

For more information, visit: https://docs.silence-vpn.com/observability
EOF
}

# Main script logic
main() {
    case "${1:-}" in
        start)
            start_stack
            ;;
        stop)
            stop_stack
            ;;
        restart)
            restart_stack
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        health)
            health_check
            ;;
        cleanup)
            cleanup
            ;;
        update)
            update_stack
            ;;
        backup)
            backup
            ;;
        restore)
            restore "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
