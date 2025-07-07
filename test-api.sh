#!/bin/bash

# Comprehensive API Testing Script for Silence Project
# Tests all services and their endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost"
AUTH_PORT="8081"
GATEWAY_PORT="8080"
ANALYTICS_PORT="8082"
DPI_BYPASS_PORT="8083"
VPN_CORE_PORT="8084"
SERVER_MANAGER_PORT="8085"
NOTIFICATIONS_PORT="8087"

# Test data
TEST_USER_EMAIL="testuser@example.com"
TEST_USER_PASSWORD="testpassword123"
TEST_USER_USERNAME="testuser"

# Functions
print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Test service health
test_health() {
    local service_name=$1
    local port=$2
    local endpoint=${3:-"/health"}

    print_test "Testing $service_name health check"

    response=$(curl -s -w "%{http_code}" -o /tmp/health_response "$BASE_URL:$port$endpoint")
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ]; then
        content=$(cat /tmp/health_response)
        print_success "$service_name is healthy: $content"
        return 0
    else
        print_error "$service_name health check failed (HTTP $http_code)"
        return 1
    fi
}

# Get auth token
get_auth_token() {
    print_test "Getting authentication token"

    # First try to register user
    register_response=$(curl -s -X POST "$BASE_URL:$AUTH_PORT/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$TEST_USER_USERNAME\",\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}" \
        2>/dev/null || echo "")

    # Then login to get token
    login_response=$(curl -s -X POST "$BASE_URL:$AUTH_PORT/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")

    if echo "$login_response" | grep -q "token"; then
        TOKEN=$(echo "$login_response" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "")
        if [ -n "$TOKEN" ]; then
            print_success "Authentication token obtained"
            return 0
        fi
    fi

    print_error "Failed to get authentication token"
    return 1
}

# Test auth service
test_auth_service() {
    print_header "Testing Auth Service"

    # Health check
    test_health "Auth Service" "$AUTH_PORT"

    # Register user
    print_test "Testing user registration"
    register_response=$(curl -s -X POST "$BASE_URL:$AUTH_PORT/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"newuser\",\"email\":\"newuser@example.com\",\"password\":\"password123\"}")

    if echo "$register_response" | grep -q "token\|already exists"; then
        print_success "User registration working"
    else
        print_error "User registration failed: $register_response"
    fi

    # Login
    print_test "Testing user login"
    login_response=$(curl -s -X POST "$BASE_URL:$AUTH_PORT/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")

    if echo "$login_response" | grep -q "token"; then
        print_success "User login working"
    else
        print_error "User login failed: $login_response"
    fi
}

# Test gateway service
test_gateway_service() {
    print_header "Testing Gateway Service"

    # Health check
    test_health "Gateway Service" "$GATEWAY_PORT"

    # Test auth endpoints through gateway
    print_test "Testing auth through gateway"
    gateway_auth_response=$(curl -s -X POST "$BASE_URL:$GATEWAY_PORT/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")

    if echo "$gateway_auth_response" | grep -q "token"; then
        print_success "Gateway auth proxy working"
    else
        print_error "Gateway auth proxy failed: $gateway_auth_response"
    fi

    # Test with authentication
    if [ -n "$TOKEN" ]; then
        print_test "Testing authenticated endpoint through gateway"
        vpn_response=$(curl -s "$BASE_URL:$GATEWAY_PORT/api/v1/vpn/tunnels/list" \
            -H "Authorization: Bearer $TOKEN")

        if echo "$vpn_response" | grep -q -v "missing.*Authorization\|invalid.*token"; then
            print_success "Authenticated gateway request working"
        else
            print_info "Gateway authentication may need configuration: $vpn_response"
        fi
    fi
}

# Test analytics service
test_analytics_service() {
    print_header "Testing Analytics Service"

    # Health check
    test_health "Analytics Service" "$ANALYTICS_PORT"

    # Test connections endpoint
    print_test "Testing analytics connections endpoint"
    analytics_response=$(curl -s "$BASE_URL:$ANALYTICS_PORT/api/v1/analytics/connections")

    if echo "$analytics_response" | grep -q "metrics\|total"; then
        print_success "Analytics connections endpoint working"
    else
        print_error "Analytics connections endpoint failed: $analytics_response"
    fi

    # Test other endpoints
    endpoints=("bypass-effectiveness" "user-activity" "server-load" "errors")
    for endpoint in "${endpoints[@]}"; do
        print_test "Testing analytics $endpoint endpoint"
        response=$(curl -s "$BASE_URL:$ANALYTICS_PORT/api/v1/analytics/$endpoint")
        if [ $? -eq 0 ]; then
            print_success "Analytics $endpoint endpoint responding"
        else
            print_error "Analytics $endpoint endpoint failed"
        fi
    done
}

# Test DPI bypass service
test_dpi_bypass_service() {
    print_header "Testing DPI Bypass Service"

    # Health check
    test_health "DPI Bypass Service" "$DPI_BYPASS_PORT"

    # Test bypass list
    print_test "Testing DPI bypass list endpoint"
    bypass_response=$(curl -s "$BASE_URL:$DPI_BYPASS_PORT/api/v1/bypass")

    if [ $? -eq 0 ]; then
        print_success "DPI bypass list endpoint working: $bypass_response"
    else
        print_error "DPI bypass list endpoint failed"
    fi

    # Test bypass creation
    print_test "Testing DPI bypass creation"
    create_response=$(curl -s -X POST "$BASE_URL:$DPI_BYPASS_PORT/api/v1/bypass" \
        -H "Content-Type: application/json" \
        -d '{"name":"test-bypass","target":"example.com","method":"domain-fronting"}')

    if [ $? -eq 0 ]; then
        print_success "DPI bypass creation endpoint responding"
    else
        print_error "DPI bypass creation endpoint failed"
    fi
}

# Test VPN core service
test_vpn_core_service() {
    print_header "Testing VPN Core Service"

    # Health check
    test_health "VPN Core Service" "$VPN_CORE_PORT"

    # Test tunnels list
    print_test "Testing VPN tunnels list endpoint"
    tunnels_response=$(curl -s "$BASE_URL:$VPN_CORE_PORT/tunnels/list")

    if [ $? -eq 0 ]; then
        print_success "VPN tunnels list endpoint working: $tunnels_response"
    else
        print_error "VPN tunnels list endpoint failed"
    fi

    # Test peers list
    print_test "Testing VPN peers list endpoint"
    peers_response=$(curl -s "$BASE_URL:$VPN_CORE_PORT/peers/list")

    if [ $? -eq 0 ]; then
        print_success "VPN peers list endpoint working"
    else
        print_error "VPN peers list endpoint failed"
    fi
}

# Test server manager service
test_server_manager_service() {
    print_header "Testing Server Manager Service"

    # Health check
    test_health "Server Manager Service" "$SERVER_MANAGER_PORT"

    # Test servers list
    print_test "Testing server manager servers endpoint"
    servers_response=$(curl -s "$BASE_URL:$SERVER_MANAGER_PORT/api/v1/servers")

    if [ $? -eq 0 ]; then
        print_success "Server manager servers endpoint working: $servers_response"
    else
        print_error "Server manager servers endpoint failed"
    fi
}

# Test notifications service
test_notifications_service() {
    print_header "Testing Notifications Service"

    # Health check
    test_health "Notifications Service" "$NOTIFICATIONS_PORT" "/healthz"

    # Test notification sending
    print_test "Testing notification sending"
    notification_response=$(curl -s -X POST "$BASE_URL:$NOTIFICATIONS_PORT/notifications" \
        -H "Content-Type: application/json" \
        -d '{"type":"test","recipients":["test@example.com"],"channels":["email"],"title":"Test Notification","message":"This is a test notification"}')

    if echo "$notification_response" | grep -q "ok"; then
        print_success "Notification sending working"
    else
        print_error "Notification sending failed: $notification_response"
    fi
}

# Test infrastructure services
test_infrastructure() {
    print_header "Testing Infrastructure Services"

    # Test PostgreSQL
    print_test "Testing PostgreSQL connection"
    if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
        print_success "PostgreSQL is accessible"
    else
        print_error "PostgreSQL is not accessible"
    fi

    # Test Redis
    print_test "Testing Redis connection"
    if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
        print_success "Redis is accessible"
    else
        print_error "Redis is not accessible"
    fi

    # Test RabbitMQ
    print_test "Testing RabbitMQ connection"
    rabbitmq_status=$(curl -s -u admin:admin "http://localhost:15672/api/overview" 2>/dev/null || echo "failed")
    if echo "$rabbitmq_status" | grep -q "management_version\|cluster_name"; then
        print_success "RabbitMQ is accessible"
    else
        print_error "RabbitMQ is not accessible"
    fi

    # Test InfluxDB
    print_test "Testing InfluxDB connection"
    influxdb_status=$(curl -s "http://localhost:8086/health" 2>/dev/null || echo "failed")
    if echo "$influxdb_status" | grep -q "pass\|name"; then
        print_success "InfluxDB is accessible"
    else
        print_error "InfluxDB is not accessible"
    fi
}

# Check if services are running
check_services_running() {
    print_header "Checking Service Status"

    services=("postgres" "redis" "rabbitmq" "influxdb" "auth" "gateway" "analytics" "dpi-bypass" "vpn-core" "server-manager" "notifications")

    for service in "${services[@]}"; do
        if docker-compose ps | grep -q "${service}.*Up"; then
            print_success "$service is running"
        else
            print_error "$service is not running"
            return 1
        fi
    done
}

# Main test execution
main() {
    print_header "Silence Project API Testing"

    # Check if services are running
    if ! check_services_running; then
        print_error "Some services are not running. Please start them first with: docker-compose up -d"
        exit 1
    fi

    # Get authentication token
    get_auth_token

    # Test infrastructure
    test_infrastructure

    # Test application services
    test_auth_service
    test_gateway_service
    test_analytics_service
    test_dpi_bypass_service
    test_vpn_core_service
    test_server_manager_service
    test_notifications_service

    print_header "Testing Complete"
    print_info "All tests completed. Check output above for results."

    # Cleanup
    rm -f /tmp/health_response token.txt 2>/dev/null || true
}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  --auth-only    Test only auth service"
    echo "  --gateway-only Test only gateway service"
    echo "  --infra-only   Test only infrastructure services"
    echo ""
    echo "Examples:"
    echo "  $0                # Run all tests"
    echo "  $0 --auth-only    # Test only auth service"
    echo "  $0 --infra-only   # Test only infrastructure"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    --auth-only)
        check_services_running && get_auth_token && test_auth_service
        ;;
    --gateway-only)
        check_services_running && get_auth_token && test_gateway_service
        ;;
    --infra-only)
        check_services_running && test_infrastructure
        ;;
    "")
        main
        ;;
    *)
        echo "Unknown option: $1"
        show_help
        exit 1
        ;;
esac
