#!/bin/bash

# Build script for Silence VPN services without buildx
# This script builds all Docker images using standard docker build command

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}🐳 Building Docker images for Silence VPN services...${NC}"

# Function to build a service
build_service() {
    local service_name="$1"
    local dockerfile_path="$2"
    local image_name="silence-${service_name}"

    echo -e "${YELLOW}📦 Building ${service_name} service...${NC}"

    if docker build -t "$image_name" -f "$dockerfile_path" "$PROJECT_ROOT"; then
        echo -e "${GREEN}✅ Successfully built ${service_name}${NC}"
    else
        echo -e "${RED}❌ Failed to build ${service_name}${NC}"
        exit 1
    fi
}

# Build all services
echo -e "${BLUE}Starting build process...${NC}"

# Build Auth Service
build_service "auth" "api/auth/Dockerfile"

# Build Gateway Service
build_service "gateway" "api/gateway/Dockerfile"

# Build Analytics Service
build_service "analytics" "rpc/analytics/Dockerfile"

# Build Server Manager Service
build_service "server-manager" "rpc/server-manager/Dockerfile"

# Build DPI Bypass Service
build_service "dpi-bypass" "rpc/dpi-bypass/Dockerfile"

# Build VPN Core Service
build_service "vpn-core" "rpc/vpn-core/Dockerfile"

# Build Notifications Service
build_service "notifications" "rpc/notifications/Dockerfile"

echo -e "${GREEN}🎉 All Docker images built successfully!${NC}"

# List built images
echo -e "${BLUE}📋 Built images:${NC}"
docker images | grep "silence-" | awk '{print $1 ":" $2 " (" $7 " " $8 ")"}'

echo -e "${GREEN}✅ Build process completed successfully!${NC}"
