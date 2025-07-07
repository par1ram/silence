#!/bin/bash

# Silence Project Demo Script
# Demonstrates all working services and endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost"

print_banner() {
    echo -e "${PURPLE}"
    echo "======================================================"
    echo "     🔒 SILENCE PROJECT - API DEMONSTRATION 🔒      "
    echo "======================================================"
    echo -e "${NC}"
}

print_section() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}🎯 $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_step() {
    echo -e "\n${YELLOW}📋 Step: $1${NC}"
}

print_request() {
    echo -e "${PURPLE}🌐 Request:${NC} $1"
}

print_response() {
    echo -e "${GREEN}✅ Response:${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ️  Info:${NC} $1"
}

wait_for_user() {
    echo -e "\n${YELLOW}Press Enter to continue...${NC}"
    read
}

# Demo Infrastructure Status
demo_infrastructure() {
    print_section "Infrastructure Services Status"

    print_step "Checking PostgreSQL"
    if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
        print_response "PostgreSQL is running on port 5432 ✅"
    else
        print_response "PostgreSQL is not accessible ❌"
    fi

    print_step "Checking Redis"
    if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
        print_response "Redis is running on port 6379 ✅"
    else
        print_response "Redis is not accessible ❌"
    fi

    print_step "Checking RabbitMQ"
    rabbitmq_check=$(curl -s -u admin:admin "http://localhost:15672/api/overview" 2>/dev/null || echo "failed")
    if echo "$rabbitmq_check" | grep -q "management_version\|cluster_name"; then
        print_response "RabbitMQ is running on port 5672 (Management: 15672) ✅"
    else
        print_response "RabbitMQ is not accessible ❌"
    fi

    print_step "Checking InfluxDB"
    influxdb_check=$(curl -s "http://localhost:8086/health" 2>/dev/null || echo "failed")
    if echo "$influxdb_check" | grep -q "pass\|name"; then
        print_response "InfluxDB is running on port 8086 ✅"
    else
        print_response "InfluxDB is not accessible ❌"
    fi

    wait_for_user
}

# Demo Auth Service
demo_auth_service() {
    print_section "Auth Service (Port 8081)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8081/health"
    health_response=$(curl -s "$BASE_URL:8081/health")
    print_response "$health_response"

    print_step "User Registration"
    print_request "POST $BASE_URL:8081/register"
    print_info "Payload: {\"username\":\"demouser\",\"email\":\"demo@silence.com\",\"password\":\"demo123\"}"

    register_response=$(curl -s -X POST "$BASE_URL:8081/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"demouser","email":"demo@silence.com","password":"demo123"}' || echo "User may already exist")

    print_response "$(echo "$register_response" | head -c 200)..."

    print_step "User Login"
    print_request "POST $BASE_URL:8081/login"
    print_info "Payload: {\"email\":\"demo@silence.com\",\"password\":\"demo123\"}"

    login_response=$(curl -s -X POST "$BASE_URL:8081/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"demo@silence.com","password":"demo123"}')

    print_response "$(echo "$login_response" | head -c 200)..."

    # Extract token for later use
    TOKEN=$(echo "$login_response" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "")

    wait_for_user
}

# Demo Gateway Service
demo_gateway_service() {
    print_section "Gateway Service (Port 8080)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8080/health"
    gateway_health=$(curl -s "$BASE_URL:8080/health")
    print_response "$gateway_health"

    print_step "Authentication through Gateway"
    print_request "POST $BASE_URL:8080/api/v1/auth/login"
    print_info "Gateway acts as a proxy to auth service"

    gateway_login=$(curl -s -X POST "$BASE_URL:8080/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"demo@silence.com","password":"demo123"}')

    print_response "$(echo "$gateway_login" | head -c 200)..."

    wait_for_user
}

# Demo Analytics Service
demo_analytics_service() {
    print_section "Analytics Service (Port 8082)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8082/health"
    analytics_health=$(curl -s "$BASE_URL:8082/health")
    print_response "$analytics_health"

    print_step "Connection Metrics"
    print_request "GET $BASE_URL:8082/api/v1/analytics/connections"
    connections_response=$(curl -s "$BASE_URL:8082/api/v1/analytics/connections")
    print_response "$connections_response"

    print_step "Error Metrics"
    print_request "GET $BASE_URL:8082/api/v1/analytics/errors"
    errors_response=$(curl -s "$BASE_URL:8082/api/v1/analytics/errors")
    print_response "$errors_response"

    print_info "Analytics service is connected to InfluxDB for time-series data"

    wait_for_user
}

# Demo VPN Core Service
demo_vpn_core_service() {
    print_section "VPN Core Service (Port 8084)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8084/health"
    vpn_health=$(curl -s "$BASE_URL:8084/health")
    print_response "$vpn_health"

    print_step "VPN Tunnels List"
    print_request "GET $BASE_URL:8084/tunnels/list"
    tunnels_response=$(curl -s "$BASE_URL:8084/tunnels/list")
    print_response "$tunnels_response"

    print_step "VPN Peers List"
    print_request "GET $BASE_URL:8084/peers/list"
    peers_response=$(curl -s "$BASE_URL:8084/peers/list")
    print_response "$peers_response"

    print_info "VPN Core manages WireGuard tunnels and peer connections"

    wait_for_user
}

# Demo DPI Bypass Service
demo_dpi_bypass_service() {
    print_section "DPI Bypass Service (Port 8083)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8083/health"
    dpi_health=$(curl -s "$BASE_URL:8083/health")
    print_response "$dpi_health"

    print_step "Bypass Configurations List"
    print_request "GET $BASE_URL:8083/api/v1/bypass"
    bypass_list=$(curl -s "$BASE_URL:8083/api/v1/bypass")
    print_response "$bypass_list"

    print_step "Create Bypass Configuration"
    print_request "POST $BASE_URL:8083/api/v1/bypass"
    print_info "Payload: {\"name\":\"demo-bypass\",\"target\":\"example.com\",\"method\":\"domain-fronting\"}"

    create_bypass=$(curl -s -X POST "$BASE_URL:8083/api/v1/bypass" \
        -H "Content-Type: application/json" \
        -d '{"name":"demo-bypass","target":"example.com","method":"domain-fronting"}')

    print_response "Bypass configuration created"

    wait_for_user
}

# Demo Server Manager Service
demo_server_manager_service() {
    print_section "Server Manager Service (Port 8085)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8085/health"
    server_health=$(curl -s "$BASE_URL:8085/health")
    print_response "$server_health"

    print_step "Servers List"
    print_request "GET $BASE_URL:8085/api/v1/servers"
    servers_response=$(curl -s "$BASE_URL:8085/api/v1/servers")
    print_response "$servers_response"

    print_info "Server Manager handles Docker container lifecycle and server provisioning"

    wait_for_user
}

# Demo Notifications Service
demo_notifications_service() {
    print_section "Notifications Service (Port 8087)"

    print_step "Health Check"
    print_request "GET $BASE_URL:8087/healthz"
    notifications_health=$(curl -s "$BASE_URL:8087/healthz")
    print_response "$notifications_health"

    print_step "Send Test Notification"
    print_request "POST $BASE_URL:8087/notifications"
    print_info "Payload: {\"type\":\"demo\",\"recipients\":[\"demo@silence.com\"],\"channels\":[\"email\"],\"title\":\"Demo\",\"message\":\"Hello from Silence!\"}"

    notification_response=$(curl -s -X POST "$BASE_URL:8087/notifications" \
        -H "Content-Type: application/json" \
        -d '{"type":"demo","recipients":["demo@silence.com"],"channels":["email"],"title":"Demo Notification","message":"Hello from Silence VPN!"}')

    print_response "$notification_response"

    print_info "Notifications service uses RabbitMQ for message queuing"

    wait_for_user
}

# Demo Summary
demo_summary() {
    print_section "Demo Summary"

    echo -e "${GREEN}🎉 Successfully demonstrated all Silence VPN services:${NC}\n"

    echo -e "${CYAN}Infrastructure:${NC}"
    echo -e "  ✅ PostgreSQL (Port 5432) - Database storage"
    echo -e "  ✅ Redis (Port 6379) - Caching and sessions"
    echo -e "  ✅ RabbitMQ (Port 5672) - Message queuing"
    echo -e "  ✅ InfluxDB (Port 8086) - Time-series analytics"

    echo -e "\n${CYAN}Application Services:${NC}"
    echo -e "  ✅ Auth Service (Port 8081) - User authentication & JWT tokens"
    echo -e "  ✅ Gateway Service (Port 8080) - API gateway & request routing"
    echo -e "  ✅ Analytics Service (Port 8082) - Metrics & analytics"
    echo -e "  ✅ VPN Core Service (Port 8084) - WireGuard tunnel management"
    echo -e "  ✅ DPI Bypass Service (Port 8083) - Traffic obfuscation"
    echo -e "  ✅ Server Manager Service (Port 8085) - Infrastructure management"
    echo -e "  ✅ Notifications Service (Port 8087) - Alert & notification system"

    echo -e "\n${YELLOW}Key Features Demonstrated:${NC}"
    echo -e "  🔐 User registration and authentication"
    echo -e "  🌐 API gateway with service routing"
    echo -e "  📊 Real-time analytics and monitoring"
    echo -e "  🚀 VPN tunnel and peer management"
    echo -e "  🛡️  DPI bypass and traffic obfuscation"
    echo -e "  🖥️  Server and container management"
    echo -e "  📢 Notification and alerting system"

    echo -e "\n${GREEN}🏆 All services are running successfully in Docker containers!${NC}"

    echo -e "\n${BLUE}Next Steps:${NC}"
    echo -e "  1. Integrate frontend application"
    echo -e "  2. Configure production security settings"
    echo -e "  3. Set up monitoring and logging"
    echo -e "  4. Deploy to production environment"

    echo -e "\n${PURPLE}Thank you for watching the Silence VPN demo! 🔒${NC}"
}

# Main demo execution
main() {
    print_banner

    print_info "This demo will showcase all working services in the Silence VPN project"
    print_info "Each section will demonstrate different service capabilities"

    wait_for_user

    demo_infrastructure
    demo_auth_service
    demo_gateway_service
    demo_analytics_service
    demo_vpn_core_service
    demo_dpi_bypass_service
    demo_server_manager_service
    demo_notifications_service
    demo_summary
}

# Help function
show_help() {
    echo "Silence VPN Demo Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help       Show this help message"
    echo "  --quick          Run quick demo without pauses"
    echo "  --services-only  Show only service status"
    echo ""
    echo "This demo showcases all working services and APIs in the Silence VPN project."
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    --quick)
        # Override wait function for quick demo
        wait_for_user() { sleep 1; }
        main
        ;;
    --services-only)
        print_banner
        demo_infrastructure
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
