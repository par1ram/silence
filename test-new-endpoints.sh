#!/bin/bash

# Test script for new connection endpoints
# Tests the specialized connection endpoints in Gateway Service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
TEST_USER_EMAIL="testuser@example.com"
TEST_USER_PASSWORD="testpassword123"

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

# Get auth token
get_auth_token() {
    print_test "Getting authentication token"

    login_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")

    if echo "$login_response" | grep -q "token"; then
        TOKEN=$(echo "$login_response" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
        if [ -n "$TOKEN" ]; then
            print_success "Authentication token obtained"
            return 0
        fi
    fi

    print_error "Failed to get authentication token"
    echo "Response: $login_response"
    return 1
}

# Test connection status endpoint
test_connection_status() {
    print_test "Testing connection status endpoint"

    response=$(curl -s "$BASE_URL/api/v1/connect/status" \
        -H "Authorization: Bearer ${TOKEN}")

    if echo "$response" | grep -q "vpn_tunnels\|dpi_bypasses\|active_connections"; then
        print_success "Connection status endpoint working"
        echo "Response: $response"
    else
        print_error "Connection status endpoint failed"
        echo "Response: $response"
    fi
}

# Test VPN-only connection
test_vpn_connection() {
    print_test "Testing VPN-only connection endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/connect/vpn" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "test-vpn-tunnel",
            "listen_port": 51820,
            "mtu": 1420,
            "auto_recovery": true
        }')

    if echo "$response" | grep -q "tunnel_id\|status"; then
        print_success "VPN-only connection endpoint working"
        echo "Response: $response"
    else
        print_error "VPN-only connection endpoint failed"
        echo "Response: $response"
    fi
}

# Test DPI-only connection
test_dpi_connection() {
    print_test "Testing DPI-only connection endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/connect/dpi" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "method": "shadowsocks",
            "name": "test-dpi-bypass",
            "remote_host": "example.com",
            "remote_port": 443,
            "password": "testpassword",
            "encryption": "aes-256-gcm"
        }')

    if echo "$response" | grep -q "bypass_id\|status"; then
        print_success "DPI-only connection endpoint working"
        echo "Response: $response"
    else
        print_error "DPI-only connection endpoint failed"
        echo "Response: $response"
    fi
}

# Test Shadowsocks connection
test_shadowsocks_connection() {
    print_test "Testing Shadowsocks connection endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/connect/shadowsocks" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "test-shadowsocks",
            "server_host": "example.com",
            "server_port": 443,
            "password": "testpassword",
            "encryption": "aes-256-gcm",
            "local_port": 1080
        }')

    if echo "$response" | grep -q "connection_id\|status"; then
        print_success "Shadowsocks connection endpoint working"
        echo "Response: $response"
    else
        print_error "Shadowsocks connection endpoint failed"
        echo "Response: $response"
    fi
}

# Test V2Ray connection
test_v2ray_connection() {
    print_test "Testing V2Ray connection endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/connect/v2ray" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "test-v2ray",
            "server_host": "example.com",
            "server_port": 443,
            "uuid": "550e8400-e29b-41d4-a716-446655440000",
            "alter_id": 0,
            "security": "auto",
            "network": "tcp"
        }')

    if echo "$response" | grep -q "connection_id\|status"; then
        print_success "V2Ray connection endpoint working"
        echo "Response: $response"
    else
        print_error "V2Ray connection endpoint failed"
        echo "Response: $response"
    fi
}

# Test Obfs4 connection
test_obfs4_connection() {
    print_test "Testing Obfs4 connection endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/connect/obfs4" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "test-obfs4",
            "bridge": "192.168.1.1:443",
            "cert": "test-certificate-data",
            "iat_mode": "0"
        }')

    if echo "$response" | grep -q "connection_id\|status"; then
        print_success "Obfs4 connection endpoint working"
        echo "Response: $response"
    else
        print_error "Obfs4 connection endpoint failed"
        echo "Response: $response"
    fi
}

# Test disconnect endpoint
test_disconnect() {
    print_test "Testing disconnect endpoint"

    response=$(curl -s -X POST "$BASE_URL/api/v1/disconnect" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "all": true
        }')

    if echo "$response" | grep -q "status\|disconnected"; then
        print_success "Disconnect endpoint working"
        echo "Response: $response"
    else
        print_error "Disconnect endpoint failed"
        echo "Response: $response"
    fi
}

# Test WebSocket endpoint
test_websocket() {
    print_test "Testing WebSocket endpoint"

    if command -v wscat >/dev/null 2>&1; then
        print_info "Testing WebSocket with wscat..."
        timeout 5 wscat -c "ws://localhost:8080/ws" -x '{"type":"ping"}' 2>/dev/null && \
            print_success "WebSocket endpoint working" || \
            print_error "WebSocket connection failed"
    else
        print_info "WebSocket endpoint registered (install 'npm install -g wscat' for full testing)"

        # Basic connectivity test
        response=$(curl -s -I "$BASE_URL/ws" 2>/dev/null | head -n1)
        if echo "$response" | grep -q "101\|400\|426"; then
            print_success "WebSocket endpoint accessible (upgrade required)"
        else
            print_error "WebSocket endpoint not accessible"
        fi
    fi
}

# Check gateway health
check_gateway_health() {
    print_test "Checking Gateway health"

    response=$(curl -s "$BASE_URL/health" 2>/dev/null)
    if echo "$response" | grep -q "ok\|healthy"; then
        print_success "Gateway is healthy"
    else
        print_error "Gateway is not healthy"
        echo "Response: $response"
        return 1
    fi
}

# Main test execution
main() {
    print_header "Testing New Connection Endpoints"

    # Check if Gateway is running
    if ! check_gateway_health; then
        print_error "Gateway is not running. Please start it first with: docker-compose up gateway"
        exit 1
    fi

    # Get authentication token
    if ! get_auth_token; then
        print_error "Cannot proceed without authentication token"
        exit 1
    fi

    # Test all new endpoints
    test_connection_status
    echo
    test_vpn_connection
    echo
    test_dpi_connection
    echo
    test_shadowsocks_connection
    echo
    test_v2ray_connection
    echo
    test_obfs4_connection
    echo
    test_disconnect
    echo
    test_websocket

    print_header "Testing Complete"
    print_info "All new connection endpoints have been tested."
    print_info "Note: Some tests may fail if underlying services (VPN Core, DPI Bypass) are not running."
}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Test script for new connection endpoints"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  --status-only  Test only status endpoint"
    echo "  --ws-only      Test only WebSocket endpoint"
    echo ""
    echo "Examples:"
    echo "  $0                # Run all tests"
    echo "  $0 --status-only  # Test only status endpoint"
    echo "  $0 --ws-only      # Test only WebSocket"
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    --status-only)
        check_gateway_health && get_auth_token && test_connection_status
        ;;
    --ws-only)
        check_gateway_health && test_websocket
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
