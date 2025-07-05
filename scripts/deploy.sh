#!/bin/bash

# Silence VPN - Automated Deployment Script
# Usage: ./scripts/deploy.sh [environment] [version]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT=${1:-staging}
VERSION=${2:-latest}
NAMESPACE="silence-${ENVIRONMENT}"
REGISTRY="ghcr.io"

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(staging|production)$ ]]; then
    echo -e "${RED}Error: Environment must be 'staging' or 'production'${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Deploying Silence VPN to ${ENVIRONMENT} environment${NC}"
echo -e "${BLUE}Version: ${VERSION}${NC}"
echo -e "${BLUE}Namespace: ${NAMESPACE}${NC}"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo -e "${RED}Error: helm is not installed${NC}"
    exit 1
fi

# Function to check if namespace exists
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        echo -e "${YELLOW}Creating namespace ${NAMESPACE}...${NC}"
        kubectl create namespace "$NAMESPACE"
    fi
}

# Function to create secrets
create_secrets() {
    echo -e "${BLUE}🔐 Creating secrets...${NC}"
    
    # Check if secrets already exist
    if kubectl get secret silence-secrets -n "$NAMESPACE" &> /dev/null; then
        echo -e "${YELLOW}Secrets already exist, skipping...${NC}"
        return
    fi
    
    # Generate random passwords
    DB_PASSWORD=$(openssl rand -base64 32)
    JWT_SECRET=$(openssl rand -base64 64)
    INFLUXDB_PASSWORD=$(openssl rand -base64 32)
    INFLUXDB_TOKEN=$(openssl rand -base64 32)
    RABBITMQ_USER="silence"
    RABBITMQ_PASSWORD=$(openssl rand -base64 32)
    GRAFANA_PASSWORD=$(openssl rand -base64 32)
    
    # Create secrets
    kubectl create secret generic silence-secrets \
        --from-literal=db-user=silence \
        --from-literal=db-password="$DB_PASSWORD" \
        --from-literal=jwt-secret="$JWT_SECRET" \
        --from-literal=influxdb-password="$INFLUXDB_PASSWORD" \
        --from-literal=influxdb-token="$INFLUXDB_TOKEN" \
        --from-literal=rabbitmq-user="$RABBITMQ_USER" \
        --from-literal=rabbitmq-password="$RABBITMQ_PASSWORD" \
        --from-literal=grafana-password="$GRAFANA_PASSWORD" \
        -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ Secrets created successfully${NC}"
}

# Function to create image pull secret
create_image_pull_secret() {
    echo -e "${BLUE}🔑 Creating image pull secret...${NC}"
    
    if kubectl get secret ghcr-secret -n "$NAMESPACE" &> /dev/null; then
        echo -e "${YELLOW}Image pull secret already exists, skipping...${NC}"
        return
    fi
    
    # Create image pull secret for GitHub Container Registry
    kubectl create secret docker-registry ghcr-secret \
        --docker-server=ghcr.io \
        --docker-username="$GITHUB_USERNAME" \
        --docker-password="$GITHUB_TOKEN" \
        -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ Image pull secret created successfully${NC}"
}

# Function to deploy database services
deploy_databases() {
    echo -e "${BLUE}🗄️ Deploying database services...${NC}"
    
    # Deploy PostgreSQL
    kubectl apply -f deployments/kubernetes/postgres.yaml -n "$NAMESPACE"
    
    # Deploy Redis
    kubectl apply -f deployments/kubernetes/redis.yaml -n "$NAMESPACE"
    
    # Deploy InfluxDB
    kubectl apply -f deployments/kubernetes/influxdb.yaml -n "$NAMESPACE"
    
    # Deploy RabbitMQ
    kubectl apply -f deployments/kubernetes/rabbitmq.yaml -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ Database services deployed${NC}"
}

# Function to deploy application services
deploy_services() {
    echo -e "${BLUE}🚀 Deploying application services...${NC}"
    
    # Update values file with current version
    sed "s/{{ .Values.imageTag | default .Chart.AppVersion }}/$VERSION/g" \
        deployments/kubernetes/ci-cd-values.yaml > /tmp/silence-values.yaml
    
    # Deploy using Helm
    helm upgrade --install silence-vpn ./deployments/helm/silence-vpn \
        --values /tmp/silence-values.yaml \
        --set global.environment="$ENVIRONMENT" \
        --set global.imageRegistry="$REGISTRY" \
        --namespace "$NAMESPACE" \
        --create-namespace \
        --wait \
        --timeout 10m
    
    echo -e "${GREEN}✅ Application services deployed${NC}"
}

# Function to deploy monitoring
deploy_monitoring() {
    echo -e "${BLUE}📊 Deploying monitoring stack...${NC}"
    
    # Deploy Prometheus
    kubectl apply -f deployments/kubernetes/monitoring/prometheus.yaml -n "$NAMESPACE"
    
    # Deploy Grafana
    kubectl apply -f deployments/kubernetes/monitoring/grafana.yaml -n "$NAMESPACE"
    
    # Deploy ServiceMonitor for our services
    kubectl apply -f deployments/kubernetes/monitoring/servicemonitors.yaml -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ Monitoring stack deployed${NC}"
}

# Function to deploy ingress
deploy_ingress() {
    echo -e "${BLUE}🌐 Deploying ingress...${NC}"
    
    kubectl apply -f deployments/kubernetes/ingress.yaml -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ Ingress deployed${NC}"
}

# Function to wait for services to be ready
wait_for_services() {
    echo -e "${BLUE}⏳ Waiting for services to be ready...${NC}"
    
    # Wait for all deployments to be ready
    kubectl wait --for=condition=available --timeout=300s deployment --all -n "$NAMESPACE"
    
    # Wait for all pods to be ready
    kubectl wait --for=condition=ready --timeout=300s pod --all -n "$NAMESPACE"
    
    echo -e "${GREEN}✅ All services are ready${NC}"
}

# Function to show deployment status
show_status() {
    echo -e "${BLUE}📋 Deployment Status:${NC}"
    
    echo -e "\n${YELLOW}Pods:${NC}"
    kubectl get pods -n "$NAMESPACE"
    
    echo -e "\n${YELLOW}Services:${NC}"
    kubectl get services -n "$NAMESPACE"
    
    echo -e "\n${YELLOW}Deployments:${NC}"
    kubectl get deployments -n "$NAMESPACE"
    
    echo -e "\n${YELLOW}Ingress:${NC}"
    kubectl get ingress -n "$NAMESPACE"
    
    # Show service URLs
    echo -e "\n${GREEN}🌐 Service URLs:${NC}"
    if [ "$ENVIRONMENT" = "production" ]; then
        echo "API Gateway: https://api.silence.local"
        echo "Auth Service: https://auth.silence.local"
        echo "Grafana: https://grafana.silence.local"
    else
        echo "API Gateway: http://localhost:8080"
        echo "Auth Service: http://localhost:8081"
        echo "Grafana: http://localhost:3000"
    fi
}

# Function to run health checks
health_check() {
    echo -e "${BLUE}🏥 Running health checks...${NC}"
    
    # Check if all pods are running
    FAILED_PODS=$(kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running --no-headers | wc -l)
    
    if [ "$FAILED_PODS" -gt 0 ]; then
        echo -e "${RED}❌ Found $FAILED_PODS failed pods${NC}"
        kubectl get pods -n "$NAMESPACE" --field-selector=status.phase!=Running
        return 1
    fi
    
    # Check service endpoints
    SERVICES=("auth-service" "gateway-service" "vpn-core-service" "dpi-bypass-service" "server-manager-service" "analytics-service" "notifications-service")
    
    for service in "${SERVICES[@]}"; do
        if kubectl get endpoints "$service" -n "$NAMESPACE" &> /dev/null; then
            ENDPOINTS=$(kubectl get endpoints "$service" -n "$NAMESPACE" -o jsonpath='{.subsets[0].addresses}' | jq length 2>/dev/null || echo "0")
            if [ "$ENDPOINTS" -gt 0 ]; then
                echo -e "${GREEN}✅ $service: $ENDPOINTS endpoints${NC}"
            else
                echo -e "${RED}❌ $service: no endpoints${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️ $service: not found${NC}"
        fi
    done
    
    echo -e "${GREEN}✅ Health checks completed${NC}"
}

# Main deployment flow
main() {
    echo -e "${BLUE}Starting deployment process...${NC}"
    
    # Check prerequisites
    check_namespace
    
    # Create secrets
    create_secrets
    
    # Create image pull secret if credentials are provided
    if [ -n "$GITHUB_USERNAME" ] && [ -n "$GITHUB_TOKEN" ]; then
        create_image_pull_secret
    else
        echo -e "${YELLOW}⚠️ Skipping image pull secret creation (GITHUB_USERNAME/GITHUB_TOKEN not set)${NC}"
    fi
    
    # Deploy infrastructure
    deploy_databases
    
    # Deploy application
    deploy_services
    
    # Deploy monitoring
    deploy_monitoring
    
    # Deploy ingress
    deploy_ingress
    
    # Wait for services
    wait_for_services
    
    # Health checks
    health_check
    
    # Show status
    show_status
    
    echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
}

# Handle script arguments
case "${1:-}" in
    "staging"|"production")
        main
        ;;
    "rollback")
        echo -e "${YELLOW}🔄 Rolling back deployment...${NC}"
        helm rollback silence-vpn -n "$NAMESPACE"
        echo -e "${GREEN}✅ Rollback completed${NC}"
        ;;
    "status")
        show_status
        ;;
    "health")
        health_check
        ;;
    "logs")
        SERVICE=${2:-}
        if [ -n "$SERVICE" ]; then
            kubectl logs -f deployment/"$SERVICE" -n "$NAMESPACE"
        else
            echo -e "${RED}Error: Please specify service name${NC}"
            echo "Usage: $0 logs [service-name]"
        fi
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [environment] [version]"
        echo ""
        echo "Environments:"
        echo "  staging     - Deploy to staging environment"
        echo "  production  - Deploy to production environment"
        echo ""
        echo "Commands:"
        echo "  rollback    - Rollback to previous version"
        echo "  status      - Show deployment status"
        echo "  health      - Run health checks"
        echo "  logs [svc]  - Show logs for specific service"
        echo "  help        - Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  GITHUB_USERNAME - GitHub username for image pull"
        echo "  GITHUB_TOKEN    - GitHub token for image pull"
        ;;
    *)
        echo -e "${RED}Error: Invalid environment or command${NC}"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac 