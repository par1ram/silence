# Silence VPN Project Makefile
# ============================

# Variables
PROJECT_NAME := silence
SERVICES := gateway auth vpn-core dpi-bypass server-manager analytics notifications
DOCKER_COMPOSE := docker-compose -f docker-compose.unified.yml
DOCKER_COMPOSE_DEV := docker-compose -f docker-compose.unified.yml

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# Default target
.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Silence VPN - Available Commands:$(RESET)"
	@echo "=================================="
	@echo ""
	@echo "$(GREEN)🚀 Development Commands:$(RESET)"
	@echo "  make dev-single SERVICE=auth    - Start single service with hot reload"
	@echo "  make dev-all                    - Start all services without hot reload (efficient)"
	@echo "  make dev-production             - Start all services in production-like mode"
	@echo "  make setup                      - Complete project setup (first time)"
	@echo "  make stop                       - Stop all running services"
	@echo "  make restart                    - Restart all services"
	@echo ""
	@echo "$(GREEN)🔧 Individual Service Development:$(RESET)"
	@echo "  make dev-auth                   - Start Auth service with hot reload"
	@echo "  make dev-gateway                - Start Gateway service with hot reload"
	@echo "  make dev-vpn-core               - Start VPN Core service with hot reload"
	@echo "  make dev-dpi-bypass             - Start DPI Bypass service with hot reload"
	@echo "  make dev-server-manager         - Start Server Manager service with hot reload"
	@echo "  make dev-analytics              - Start Analytics service with hot reload"
	@echo "  make dev-notifications          - Start Notifications service with hot reload"
	@echo ""
	@echo "$(GREEN)🔨 Build Commands:$(RESET)"
	@echo "  make build                      - Build all services"
	@echo "  make build-SERVICE              - Build specific service (e.g., make build-gateway)"
	@echo ""
	@echo "$(GREEN)🧪 Test Commands:$(RESET)"
	@echo "  make test                       - Run all tests"
	@echo "  make test-SERVICE               - Test specific service"
	@echo "  make lint                       - Run linter on all services"
	@echo ""
	@echo "$(GREEN)🐳 Docker Commands:$(RESET)"
	@echo "  make docker-build               - Build Docker images"
	@echo "  make docker-up                  - Start containers"
	@echo "  make docker-down                - Stop containers"
	@echo ""
	@echo "$(GREEN)🏗️ Infrastructure Commands:$(RESET)"
	@echo "  make infra-up                   - Start infrastructure services"
	@echo "  make infra-down                 - Stop infrastructure services"
	@echo "  make db-create                  - Create databases"
	@echo "  make db-reset                   - Reset databases"
	@echo ""
	@echo "$(GREEN)🔍 Monitoring Commands:$(RESET)"
	@echo "  make health                     - Run health check"
	@echo "  make status                     - Show service status"
	@echo "  make logs                       - Show Docker logs"
	@echo ""
	@echo "$(GREEN)⚙️ Utility Commands:$(RESET)"
	@echo "  make deps                       - Install dependencies"
	@echo "  make clean                      - Clean build artifacts"
	@echo "  make configs                    - Setup configuration directories"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make dev-single SERVICE=auth    - Develop auth service with hot reload"
	@echo "  make dev-all                    - Run all services efficiently"
	@echo "  make dev-production             - Test in production-like environment"

# =============================================================================
# MAIN DEVELOPMENT COMMANDS
# =============================================================================

.PHONY: dev-single
dev-single: infra-up configs deps build-single ## Start single service with hot reload (usage: make dev-single SERVICE=auth)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Error: SERVICE parameter required$(RESET)"; \
		echo "$(YELLOW)Usage: make dev-single SERVICE=auth$(RESET)"; \
		echo "$(BLUE)Available services: $(SERVICES)$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)🔥 Starting $(SERVICE) service with hot reload...$(RESET)"
	@echo "$(YELLOW)Infrastructure services are running in Docker$(RESET)"
	@echo "$(BLUE)Press Ctrl+C to stop$(RESET)"
	@$(MAKE) dev-$(SERVICE)

.PHONY: dev-all
dev-all: infra-up configs deps build ## Start all services without hot reload (efficient mode)
	@echo "$(GREEN)🚀 Starting all services in efficient mode...$(RESET)"
	@echo "$(YELLOW)No hot reload for better performance and resource usage$(RESET)"
	@echo "$(BLUE)Use 'make dev-single SERVICE=name' for hot reload development$(RESET)"
	@echo "$(BLUE)Services will start on ports:$(RESET)"
	@echo "  Gateway: 8080, Auth: 8081, Analytics: 8082"
	@echo "  DPI Bypass: 8083, VPN Core: 8084, Server Manager: 8085"
	@echo "  Notifications: 8087"
	@echo ""
	@trap 'echo "$(RED)Stopping all services...$(RESET)"; pkill -f "silence-" || true; exit 0' INT TERM; \
	echo "$(GREEN)Starting Auth service...$(RESET)"; \
	(cd api/auth && \
	HTTP_PORT=8081 GRPC_PORT=:9081 DB_HOST=localhost DB_PORT=5432 DB_USER=postgres \
	DB_PASSWORD=password DB_NAME=silence_auth DB_SSLMODE=disable REDIS_HOST=localhost \
	REDIS_PORT=6379 JWT_SECRET=development-jwt-secret-key-change-this-in-production \
	MIGRATIONS_DIR=internal/adapters/database/migrations LOG_LEVEL=info \
	./bin/auth) & \
	echo "$(GREEN)Starting Gateway service...$(RESET)"; \
	(cd api/gateway && \
	HTTP_PORT=8080 AUTH_SERVICE_URL=http://localhost:8081 AUTH_GRPC_SERVICE_URL=localhost:9081 \
	ANALYTICS_SERVICE_URL=localhost:8082 SERVER_MANAGER_SERVICE_URL=localhost:8085 \
	DPI_BYPASS_SERVICE_URL=localhost:8083 VPN_CORE_SERVICE_URL=http://localhost:8084 \
	NOTIFICATIONS_SERVICE_URL=localhost:8087 JWT_SECRET=development-jwt-secret-key-change-this-in-production \
	LOG_LEVEL=info ./bin/gateway) & \
	echo "$(GREEN)Starting VPN Core service...$(RESET)"; \
	(cd rpc/vpn-core && \
	HTTP_PORT=8084 GRPC_PORT=:9084 DB_HOST=localhost DB_PORT=5432 DB_USER=postgres \
	DB_PASSWORD=password DB_NAME=silence_vpn DB_SSLMODE=disable REDIS_HOST=localhost \
	REDIS_PORT=6379 WIREGUARD_INTERFACE=wg0 WIREGUARD_LISTEN_PORT=51820 \
	MIGRATIONS_DIR=internal/adapters/database/migrations LOG_LEVEL=info \
	./bin/vpn-core) & \
	echo "$(GREEN)Starting DPI Bypass service...$(RESET)"; \
	(cd rpc/dpi-bypass && \
	HTTP_PORT=8083 GRPC_PORT=:9083 REDIS_HOST=localhost REDIS_PORT=6379 \
	LOG_LEVEL=info ./bin/dpi-bypass) & \
	echo "$(GREEN)Starting Server Manager service...$(RESET)"; \
	(cd rpc/server-manager && \
	HTTP_PORT=8085 GRPC_PORT=:9085 DB_HOST=localhost DB_PORT=5432 DB_USER=postgres \
	DB_PASSWORD=password DB_NAME=silence_server_manager DB_SSLMODE=disable \
	DOCKER_HOST=unix:///var/run/docker.sock MIGRATIONS_DIR=internal/adapters/database/migrations \
	LOG_LEVEL=info ./bin/server-manager) & \
	echo "$(GREEN)Starting Analytics service...$(RESET)"; \
	(cd rpc/analytics && \
	HTTP_PORT=8082 GRPC_PORT=:9082 PROMETHEUS_PORT=9091 INFLUXDB_URL=http://localhost:8086 \
	CLICKHOUSE_HOST=localhost CLICKHOUSE_PORT=9000 CLICKHOUSE_DB=silence_analytics \
	REDIS_HOST=localhost REDIS_PORT=6379 LOG_LEVEL=info ./bin/analytics) & \
	echo "$(GREEN)Starting Notifications service...$(RESET)"; \
	(cd rpc/notifications && \
	HTTP_PORT=8087 GRPC_PORT=:9087 RABBITMQ_URL=amqp://admin:admin@localhost:5672/ \
	REDIS_HOST=localhost REDIS_PORT=6379 ANALYTICS_URL=http://localhost:8082 \
	LOG_LEVEL=info ./bin/notifications) & \
	echo "$(GREEN)All services started!$(RESET)"; \
	echo "$(BLUE)Waiting for services to initialize...$(RESET)"; \
	sleep 5; \
	echo "$(CYAN)Services status:$(RESET)"; \
	make status; \
	echo "$(YELLOW)Press Ctrl+C to stop all services$(RESET)"; \
	wait

.PHONY: dev-production
dev-production: docker-build ## Start all services in production-like mode (Docker)
	@echo "$(GREEN)🏭 Starting all services in production-like mode...$(RESET)"
	@echo "$(YELLOW)Using Docker containers for all services$(RESET)"
	@echo "$(BLUE)This mode simulates production environment$(RESET)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)All services started in Docker!$(RESET)"
	@echo "$(BLUE)Waiting for services to initialize...$(RESET)"
	@sleep 10
	@echo "$(CYAN)Services status:$(RESET)"
	@make status
	@echo "$(YELLOW)Use 'make docker-logs' to view logs$(RESET)"
	@echo "$(YELLOW)Use 'make docker-down' to stop$(RESET)"

.PHONY: stop
stop: ## Stop all running services
	@echo "$(RED)🛑 Stopping all services...$(RESET)"
	@pkill -f "silence-" || true
	@$(DOCKER_COMPOSE) stop || true
	@echo "$(GREEN)All services stopped$(RESET)"

.PHONY: restart
restart: stop ## Restart all services
	@echo "$(YELLOW)🔄 Restarting services...$(RESET)"
	@sleep 3
	@$(MAKE) dev-all

# =============================================================================
# INDIVIDUAL SERVICE DEVELOPMENT (Hot Reload)
# =============================================================================

.PHONY: dev-auth
dev-auth: ## Start Auth service with hot reload
	@echo "$(GREEN)🔐 Starting Auth service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure infrastructure services are running: make infra-up$(RESET)"
	@cd api/auth && \
	HTTP_PORT=8081 \
	GRPC_PORT=:9081 \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=postgres \
	DB_PASSWORD=password \
	DB_NAME=silence_auth \
	DB_SSLMODE=disable \
	REDIS_HOST=localhost \
	REDIS_PORT=6379 \
	JWT_SECRET=development-jwt-secret-key-change-this-in-production \
	JWT_EXPIRATION=24h \
	JWT_REFRESH_EXPIRATION=168h \
	BCRYPT_COST=12 \
	INTERNAL_API_TOKEN=super-secret-internal-token \
	RATE_LIMIT_REQUESTS=100 \
	RATE_LIMIT_WINDOW=1m \
	MIGRATIONS_DIR=internal/adapters/database/migrations \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-gateway
dev-gateway: ## Start Gateway service with hot reload
	@echo "$(GREEN)🌐 Starting Gateway service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure other services are running$(RESET)"
	@cd api/gateway && \
	HTTP_PORT=8080 \
	AUTH_SERVICE_URL=http://localhost:8081 \
	AUTH_GRPC_SERVICE_URL=localhost:9081 \
	ANALYTICS_SERVICE_URL=localhost:8082 \
	SERVER_MANAGER_SERVICE_URL=localhost:8085 \
	DPI_BYPASS_SERVICE_URL=localhost:8083 \
	VPN_CORE_SERVICE_URL=http://localhost:8084 \
	NOTIFICATIONS_SERVICE_URL=localhost:8087 \
	JWT_SECRET=development-jwt-secret-key-change-this-in-production \
	INTERNAL_API_TOKEN=super-secret-internal-token \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-vpn-core
dev-vpn-core: ## Start VPN Core service with hot reload
	@echo "$(GREEN)🔐 Starting VPN Core service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure infrastructure services are running: make infra-up$(RESET)"
	@cd rpc/vpn-core && \
	HTTP_PORT=8084 \
	GRPC_PORT=:9084 \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=postgres \
	DB_PASSWORD=password \
	DB_NAME=silence_vpn \
	DB_SSLMODE=disable \
	REDIS_HOST=localhost \
	REDIS_PORT=6379 \
	WIREGUARD_INTERFACE=wg0 \
	WIREGUARD_LISTEN_PORT=51820 \
	WIREGUARD_PRIVATE_KEY_PATH=/tmp/wg_private.key \
	WIREGUARD_PUBLIC_KEY_PATH=/tmp/wg_public.key \
	MIGRATIONS_DIR=internal/adapters/database/migrations \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-dpi-bypass
dev-dpi-bypass: ## Start DPI Bypass service with hot reload
	@echo "$(GREEN)🚫 Starting DPI Bypass service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure Redis is running: make infra-up$(RESET)"
	@cd rpc/dpi-bypass && \
	HTTP_PORT=8083 \
	GRPC_PORT=:9083 \
	REDIS_HOST=localhost \
	REDIS_PORT=6379 \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-server-manager
dev-server-manager: ## Start Server Manager service with hot reload
	@echo "$(GREEN)🖥️ Starting Server Manager service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure infrastructure services are running: make infra-up$(RESET)"
	@cd rpc/server-manager && \
	HTTP_PORT=8085 \
	GRPC_PORT=:9085 \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=postgres \
	DB_PASSWORD=password \
	DB_NAME=silence_server_manager \
	DB_SSLMODE=disable \
	DOCKER_HOST=unix:///var/run/docker.sock \
	MIGRATIONS_DIR=internal/adapters/database/migrations \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-analytics
dev-analytics: ## Start Analytics service with hot reload
	@echo "$(GREEN)📊 Starting Analytics service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure InfluxDB and ClickHouse are running: make infra-up$(RESET)"
	@cd rpc/analytics && \
	HTTP_PORT=8082 \
	GRPC_PORT=:9082 \
	PROMETHEUS_PORT=9091 \
	INFLUXDB_URL=http://localhost:8086 \
	INFLUXDB_TOKEN=your-influxdb-token \
	INFLUXDB_ORG=silence \
	INFLUXDB_BUCKET=analytics \
	CLICKHOUSE_HOST=localhost \
	CLICKHOUSE_PORT=9000 \
	CLICKHOUSE_DB=silence_analytics \
	REDIS_HOST=localhost \
	REDIS_PORT=6379 \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

.PHONY: dev-notifications
dev-notifications: ## Start Notifications service with hot reload
	@echo "$(GREEN)📧 Starting Notifications service with hot reload...$(RESET)"
	@echo "$(YELLOW)Make sure RabbitMQ and Redis are running: make infra-up$(RESET)"
	@cd rpc/notifications && \
	HTTP_PORT=8087 \
	GRPC_PORT=:9087 \
	RABBITMQ_URL=amqp://admin:admin@localhost:5672/ \
	REDIS_HOST=localhost \
	REDIS_PORT=6379 \
	ANALYTICS_URL=http://localhost:8082 \
	SMTP_HOST=localhost \
	SMTP_PORT=1025 \
	SMTP_USERNAME= \
	SMTP_PASSWORD= \
	LOG_LEVEL=info \
	LOG_FORMAT=json \
	VERSION=1.0.0 \
	air

# =============================================================================
# BUILD COMMANDS
# =============================================================================

.PHONY: build
build: deps proto-generate ## Build all services
	@echo "$(GREEN)🔨 Building all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Building $$service...$(RESET)"; \
		$(MAKE) build-$$service; \
	done
	@echo "$(GREEN)✅ All services built successfully$(RESET)"

.PHONY: build-single
build-single: deps proto-generate ## Build single service (usage: make build-single SERVICE=auth)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Error: SERVICE parameter required$(RESET)"; \
		echo "$(YELLOW)Usage: make build-single SERVICE=auth$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)🔨 Building $(SERVICE) service...$(RESET)"
	@$(MAKE) build-$(SERVICE)

.PHONY: build-gateway
build-gateway: ## Build Gateway service
	@echo "$(BLUE)Building Gateway service...$(RESET)"
	@cd api/gateway && mkdir -p bin && go build -o bin/gateway ./cmd

.PHONY: build-auth
build-auth: ## Build Auth service
	@echo "$(BLUE)Building Auth service...$(RESET)"
	@cd api/auth && mkdir -p bin && go build -o bin/auth ./cmd

.PHONY: build-vpn-core
build-vpn-core: ## Build VPN Core service
	@echo "$(BLUE)Building VPN Core service...$(RESET)"
	@cd rpc/vpn-core && mkdir -p bin && go build -o bin/vpn-core ./cmd

.PHONY: build-dpi-bypass
build-dpi-bypass: ## Build DPI Bypass service
	@echo "$(BLUE)Building DPI Bypass service...$(RESET)"
	@cd rpc/dpi-bypass && mkdir -p bin && go build -o bin/dpi-bypass ./cmd

.PHONY: build-server-manager
build-server-manager: ## Build Server Manager service
	@echo "$(BLUE)Building Server Manager service...$(RESET)"
	@cd rpc/server-manager && mkdir -p bin && go build -o bin/server-manager ./cmd

.PHONY: build-analytics
build-analytics: ## Build Analytics service
	@echo "$(BLUE)Building Analytics service...$(RESET)"
	@cd rpc/analytics && mkdir -p bin && go build -o bin/analytics ./cmd

.PHONY: build-notifications
build-notifications: ## Build Notifications service
	@echo "$(BLUE)Building Notifications service...$(RESET)"
	@cd rpc/notifications && mkdir -p bin && go build -o bin/notifications ./cmd

# =============================================================================
# TEST COMMANDS
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)🧪 Running all tests...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Testing $$service...$(RESET)"; \
		$(MAKE) test-$$service; \
	done

.PHONY: test-gateway
test-gateway: ## Run Gateway tests
	@echo "$(BLUE)Testing Gateway service...$(RESET)"
	@cd api/gateway && go test -v ./...

.PHONY: test-auth
test-auth: ## Run Auth tests
	@echo "$(BLUE)Testing Auth service...$(RESET)"
	@cd api/auth && go test -v ./...

.PHONY: test-vpn-core
test-vpn-core: ## Run VPN Core tests
	@echo "$(BLUE)Testing VPN Core service...$(RESET)"
	@cd rpc/vpn-core && go test -v ./...

.PHONY: test-dpi-bypass
test-dpi-bypass: ## Run DPI Bypass tests
	@echo "$(BLUE)Testing DPI Bypass service...$(RESET)"
	@cd rpc/dpi-bypass && go test -v ./...

.PHONY: test-server-manager
test-server-manager: ## Run Server Manager tests
	@echo "$(BLUE)Testing Server Manager service...$(RESET)"
	@cd rpc/server-manager && go test -v ./...

.PHONY: test-analytics
test-analytics: ## Run Analytics tests
	@echo "$(BLUE)Testing Analytics service...$(RESET)"
	@cd rpc/analytics && go test -v ./...

.PHONY: test-notifications
test-notifications: ## Run Notifications tests
	@echo "$(BLUE)Testing Notifications service...$(RESET)"
	@cd rpc/notifications && go test -v ./...

# =============================================================================
# LINT COMMANDS
# =============================================================================

.PHONY: lint
lint: ## Run linter on all services
	@echo "$(GREEN)🔍 Running linter on all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(BLUE)Linting $$service...$(RESET)"; \
		$(MAKE) lint-$$service; \
	done

.PHONY: lint-gateway
lint-gateway: ## Run linter on Gateway service
	@cd api/gateway && golangci-lint run

.PHONY: lint-auth
lint-auth: ## Run linter on Auth service
	@cd api/auth && golangci-lint run

.PHONY: lint-vpn-core
lint-vpn-core: ## Run linter on VPN Core service
	@cd rpc/vpn-core && golangci-lint run

.PHONY: lint-dpi-bypass
lint-dpi-bypass: ## Run linter on DPI Bypass service
	@cd rpc/dpi-bypass && golangci-lint run

.PHONY: lint-server-manager
lint-server-manager: ## Run linter on Server Manager service
	@cd rpc/server-manager && golangci-lint run

.PHONY: lint-analytics
lint-analytics: ## Run linter on Analytics service
	@cd rpc/analytics && golangci-lint run

.PHONY: lint-notifications
lint-notifications: ## Run linter on Notifications service
	@cd rpc/notifications && golangci-lint run

# =============================================================================
# DOCKER COMMANDS
# =============================================================================

.PHONY: docker-build
docker-build: ## Build Docker images
	@echo "$(GREEN)🐳 Building Docker images...$(RESET)"
	@COMPOSE_DOCKER_CLI_BUILD=0 DOCKER_BUILDKIT=0 $(DOCKER_COMPOSE) build

.PHONY: docker-up
docker-up: docker-build ## Start all services in Docker
	@echo "$(GREEN)🚀 Starting all services in Docker...$(RESET)"
	@$(DOCKER_COMPOSE) up -d

.PHONY: docker-down
docker-down: ## Stop all Docker containers
	@echo "$(RED)🛑 Stopping Docker containers...$(RESET)"
	@$(DOCKER_COMPOSE) down

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	@echo "$(BLUE)📋 Showing Docker logs...$(RESET)"
	@$(DOCKER_COMPOSE) logs -f

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker containers

# =============================================================================
# INFRASTRUCTURE COMMANDS
# =============================================================================

.PHONY: infra-up
infra-up: ## Start infrastructure services
	@echo "$(GREEN)🏗️ Starting infrastructure services...$(RESET)"
	@$(DOCKER_COMPOSE) up -d postgres redis rabbitmq influxdb clickhouse mailhog
	@echo "$(YELLOW)⏳ Waiting for services to be ready...$(RESET)"
	@sleep 5
	@echo "$(GREEN)✅ Infrastructure services ready$(RESET)"
	@$(MAKE) db-create

.PHONY: infra-down
infra-down: ## Stop infrastructure services
	@echo "$(RED)🛑 Stopping infrastructure services...$(RESET)"
	@$(DOCKER_COMPOSE) down

.PHONY: infra-restart
infra-restart: infra-down infra-up ## Restart infrastructure services

.PHONY: infra-status
infra-status: ## Show infrastructure services status
	@echo "$(BLUE)📊 Infrastructure services status:$(RESET)"
	@$(DOCKER_COMPOSE) ps

# =============================================================================
# DATABASE COMMANDS
# =============================================================================

.PHONY: db-create
db-create: ## Create databases
	@echo "$(GREEN)🗃️ Creating databases...$(RESET)"
	@echo "$(BLUE)Waiting for PostgreSQL to be ready...$(RESET)"
	@bash -c 'for i in {1..30}; do docker exec silence_postgres pg_isready -U postgres && break || sleep 1; done'
	@echo "$(BLUE)Creating databases...$(RESET)"
	@docker exec silence_postgres createdb -U postgres silence_auth 2>/dev/null || echo "Database silence_auth already exists"
	@docker exec silence_postgres createdb -U postgres silence_vpn 2>/dev/null || echo "Database silence_vpn already exists"
	@docker exec silence_postgres createdb -U postgres silence_server_manager 2>/dev/null || echo "Database silence_server_manager already exists"
	@echo "$(GREEN)✅ Databases created$(RESET)"

.PHONY: db-drop
db-drop: ## Drop databases
	@echo "$(RED)🗑️ Dropping databases...$(RESET)"
	@docker exec silence_postgres dropdb -U postgres silence_auth 2>/dev/null || echo "Database silence_auth does not exist"
	@docker exec silence_postgres dropdb -U postgres silence_vpn 2>/dev/null || echo "Database silence_vpn does not exist"
	@docker exec silence_postgres dropdb -U postgres silence_server_manager 2>/dev/null || echo "Database silence_server_manager does not exist"
	@echo "$(GREEN)✅ Databases dropped$(RESET)"

.PHONY: db-reset
db-reset: db-drop db-create ## Reset databases

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(GREEN)🔄 Running database migrations...$(RESET)"
	@cd api/auth && go run cmd/migrate/main.go up || true
	@cd rpc/vpn-core && go run cmd/migrate/main.go up || true
	@cd rpc/server-manager && go run cmd/migrate/main.go up || true
	@echo "$(GREEN)✅ Migrations completed$(RESET)"

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

.PHONY: deps
deps: ## Install dependencies
	@echo "$(GREEN)📦 Installing dependencies...$(RESET)"
	@go mod download
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: proto-generate
proto-generate: ## Generate protobuf files
	@echo "$(GREEN)⚙️ Generating protobuf files...$(RESET)"
	@if [ -d "rpc/proto" ]; then \
		cd rpc/proto && \
		protoc --go_out=../shared/proto --go_opt=paths=source_relative \
		       --go-grpc_out=../shared/proto --go-grpc_opt=paths=source_relative \
		       *.proto; \
	fi

.PHONY: configs
configs: ## Setup configuration directories
	@echo "$(GREEN)⚙️ Setting up configuration directories...$(RESET)"
	@mkdir -p api/gateway/configs
	@mkdir -p api/auth/configs
	@mkdir -p rpc/vpn-core/configs
	@mkdir -p rpc/dpi-bypass/configs
	@mkdir -p rpc/server-manager/configs
	@mkdir -p rpc/analytics/configs
	@mkdir -p rpc/notifications/configs

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(RED)🧹 Cleaning build artifacts...$(RESET)"
	@find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "tmp" -type d -exec rm -rf {} + 2>/dev/null || true
	@go clean -cache
	@go clean -modcache
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

# =============================================================================
# MONITORING COMMANDS
# =============================================================================

.PHONY: health
health: ## Run comprehensive health check
	@echo "$(GREEN)🏥 Running health check...$(RESET)"
	@echo "$(BLUE)Checking infrastructure services...$(RESET)"
	@curl -s http://localhost:8080/health || echo "$(RED)❌ Gateway not responding$(RESET)"
	@curl -s http://localhost:8081/health || echo "$(RED)❌ Auth not responding$(RESET)"
	@curl -s http://localhost:8082/health || echo "$(RED)❌ Analytics not responding$(RESET)"
	@curl -s http://localhost:8083/health || echo "$(RED)❌ DPI Bypass not responding$(RESET)"
	@curl -s http://localhost:8084/health || echo "$(RED)❌ VPN Core not responding$(RESET)"
	@curl -s http://localhost:8085/health || echo "$(RED)❌ Server Manager not responding$(RESET)"
	@curl -s http://localhost:8087/health || echo "$(RED)❌ Notifications not responding$(RESET)"

.PHONY: health-quick
health-quick: ## Quick health check
	@echo "$(GREEN)🩺 Quick health check...$(RESET)"
	@curl -s http://localhost:8080/health | head -1 || echo "$(RED)❌ Gateway$(RESET)"
	@curl -s http://localhost:8081/health | head -1 || echo "$(RED)❌ Auth$(RESET)"

.PHONY: status
status: ## Show service status
	@echo "$(BLUE)📊 Service Status:$(RESET)"
	@echo "=================="
	@echo "$(CYAN)Application Services:$(RESET)"
	@nc -z localhost 8080 2>/dev/null && echo "$(GREEN)✅ Gateway (8080)$(RESET)" || echo "$(RED)❌ Gateway (8080)$(RESET)"
	@nc -z localhost 8081 2>/dev/null && echo "$(GREEN)✅ Auth (8081)$(RESET)" || echo "$(RED)❌ Auth (8081)$(RESET)"
	@nc -z localhost 8082 2>/dev/null && echo "$(GREEN)✅ Analytics (8082)$(RESET)" || echo "$(RED)❌ Analytics (8082)$(RESET)"
	@nc -z localhost 8083 2>/dev/null && echo "$(GREEN)✅ DPI Bypass (8083)$(RESET)" || echo "$(RED)❌ DPI Bypass (8083)$(RESET)"
	@nc -z localhost 8084 2>/dev/null && echo "$(GREEN)✅ VPN Core (8084)$(RESET)" || echo "$(RED)❌ VPN Core (8084)$(RESET)"
	@nc -z localhost 8085 2>/dev/null && echo "$(GREEN)✅ Server Manager (8085)$(RESET)" || echo "$(RED)❌ Server Manager (8085)$(RESET)"
	@nc -z localhost 8087 2>/dev/null && echo "$(GREEN)✅ Notifications (8087)$(RESET)" || echo "$(RED)❌ Notifications (8087)$(RESET)"
	@echo "$(CYAN)Infrastructure Services:$(RESET)"
	@nc -z localhost 5432 2>/dev/null && echo "$(GREEN)✅ PostgreSQL (5432)$(RESET)" || echo "$(RED)❌ PostgreSQL (5432)$(RESET)"
	@nc -z localhost 6379 2>/dev/null && echo "$(GREEN)✅ Redis (6379)$(RESET)" || echo "$(RED)❌ Redis (6379)$(RESET)"
	@nc -z localhost 5672 2>/dev/null && echo "$(GREEN)✅ RabbitMQ (5672)$(RESET)" || echo "$(RED)❌ RabbitMQ (5672)$(RESET)"
	@nc -z localhost 8086 2>/dev/null && echo "$(GREEN)✅ InfluxDB (8086)$(RESET)" || echo "$(RED)❌ InfluxDB (8086)$(RESET)"
	@nc -z localhost 9000 2>/dev/null && echo "$(GREEN)✅ ClickHouse (9000)$(RESET)" || echo "$(RED)❌ ClickHouse (9000)$(RESET)"
	@nc -z localhost 1025 2>/dev/null && echo "$(GREEN)✅ MailHog (1025)$(RESET)" || echo "$(RED)❌ MailHog (1025)$(RESET)"

.PHONY: logs
logs: ## Show Docker logs
	@echo "$(BLUE)📋 Following Docker logs...$(RESET)"
	@echo "Press Ctrl+C to stop"
	@$(DOCKER_COMPOSE_DEV) logs -f

.PHONY: logs-service
logs-service: ## Show logs for specific service (usage: make logs-service SERVICE=postgres)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Error: SERVICE parameter required$(RESET)"; \
		echo "$(YELLOW)Usage: make logs-service SERVICE=postgres$(RESET)"; \
		exit 1; \
	fi
	@$(DOCKER_COMPOSE_DEV) logs -f $(SERVICE)

# =============================================================================
# DEVELOPMENT WORKFLOW
# =============================================================================

.PHONY: setup
setup: ## Complete project setup (first time)
	@echo "$(GREEN)🚀 Setting up Silence VPN project...$(RESET)"
	@echo "$(BLUE)This will take a few minutes...$(RESET)"
	@$(MAKE) deps
	@$(MAKE) configs
	@$(MAKE) proto-generate
	@$(MAKE) infra-up
	@$(MAKE) build
	@echo "$(GREEN)✅ Project setup completed!$(RESET)"
	@echo "$(CYAN)🎉 Available commands:$(RESET)"
	@echo "  make dev-single SERVICE=auth  - Start single service development"
	@echo "  make dev-all                  - Start all services (efficient)"
	@echo "  make dev-production           - Start production-like environment"

.PHONY: reset
reset: ## Reset project to clean state
	@echo "$(YELLOW)🔄 Resetting project...$(RESET)"
	@$(MAKE) stop
	@$(MAKE) clean
	@$(MAKE) infra-down
	@$(MAKE) db-reset
	@echo "$(GREEN)✅ Project reset completed$(RESET)"

# =============================================================================
# TESTING AND VALIDATION
# =============================================================================

.PHONY: test-endpoints
test-endpoints: ## Test API endpoints
	@echo "$(GREEN)🧪 Testing API endpoints...$(RESET)"
	@echo "$(BLUE)Testing Gateway health...$(RESET)"
	@curl -f http://localhost:8080/health 2>/dev/null && echo "$(GREEN)✅ Gateway health OK$(RESET)" || echo "$(RED)❌ Gateway health failed$(RESET)"
	@echo "$(BLUE)Testing Auth health...$(RESET)"
	@curl -f http://localhost:8081/health 2>/dev/null && echo "$(GREEN)✅ Auth health OK$(RESET)" || echo "$(RED)❌ Auth health failed$(RESET)"
	@echo "$(BLUE)Testing VPN Core health...$(RESET)"
	@curl -f http://localhost:8084/health 2>/dev/null && echo "$(GREEN)✅ VPN Core health OK$(RESET)" || echo "$(RED)❌ VPN Core health failed$(RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)🧪 Running integration tests...$(RESET)"
	@echo "$(BLUE)Starting test environment...$(RESET)"
	@$(MAKE) infra-up
	@$(MAKE) build
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v -tags=integration ./tests/integration/...
	@echo "$(GREEN)✅ Integration tests completed$(RESET)"

# =============================================================================
# MONITORING AND DEBUGGING
# =============================================================================

.PHONY: debug
debug: ## Show debugging information
	@echo "$(BLUE)🔍 Debug Information:$(RESET)"
	@echo "===================="
	@echo "Go version: $(shell go version)"
	@echo "Docker version: $(shell docker --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker Compose version: $(shell docker-compose --version 2>/dev/null || echo 'Not installed')"
	@echo "Air version: $(shell air -v 2>/dev/null || echo 'Not installed')"
	@echo "Project directory: $(shell pwd)"
	@echo ""
	@echo "$(CYAN)Running processes:$(RESET)"
	@ps aux | grep -E "(air|silence-)" | grep -v grep || echo "No services running"
	@echo ""
	@echo "$(CYAN)Docker containers:$(RESET)"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep silence || echo "No containers running"

.PHONY: ps
ps: ## Show running processes
	@echo "$(BLUE)🔍 Running Processes:$(RESET)"
	@ps aux | grep -E "(air|silence-)" | grep -v grep || echo "No services running"

.PHONY: ports
ports: ## Show port usage
	@echo "$(BLUE)🔍 Port Usage:$(RESET)"
	@echo "$(CYAN)Expected ports:$(RESET)"
	@echo "  8080  - Gateway"
	@echo "  8081  - Auth"
	@echo "  8082  - Analytics"
	@echo "  8083  - DPI Bypass"
	@echo "  8084  - VPN Core"
	@echo "  8085  - Server Manager"
	@echo "  8087  - Notifications"
	@echo "  5432  - PostgreSQL"
	@echo "  6379  - Redis"
	@echo "  5672  - RabbitMQ"
	@echo "  8086  - InfluxDB"
	@echo "  9000  - ClickHouse"
	@echo "  1025  - MailHog SMTP"
	@echo "  8025  - MailHog Web UI"
	@echo ""
	@echo "$(CYAN)Currently listening:$(RESET)"
	@netstat -tuln | grep -E "(8080|8081|8082|8083|8084|8085|8087|5432|6379|5672|8086|9000|1025|8025)" || echo "No services listening on expected ports"

.PHONY: env
env: ## Show environment variables for development
	@echo "$(BLUE)🔍 Development Environment Variables:$(RESET)"
	@echo "====================================="
	@echo "$(CYAN)Database:$(RESET)"
	@echo "  DB_HOST=localhost"
	@echo "  DB_PORT=5432"
	@echo "  DB_USER=postgres"
	@echo "  DB_PASSWORD=password"
	@echo "  DB_SSLMODE=disable"
	@echo ""
	@echo "$(CYAN)Redis:$(RESET)"
	@echo "  REDIS_HOST=localhost"
	@echo "  REDIS_PORT=6379"
	@echo ""
	@echo "$(CYAN)RabbitMQ:$(RESET)"
	@echo "  RABBITMQ_URL=amqp://admin:admin@localhost:5672/"
	@echo ""
	@echo "$(CYAN)JWT:$(RESET)"
	@echo "  JWT_SECRET=development-jwt-secret-key-change-this-in-production"
	@echo ""
	@echo "$(CYAN)Log Level:$(RESET)"
	@echo "  LOG_LEVEL=info"
	@echo "  LOG_FORMAT=json"

# =============================================================================
# PERFORMANCE AND OPTIMIZATION
# =============================================================================

.PHONY: profile
profile: ## Profile running services
	@echo "$(GREEN)📊 Profiling services...$(RESET)"
	@echo "$(BLUE)Gateway profile:$(RESET)"
	@curl -s http://localhost:8080/debug/pprof/goroutine?debug=1 | head -20 || echo "Profile not available"
	@echo "$(BLUE)Auth profile:$(RESET)"
	@curl -s http://localhost:8081/debug/pprof/goroutine?debug=1 | head -20 || echo "Profile not available"

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(GREEN)🏃 Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...

# =============================================================================
# QUALITY ASSURANCE
# =============================================================================

.PHONY: security
security: ## Run security checks
	@echo "$(GREEN)🔒 Running security checks...$(RESET)"
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...

.PHONY: coverage
coverage: ## Run test coverage
	@echo "$(GREEN)📊 Running test coverage...$(RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

# =============================================================================
# DOCUMENTATION
# =============================================================================

.PHONY: docs
docs: ## Generate documentation
	@echo "$(GREEN)📚 Generating documentation...$(RESET)"
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "$(BLUE)Starting documentation server...$(RESET)"
	@echo "$(YELLOW)Visit http://localhost:6060 for documentation$(RESET)"
	@godoc -http=:6060

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "$(GREEN)📖 Generating Swagger documentation...$(RESET)"
	@go install github.com/swaggo/swag/cmd/swag@latest
	@cd api/gateway && swag init
	@cd api/auth && swag init
	@echo "$(GREEN)✅ Swagger documentation generated$(RESET)"

# Make sure to use tabs instead of spaces for indentation
.DEFAULT_GOAL := help
