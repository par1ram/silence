#!/bin/bash

# Project Health Check Script for Silence VPN
# This script performs a comprehensive health check of the entire project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1"
    ((PASSED_CHECKS++))
}

log_error() {
    echo -e "${RED}[✗ FAIL]${NC} $1"
    ((FAILED_CHECKS++))
}

log_warning() {
    echo -e "${YELLOW}[⚠ WARN]${NC} $1"
    ((WARNING_CHECKS++))
}

log_section() {
    echo
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE} $1${NC}"
    echo -e "${PURPLE}========================================${NC}"
    echo
}

# Check if command exists
check_command() {
    local cmd=$1
    local name=${2:-$cmd}

    ((TOTAL_CHECKS++))
    if command -v $cmd >/dev/null 2>&1; then
        log_success "$name is installed"
        return 0
    else
        log_error "$name is not installed"
        return 1
    fi
}

# Check if port is open
check_port() {
    local port=$1
    local service=$2

    ((TOTAL_CHECKS++))
    if nc -z localhost $port 2>/dev/null; then
        log_success "$service is running on port $port"
        return 0
    else
        log_error "$service is not running on port $port"
        return 1
    fi
}

# Check if file exists
check_file() {
    local file=$1
    local description=${2:-$file}

    ((TOTAL_CHECKS++))
    if [ -f "$file" ]; then
        log_success "$description exists"
        return 0
    else
        log_error "$description is missing"
        return 1
    fi
}

# Check if directory exists
check_directory() {
    local dir=$1
    local description=${2:-$dir}

    ((TOTAL_CHECKS++))
    if [ -d "$dir" ]; then
        log_success "$description exists"
        return 0
    else
        log_error "$description is missing"
        return 1
    fi
}

# Check Docker containers
check_docker_container() {
    local container=$1
    local service=$2

    ((TOTAL_CHECKS++))
    if docker ps --format "table {{.Names}}" | grep -q "^$container$"; then
        log_success "$service container is running"
        return 0
    else
        log_error "$service container is not running"
        return 1
    fi
}

# Check HTTP endpoint
check_http_endpoint() {
    local url=$1
    local description=$2
    local expected_status=${3:-200}

    ((TOTAL_CHECKS++))
    local status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")

    if [ "$status" = "$expected_status" ]; then
        log_success "$description (HTTP $status)"
        return 0
    else
        log_error "$description (Expected HTTP $expected_status, got $status)"
        return 1
    fi
}

# Main health check functions
check_prerequisites() {
    log_section "Prerequisites Check"

    check_command "go" "Go"
    check_command "docker" "Docker"
    check_command "docker-compose" "Docker Compose"
    check_command "task" "Task CLI"
    check_command "curl" "curl"
    check_command "nc" "netcat"
    check_command "jq" "jq (JSON processor)" || log_warning "jq not found - some tests may be limited"

    # Check Go version
    ((TOTAL_CHECKS++))
    if command -v go >/dev/null 2>&1; then
        local go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1)
        local major=$(echo $go_version | cut -d'.' -f1 | sed 's/go//')
        local minor=$(echo $go_version | cut -d'.' -f2)

        if [ "$major" -gt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -ge 23 ]); then
            log_success "Go version $go_version (>= 1.23 required)"
        else
            log_error "Go version $go_version (>= 1.23 required)"
        fi
    fi
}

check_project_structure() {
    log_section "Project Structure Check"

    # Core directories
    check_directory "api" "API services directory"
    check_directory "rpc" "RPC services directory"
    check_directory "shared" "Shared libraries directory"
    check_directory "scripts" "Scripts directory"
    check_directory "docs" "Documentation directory"

    # Service directories
    check_directory "api/auth" "Auth service"
    check_directory "api/gateway" "Gateway service"
    check_directory "rpc/analytics" "Analytics service"
    check_directory "rpc/vpn-core" "VPN Core service"
    check_directory "rpc/server-manager" "Server Manager service"
    check_directory "rpc/dpi-bypass" "DPI Bypass service"
    check_directory "rpc/notifications" "Notifications service"

    # Key files
    check_file "docker-compose.yml" "Docker Compose configuration"
    check_file "Taskfile.yml" "Task configuration"
    check_file ".env.development" "Development environment variables"
    check_file "README.md" "README file"
    check_file "PROJECT_STATUS_REPORT.md" "Project status report"
    check_file "QUICK_START.md" "Quick start guide"
}

check_configuration_files() {
    log_section "Configuration Files Check"

    # Check for go.mod files
    local services=("api/auth" "api/gateway" "rpc/analytics" "rpc/vpn-core" "rpc/server-manager" "rpc/dpi-bypass" "rpc/notifications" "shared")

    for service in "${services[@]}"; do
        check_file "$service/go.mod" "$service/go.mod"
    done

    # Check for Dockerfiles
    local docker_services=("api/auth" "api/gateway" "rpc/analytics" "rpc/vpn-core" "rpc/server-manager" "rpc/dpi-bypass" "rpc/notifications")

    for service in "${docker_services[@]}"; do
        check_file "$service/Dockerfile" "$service/Dockerfile"
    done

    # Check for Air config files
    for service in "${docker_services[@]}"; do
        if [ -f "$service/.air.toml" ]; then
            log_success "$service/.air.toml (hot reload config)"
        else
            log_warning "$service/.air.toml missing (hot reload may not work)"
        fi
        ((TOTAL_CHECKS++))
    done
}

check_build_artifacts() {
    log_section "Build Artifacts Check"

    # Check for compiled binaries
    local services=("api/auth" "api/gateway" "rpc/analytics" "rpc/vpn-core" "rpc/server-manager" "rpc/dpi-bypass" "rpc/notifications")

    for service in "${services[@]}"; do
        local service_name=$(basename $service)
        if [ -f "$service/bin/$service_name" ]; then
            log_success "$service binary is built"
        else
            log_warning "$service binary not found (run 'task build' to build)"
        fi
        ((TOTAL_CHECKS++))
    done
}

check_infrastructure() {
    log_section "Infrastructure Check"

    # Check Docker containers
    check_docker_container "silence_postgres" "PostgreSQL"
    check_docker_container "silence_redis" "Redis"
    check_docker_container "silence_rabbitmq" "RabbitMQ"
    check_docker_container "silence_influxdb" "InfluxDB"
    check_docker_container "silence_clickhouse" "ClickHouse"

    # Check infrastructure ports
    check_port 5432 "PostgreSQL"
    check_port 6379 "Redis"
    check_port 5672 "RabbitMQ"
    check_port 15672 "RabbitMQ Management"
    check_port 8086 "InfluxDB"
    check_port 9000 "ClickHouse"
    check_port 8123 "ClickHouse HTTP"
}

check_application_services() {
    log_section "Application Services Check"

    # Check service ports
    local services=(
        "8081:Auth Service"
        "8080:Gateway Service"
        "8082:Analytics Service"
        "8084:VPN Core Service"
        "8085:Server Manager Service"
        "8083:DPI Bypass Service"
        "8087:Notifications Service"
    )

    for service_info in "${services[@]}"; do
        local port=$(echo $service_info | cut -d':' -f1)
        local name=$(echo $service_info | cut -d':' -f2)
        check_port $port "$name"
    done
}

check_api_endpoints() {
    log_section "API Endpoints Check"

    # Basic connectivity checks
    local gateway_url="http://localhost:8080"

    # Check if gateway is accessible
    ((TOTAL_CHECKS++))
    if curl -s --connect-timeout 5 "$gateway_url" >/dev/null 2>&1; then
        log_success "Gateway is accessible"

        # Check specific endpoints
        check_http_endpoint "$gateway_url/api/docs" "API documentation endpoint"
        check_http_endpoint "$gateway_url/docs/swagger/index.html" "Swagger UI" "200"

        # Check service-specific health endpoints (if implemented)
        check_http_endpoint "$gateway_url/api/v1/auth/health" "Auth service health" || true
        check_http_endpoint "$gateway_url/api/v1/vpn/health" "VPN service health" || true

    else
        log_error "Gateway is not accessible"
    fi
}

check_databases() {
    log_section "Database Check"

    # Check PostgreSQL databases
    local databases=("silence_auth" "silence_server_manager" "silence_vpn")

    for db in "${databases[@]}"; do
        ((TOTAL_CHECKS++))
        if docker exec silence_postgres psql -U postgres -d $db -c "SELECT 1;" >/dev/null 2>&1; then
            log_success "PostgreSQL database '$db' is accessible"
        else
            log_error "PostgreSQL database '$db' is not accessible"
        fi
    done

    # Check Redis connectivity
    ((TOTAL_CHECKS++))
    if docker exec silence_redis redis-cli ping >/dev/null 2>&1; then
        log_success "Redis is accessible"
    else
        log_error "Redis is not accessible"
    fi

    # Check ClickHouse connectivity
    ((TOTAL_CHECKS++))
    if docker exec silence_clickhouse clickhouse-client --query "SELECT 1" >/dev/null 2>&1; then
        log_success "ClickHouse is accessible"
    else
        log_error "ClickHouse is not accessible"
    fi
}

check_scripts() {
    log_section "Scripts Check"

    # Check if scripts are executable
    local scripts=(
        "scripts/start-services.sh"
        "scripts/run-single-service.sh"
        "scripts/test-endpoints.sh"
        "scripts/test-swagger.sh"
        "scripts/project-health-check.sh"
    )

    for script in "${scripts[@]}"; do
        ((TOTAL_CHECKS++))
        if [ -x "$script" ]; then
            log_success "$script is executable"
        else
            log_warning "$script is not executable (run 'chmod +x $script')"
        fi
    done

    # Check script syntax (basic)
    for script in "${scripts[@]}"; do
        if [ -f "$script" ]; then
            ((TOTAL_CHECKS++))
            if bash -n "$script" 2>/dev/null; then
                log_success "$script has valid syntax"
            else
                log_error "$script has syntax errors"
            fi
        fi
    done
}

check_documentation() {
    log_section "Documentation Check"

    # Check for documentation files
    local docs=(
        "README.md:Main README"
        "QUICK_START.md:Quick start guide"
        "PROJECT_STATUS_REPORT.md:Project status report"
        "API_ENDPOINTS_TEST_REPORT.md:API test report"
        "SWAGGER_VERIFICATION_REPORT.md:Swagger verification report"
    )

    for doc_info in "${docs[@]}"; do
        local file=$(echo $doc_info | cut -d':' -f1)
        local desc=$(echo $doc_info | cut -d':' -f2)
        check_file "$file" "$desc"
    done

    # Check documentation quality
    ((TOTAL_CHECKS++))
    local readme_size=$(wc -c < "README.md" 2>/dev/null || echo "0")
    if [ "$readme_size" -gt 1000 ]; then
        log_success "README.md has substantial content ($readme_size bytes)"
    else
        log_warning "README.md seems too short ($readme_size bytes)"
    fi
}

check_security() {
    log_section "Security Check"

    # Check for sensitive files that shouldn't be committed
    local sensitive_files=(".env" "*.key" "*.pem" "*.crt" "config/secrets*")

    for pattern in "${sensitive_files[@]}"; do
        ((TOTAL_CHECKS++))
        if find . -name "$pattern" -type f 2>/dev/null | grep -q .; then
            log_warning "Found potentially sensitive files matching '$pattern'"
        else
            log_success "No sensitive files found for pattern '$pattern'"
        fi
    done

    # Check for default passwords in development config
    ((TOTAL_CHECKS++))
    if grep -q "password.*password" .env.development 2>/dev/null; then
        log_warning "Default passwords found in .env.development (OK for development)"
    else
        log_success "No obvious default passwords in configuration"
    fi

    # Check for TODO security items
    ((TOTAL_CHECKS++))
    local security_todos=$(find . -name "*.go" -o -name "*.md" | xargs grep -l "TODO.*security\|FIXME.*security" 2>/dev/null | wc -l)
    if [ "$security_todos" -gt 0 ]; then
        log_warning "$security_todos files contain security TODOs"
    else
        log_success "No security TODOs found in codebase"
    fi
}

run_smoke_tests() {
    log_section "Smoke Tests"

    # Test if we can run basic task commands
    ((TOTAL_CHECKS++))
    if task --version >/dev/null 2>&1; then
        log_success "Task CLI is working"
    else
        log_error "Task CLI is not working"
    fi

    # Test if we can build (dry run)
    ((TOTAL_CHECKS++))
    if task build --dry >/dev/null 2>&1; then
        log_success "Build task is configured correctly"
    else
        log_warning "Build task may have issues"
    fi

    # Test Docker Compose syntax
    ((TOTAL_CHECKS++))
    if docker-compose config >/dev/null 2>&1; then
        log_success "Docker Compose configuration is valid"
    else
        log_error "Docker Compose configuration has errors"
    fi
}

generate_summary() {
    log_section "Health Check Summary"

    echo -e "${CYAN}Total Checks: $TOTAL_CHECKS${NC}"
    echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
    echo -e "${RED}Failed: $FAILED_CHECKS${NC}"
    echo -e "${YELLOW}Warnings: $WARNING_CHECKS${NC}"
    echo

    local success_rate=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))

    if [ $FAILED_CHECKS -eq 0 ]; then
        echo -e "${GREEN}🎉 Project health: EXCELLENT ($success_rate% passed)${NC}"
        echo -e "${GREEN}✅ All critical checks passed!${NC}"
    elif [ $FAILED_CHECKS -le 3 ]; then
        echo -e "${YELLOW}⚠️  Project health: GOOD ($success_rate% passed)${NC}"
        echo -e "${YELLOW}🔧 Minor issues detected, but project is functional${NC}"
    elif [ $FAILED_CHECKS -le 10 ]; then
        echo -e "${YELLOW}⚠️  Project health: FAIR ($success_rate% passed)${NC}"
        echo -e "${YELLOW}🔧 Several issues detected, some functionality may be impaired${NC}"
    else
        echo -e "${RED}❌ Project health: POOR ($success_rate% passed)${NC}"
        echo -e "${RED}🚨 Many issues detected, project needs attention${NC}"
    fi

    echo

    if [ $FAILED_CHECKS -gt 0 ]; then
        echo -e "${BLUE}💡 Recommendations:${NC}"
        if [ $FAILED_CHECKS -gt 5 ]; then
            echo "  • Run 'task infra:up' to start infrastructure"
            echo "  • Run 'task build' to build all services"
            echo "  • Run './scripts/start-services.sh start' to start services"
        fi
        echo "  • Check the failed items above and fix them"
        echo "  • Refer to QUICK_START.md for setup instructions"
        echo "  • Check PROJECT_STATUS_REPORT.md for detailed information"
    else
        echo -e "${BLUE}🚀 Next steps:${NC}"
        echo "  • Run './scripts/test-endpoints.sh' to test API functionality"
        echo "  • Run './scripts/test-swagger.sh' to verify documentation"
        echo "  • Start developing new features!"
    fi
}

# Parse command line arguments
VERBOSE=false
QUICK=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quick)
            QUICK=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -v, --verbose    Enable verbose output"
            echo "  -q, --quick      Quick check (skip some tests)"
            echo "  -h, --help       Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo -e "${CYAN}"
    echo "  ███████╗██╗██╗     ███████╗███╗   ██╗ ██████╗███████╗"
    echo "  ██╔════╝██║██║     ██╔════╝████╗  ██║██╔════╝██╔════╝"
    echo "  ███████╗██║██║     █████╗  ██╔██╗ ██║██║     █████╗  "
    echo "  ╚════██║██║██║     ██╔══╝  ██║╚██╗██║██║     ██╔══╝  "
    echo "  ███████║██║███████╗███████╗██║ ╚████║╚██████╗███████╗"
    echo "  ╚══════╝╚═╝╚══════╝╚══════╝╚═╝  ╚═══╝ ╚═════╝╚══════╝"
    echo "                                                        "
    echo "              🔍 PROJECT HEALTH CHECK 🔍"
    echo -e "${NC}"
    echo

    log_info "Starting comprehensive health check..."
    log_info "Verbose mode: $VERBOSE"
    log_info "Quick mode: $QUICK"
    echo

    # Run all checks
    check_prerequisites
    check_project_structure
    check_configuration_files

    if [ "$QUICK" = false ]; then
        check_build_artifacts
        check_infrastructure
        check_application_services
        check_api_endpoints
        check_databases
        check_scripts
        check_documentation
        check_security
        run_smoke_tests
    else
        log_info "Skipping detailed checks (quick mode)"
    fi

    generate_summary

    # Exit with appropriate code
    if [ $FAILED_CHECKS -eq 0 ]; then
        exit 0
    elif [ $FAILED_CHECKS -le 3 ]; then
        exit 1
    else
        exit 2
    fi
}

# Run main function
main "$@"
