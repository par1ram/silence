#!/bin/bash

# Health Check Script for Silence Project
# This script checks the health of all services and provides detailed status information

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
TIMEOUT=30
MAX_RETRIES=5

# Service endpoints
declare -A SERVICE_ENDPOINTS=(
    ["gateway"]="http://localhost:8080/health"
    ["auth"]="http://localhost:8081/health"
    ["analytics"]="http://localhost:8082/health"
    ["dpi-bypass"]="http://localhost:8083/health"
    ["vpn-core"]="http://localhost:8084/health"
    ["server-manager"]="http://localhost:8085/health"
    ["notifications"]="http://localhost:8087/health"
)

# Infrastructure services
declare -A INFRA_SERVICES=(
    ["postgres"]="5432"
    ["redis"]="6379"
    ["rabbitmq"]="5672"
    ["influxdb"]="8086"
)

# Functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        error "Docker is not running"
        return 1
    fi
    return 0
}

# Check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose > /dev/null 2>&1; then
        error "Docker Compose is not installed"
        return 1
    fi
    return 0
}

# Check container status
check_container_status() {
    local service=$1
    local container_name="silence_${service}"

    if [ "$service" = "server-manager" ]; then
        container_name="silence_server_manager"
    fi

    local status=$(docker inspect --format='{{.State.Status}}' "$container_name" 2>/dev/null || echo "not_found")

    case $status in
        "running")
            success "$service container is running"
            return 0
            ;;
        "exited")
            error "$service container has exited"
            return 1
            ;;
        "not_found")
            error "$service container not found"
            return 1
            ;;
        *)
            warning "$service container status: $status"
            return 1
            ;;
    esac
}

# Check container health
check_container_health() {
    local service=$1
    local container_name="silence_${service}"

    if [ "$service" = "server-manager" ]; then
        container_name="silence_server_manager"
    fi

    local health=$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}no-health-check{{end}}' "$container_name" 2>/dev/null || echo "not_found")

    case $health in
        "healthy")
            success "$service container is healthy"
            return 0
            ;;
        "unhealthy")
            error "$service container is unhealthy"
            return 1
            ;;
        "starting")
            warning "$service container health check is starting"
            return 1
            ;;
        "no-health-check")
            warning "$service container has no health check configured"
            return 0
            ;;
        "not_found")
            error "$service container not found"
            return 1
            ;;
        *)
            warning "$service container health: $health"
            return 1
            ;;
    esac
}

# Check HTTP endpoint
check_http_endpoint() {
    local service=$1
    local endpoint=$2
    local retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -s -f --max-time $TIMEOUT "$endpoint" > /dev/null 2>&1; then
            success "$service HTTP endpoint is responding"
            return 0
        fi

        retries=$((retries + 1))
        if [ $retries -lt $MAX_RETRIES ]; then
            log "Retrying $service endpoint ($retries/$MAX_RETRIES)..."
            sleep 2
        fi
    done

    error "$service HTTP endpoint is not responding"
    return 1
}

# Check TCP port
check_tcp_port() {
    local service=$1
    local port=$2
    local host="localhost"

    if nc -z -w$TIMEOUT "$host" "$port" > /dev/null 2>&1; then
        success "$service TCP port $port is open"
        return 0
    else
        error "$service TCP port $port is not accessible"
        return 1
    fi
}

# Check database connection
check_database() {
    local db_name=$1
    local retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        if docker-compose exec -T postgres psql -U postgres -d "$db_name" -c "SELECT 1;" > /dev/null 2>&1; then
            success "Database $db_name is accessible"
            return 0
        fi

        retries=$((retries + 1))
        if [ $retries -lt $MAX_RETRIES ]; then
            log "Retrying database connection ($retries/$MAX_RETRIES)..."
            sleep 2
        fi
    done

    error "Database $db_name is not accessible"
    return 1
}

# Check Redis connection
check_redis() {
    if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
        success "Redis is responding"
        return 0
    else
        error "Redis is not responding"
        return 1
    fi
}

# Check RabbitMQ
check_rabbitmq() {
    if docker-compose exec -T rabbitmq rabbitmq-diagnostics ping > /dev/null 2>&1; then
        success "RabbitMQ is responding"
        return 0
    else
        error "RabbitMQ is not responding"
        return 1
    fi
}

# Check InfluxDB
check_influxdb() {
    if curl -s -f --max-time $TIMEOUT "http://localhost:8086/ping" > /dev/null 2>&1; then
        success "InfluxDB is responding"
        return 0
    else
        error "InfluxDB is not responding"
        return 1
    fi
}

# Get container resource usage
get_container_stats() {
    local service=$1
    local container_name="silence_${service}"

    if [ "$service" = "server-manager" ]; then
        container_name="silence_server_manager"
    fi

    local stats=$(docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" "$container_name" 2>/dev/null || echo "N/A")

    if [ "$stats" != "N/A" ]; then
        echo "$stats"
    else
        echo "$container_name: Stats not available"
    fi
}

# Show service logs
show_service_logs() {
    local service=$1
    local lines=${2:-10}

    echo "=== Last $lines lines of $service logs ==="
    docker-compose logs --tail=$lines "$service" 2>/dev/null || echo "No logs available for $service"
    echo
}

# Main health check function
run_health_check() {
    local failed_services=()

    log "Starting health check for Silence project..."
    echo

    # Check prerequisites
    log "Checking prerequisites..."
    if ! check_docker; then
        exit 1
    fi

    if ! check_docker_compose; then
        exit 1
    fi

    echo

    # Check infrastructure services
    log "Checking infrastructure services..."

    for service in "${!INFRA_SERVICES[@]}"; do
        local port=${INFRA_SERVICES[$service]}

        if ! check_container_status "$service"; then
            failed_services+=("$service")
            continue
        fi

        if ! check_container_health "$service"; then
            failed_services+=("$service")
            continue
        fi

        if ! check_tcp_port "$service" "$port"; then
            failed_services+=("$service")
            continue
        fi
    done

    # Additional infrastructure checks
    if check_container_status "postgres"; then
        check_database "silence_auth"
        check_database "silence_server_manager"
        check_database "silence_vpn"
    fi

    if check_container_status "redis"; then
        check_redis
    fi

    if check_container_status "rabbitmq"; then
        check_rabbitmq
    fi

    if check_container_status "influxdb"; then
        check_influxdb
    fi

    echo

    # Check application services
    log "Checking application services..."

    for service in "${!SERVICE_ENDPOINTS[@]}"; do
        local endpoint=${SERVICE_ENDPOINTS[$service]}

        if ! check_container_status "$service"; then
            failed_services+=("$service")
            continue
        fi

        if ! check_container_health "$service"; then
            failed_services+=("$service")
            continue
        fi

        if ! check_http_endpoint "$service" "$endpoint"; then
            failed_services+=("$service")
            continue
        fi
    done

    echo

    # Show summary
    log "Health check summary:"

    if [ ${#failed_services[@]} -eq 0 ]; then
        success "All services are healthy!"
        return 0
    else
        error "Failed services: ${failed_services[*]}"

        echo
        log "Showing logs for failed services:"
        for service in "${failed_services[@]}"; do
            show_service_logs "$service" 20
        done

        return 1
    fi
}

# Show detailed status
show_detailed_status() {
    log "Detailed service status:"
    echo

    # Docker Compose status
    docker-compose ps
    echo

    # Resource usage
    log "Resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
    echo

    # Disk usage
    log "Volume usage:"
    docker volume ls -q | grep -E "^silence_" | xargs -I {} docker volume inspect {} --format "{{.Name}}: {{.Mountpoint}}" 2>/dev/null || echo "No volumes found"
    echo

    # Network information
    log "Network information:"
    docker network ls | grep silence
    echo
}

# Show service URLs
show_service_urls() {
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
    echo
}

# Help function
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  check           Run full health check (default)"
    echo "  status          Show detailed service status"
    echo "  urls            Show service URLs"
    echo "  logs [SERVICE]  Show logs for specific service"
    echo "  help            Show this help message"
    echo
    echo "Examples:"
    echo "  $0 check"
    echo "  $0 status"
    echo "  $0 logs gateway"
    echo "  $0 urls"
}

# Main script logic
case "${1:-check}" in
    "check")
        run_health_check
        ;;
    "status")
        show_detailed_status
        ;;
    "urls")
        show_service_urls
        ;;
    "logs")
        if [ -n "$2" ]; then
            show_service_logs "$2" 50
        else
            docker-compose logs --tail=20 -f
        fi
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
