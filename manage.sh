#!/bin/bash

# Silence Project Management Script
# This is the main entry point for managing the Silence VPN project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Project configuration
PROJECT_NAME="Silence VPN"
PROJECT_VERSION="1.0.0"
COMPOSE_FILE="docker-compose.yml"

# ASCII Art Logo
show_logo() {
    echo -e "${CYAN}"
    cat << 'EOF'
 ____  _ _
/ ___|(_) | ___ _ __   ___ ___
\___ \| | |/ _ \ '_ \ / __/ _ \
 ___) | | |  __/ | | | (_|  __/
|____/|_|_|\___|_| |_|\___\___|

    VPN Project Management
EOF
    echo -e "${NC}"
}

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
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

info() {
    echo -e "${CYAN}[i]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    local missing=0

    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed"
        missing=1
    elif ! docker info &> /dev/null; then
        error "Docker is not running"
        missing=1
    fi

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed"
        missing=1
    fi

    # Check required files
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "Docker Compose file not found: $COMPOSE_FILE"
        missing=1
    fi

    if [ ! -f ".env" ]; then
        warning ".env file not found. Using default values."
        if [ -f ".env.example" ]; then
            info "Copying .env.example to .env"
            cp .env.example .env
        fi
    fi

    if [ $missing -eq 1 ]; then
        error "Prerequisites not met. Please install missing dependencies."
        exit 1
    fi

    success "Prerequisites check passed"
}

# Development commands
dev_start() {
    log "Starting development environment..."

    # Start infrastructure first
    log "Starting infrastructure services..."
    docker-compose up -d postgres redis rabbitmq influxdb

    # Wait for health checks
    log "Waiting for infrastructure to be ready..."
    sleep 15

    # Start application services
    log "Starting application services..."
    docker-compose up -d auth
    sleep 10
    docker-compose up -d analytics server-manager dpi-bypass vpn-core notifications
    sleep 10
    docker-compose up -d gateway

    success "Development environment started"
    show_service_urls
}

dev_stop() {
    log "Stopping development environment..."
    docker-compose down
    success "Development environment stopped"
}

dev_restart() {
    log "Restarting development environment..."
    dev_stop
    sleep 5
    dev_start
}

dev_logs() {
    local service=$1
    if [ -n "$service" ]; then
        log "Showing logs for $service..."
        docker-compose logs -f "$service"
    else
        log "Showing logs for all services..."
        docker-compose logs -f
    fi
}

# Build commands
build_all() {
    log "Building all services..."
    docker-compose build --parallel
    success "All services built successfully"
}

build_service() {
    local service=$1
    if [ -z "$service" ]; then
        error "Service name required"
        return 1
    fi

    log "Building $service..."
    docker-compose build "$service"
    success "$service built successfully"
}

# Clean up commands
clean_containers() {
    log "Cleaning up containers..."
    docker-compose down --remove-orphans
    success "Containers cleaned up"
}

clean_images() {
    log "Cleaning up images..."
    docker image prune -a -f
    success "Images cleaned up"
}

clean_volumes() {
    log "Cleaning up volumes..."
    docker volume prune -f
    success "Volumes cleaned up"
}

clean_all() {
    log "Performing full cleanup..."
    clean_containers
    clean_images
    clean_volumes
    docker system prune -a -f
    success "Full cleanup completed"
}

# Status and monitoring
show_status() {
    log "Service status:"
    docker-compose ps
    echo

    log "Resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" 2>/dev/null || echo "No running containers"
}

show_service_urls() {
    echo
    info "Service URLs:"
    echo "  🌐 Gateway:        http://localhost:8080"
    echo "  🔐 Auth:           http://localhost:8081"
    echo "  📊 Analytics:      http://localhost:8082"
    echo "  🔒 DPI Bypass:     http://localhost:8083"
    echo "  🔑 VPN Core:       http://localhost:8084"
    echo "  ⚙️  Server Manager: http://localhost:8085"
    echo "  📢 Notifications:  http://localhost:8087"
    echo
    echo "  🗄️  PostgreSQL:     localhost:5432"
    echo "  🔄 Redis:          localhost:6379"
    echo "  🐰 RabbitMQ:       http://localhost:15672 (admin/admin)"
    echo "  📈 InfluxDB:       http://localhost:8086"
    echo
}

# Health check
health_check() {
    if [ -f "scripts/health-check.sh" ]; then
        log "Running health check..."
        ./scripts/health-check.sh check
    else
        warning "Health check script not found"
        show_status
    fi
}

# Database operations
db_migrate() {
    log "Running database migrations..."

    # Auth service migrations
    if docker-compose ps auth | grep -q "Up"; then
        log "Running auth migrations..."
        docker-compose exec auth /app/auth migrate || warning "Auth migrations failed"
    fi

    # Server manager migrations
    if docker-compose ps server-manager | grep -q "Up"; then
        log "Running server-manager migrations..."
        docker-compose exec server-manager /app/server-manager migrate || warning "Server-manager migrations failed"
    fi

    success "Database migrations completed"
}

db_reset() {
    warning "This will destroy all data!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Resetting database..."
        docker-compose down -v
        docker-compose up -d postgres
        sleep 10
        db_migrate
        success "Database reset completed"
    else
        info "Database reset cancelled"
    fi
}

# Testing
run_tests() {
    log "Running tests..."

    # Check if services are running
    if ! docker-compose ps | grep -q "Up"; then
        error "Services are not running. Start them first with: $0 start"
        exit 1
    fi

    # Run integration tests if available
    if [ -f "scripts/test-integration.sh" ]; then
        ./scripts/test-integration.sh
    else
        warning "Integration tests not found"
    fi

    success "Tests completed"
}

# Configuration
show_config() {
    log "Current configuration:"
    echo

    if [ -f ".env" ]; then
        echo "Environment variables (.env):"
        cat .env | grep -E "^[A-Z_]+" | head -20
        echo
    fi

    echo "Docker Compose services:"
    docker-compose config --services
    echo

    echo "Docker Compose version:"
    docker-compose version --short
}

edit_config() {
    local editor=${EDITOR:-nano}

    if [ -f ".env" ]; then
        log "Opening .env file with $editor..."
        $editor .env
    else
        warning ".env file not found. Creating from template..."
        if [ -f ".env.example" ]; then
            cp .env.example .env
            $editor .env
        else
            error "No .env.example file found"
            exit 1
        fi
    fi
}

# Deployment
deploy_production() {
    warning "This will deploy to production!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Deploying to production..."

        # Check if production compose file exists
        if [ -f "docker-compose.prod.yml" ]; then
            docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
        else
            docker-compose up -d
        fi

        success "Production deployment completed"
        show_service_urls
    else
        info "Production deployment cancelled"
    fi
}

# Backup and restore
backup_data() {
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"

    log "Creating backup in $backup_dir..."

    # Backup databases
    if docker-compose ps postgres | grep -q "Up"; then
        log "Backing up PostgreSQL databases..."
        docker-compose exec -T postgres pg_dumpall -U postgres > "$backup_dir/postgres.sql"
    fi

    # Backup Redis
    if docker-compose ps redis | grep -q "Up"; then
        log "Backing up Redis..."
        docker-compose exec -T redis redis-cli BGSAVE
        docker cp $(docker-compose ps -q redis):/data/dump.rdb "$backup_dir/redis.rdb"
    fi

    # Backup configuration
    cp .env "$backup_dir/env.backup" 2>/dev/null || true
    cp docker-compose.yml "$backup_dir/docker-compose.yml.backup"

    success "Backup completed: $backup_dir"
}

# Update system
update_system() {
    log "Updating system..."

    # Pull latest images
    log "Pulling latest images..."
    docker-compose pull

    # Rebuild services
    log "Rebuilding services..."
    docker-compose build --pull

    # Restart services
    log "Restarting services..."
    docker-compose up -d

    success "System updated"
}

# Help and documentation
show_help() {
    show_logo
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo
    echo "🚀 DEVELOPMENT COMMANDS:"
    echo "  start                   Start all services"
    echo "  stop                    Stop all services"
    echo "  restart                 Restart all services"
    echo "  logs [SERVICE]          Show logs (optionally for specific service)"
    echo
    echo "🔧 BUILD COMMANDS:"
    echo "  build [SERVICE]         Build all services or specific service"
    echo "  rebuild                 Force rebuild all services"
    echo
    echo "🧹 CLEANUP COMMANDS:"
    echo "  clean                   Clean containers"
    echo "  clean-images            Clean images"
    echo "  clean-volumes           Clean volumes"
    echo "  clean-all               Full cleanup"
    echo
    echo "📊 MONITORING COMMANDS:"
    echo "  status                  Show service status"
    echo "  health                  Run health check"
    echo "  urls                    Show service URLs"
    echo
    echo "🗄️ DATABASE COMMANDS:"
    echo "  migrate                 Run database migrations"
    echo "  db-reset                Reset database (destroys data!)"
    echo
    echo "🧪 TESTING COMMANDS:"
    echo "  test                    Run tests"
    echo
    echo "⚙️ CONFIGURATION:"
    echo "  config                  Show current configuration"
    echo "  edit-config             Edit .env configuration"
    echo
    echo "🚢 DEPLOYMENT:"
    echo "  deploy                  Deploy to production"
    echo "  backup                  Backup data"
    echo "  update                  Update system"
    echo
    echo "📚 HELP:"
    echo "  help                    Show this help"
    echo "  version                 Show version"
    echo
    echo "Examples:"
    echo "  $0 start                # Start all services"
    echo "  $0 logs gateway         # Show gateway logs"
    echo "  $0 build auth           # Build auth service"
    echo "  $0 health               # Check system health"
}

show_version() {
    echo "$PROJECT_NAME v$PROJECT_VERSION"
    echo "Docker version: $(docker --version)"
    echo "Docker Compose version: $(docker-compose version --short)"
}

# Main command dispatcher
main() {
    # Show logo for interactive usage
    if [ -t 1 ]; then
        show_logo
    fi

    # Check prerequisites for most commands
    case "${1:-help}" in
        "help"|"version"|"-h"|"--help"|"-v"|"--version")
            # Skip prerequisite check for help commands
            ;;
        *)
            check_prerequisites
            ;;
    esac

    # Execute command
    case "${1:-help}" in
        # Development commands
        "start"|"up")
            dev_start
            ;;
        "stop"|"down")
            dev_stop
            ;;
        "restart")
            dev_restart
            ;;
        "logs"|"log")
            dev_logs "$2"
            ;;

        # Build commands
        "build")
            if [ -n "$2" ]; then
                build_service "$2"
            else
                build_all
            fi
            ;;
        "rebuild")
            clean_images
            build_all
            ;;

        # Cleanup commands
        "clean")
            clean_containers
            ;;
        "clean-images")
            clean_images
            ;;
        "clean-volumes")
            clean_volumes
            ;;
        "clean-all")
            clean_all
            ;;

        # Monitoring commands
        "status"|"ps")
            show_status
            ;;
        "health"|"check")
            health_check
            ;;
        "urls")
            show_service_urls
            ;;

        # Database commands
        "migrate")
            db_migrate
            ;;
        "db-reset")
            db_reset
            ;;

        # Testing
        "test"|"tests")
            run_tests
            ;;

        # Configuration
        "config")
            show_config
            ;;
        "edit-config")
            edit_config
            ;;

        # Deployment
        "deploy")
            deploy_production
            ;;
        "backup")
            backup_data
            ;;
        "update")
            update_system
            ;;

        # Help and version
        "help"|"-h"|"--help")
            show_help
            ;;
        "version"|"-v"|"--version")
            show_version
            ;;

        # Unknown command
        *)
            error "Unknown command: $1"
            echo
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
