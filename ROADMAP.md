# Roadmap

This document outlines the development roadmap for the Silence VPN project. 

## Completed

- [x] **Infrastructure**: Go workspace, microservice structure, hot-reload, DI container, logger, graceful shutdown.
- [x] **Authentication**: Auth Service, JWT, registration/login, user management (CRUD, roles, status), admin endpoints, password reset.
- [x] **API Gateway**: Gateway Service, request routing, rate limiting, core feature integration (`/api/v1/connect`).
- [x] **VPN Core**: VPN Core Service, WireGuard integration, tunnel and peer management, gRPC service, auto-recovery.
- [x] **DPI Bypass**: DPI Bypass Service, Shadowsocks, V2Ray, obfs4 support, REST API for management.
- [x] **Server Manager**: Full implementation for Docker and Kubernetes, including deployment, monitoring, scaling, updates, and backups.
- [x] **Analytics**: Analytics Service, metrics collection, InfluxDB & Redis integration, AlertService, DashboardRepository.
- [x] **Notifications**: Notification Service with stub adapters for Email, Push, SMS, Telegram, Slack, Webhooks; RabbitMQ integration.
- [x] **Deployment**: CI/CD pipeline (GitHub Actions), Kubernetes manifests, security scanning, automated releases.

## In Progress / To Do

### Core Features
- [ ] **API Gateway**: Add specialized connection endpoints for better UX and management:
  - [ ] `POST /api/v1/connect/vpn` - VPN-only connection
  - [ ] `POST /api/v1/connect/dpi` - DPI Bypass-only connection
  - [ ] `POST /api/v1/connect/shadowsocks` - Shadowsocks-specific connection
  - [ ] `POST /api/v1/connect/v2ray` - V2Ray-specific connection
  - [ ] `POST /api/v1/connect/obfs4` - Obfs4-specific connection
  - [ ] `POST /api/v1/disconnect` - Disconnect active connections
  - [ ] `GET /api/v1/connect/status` - Real-time connection status
- [ ] **Real-time Features**: Implement WebSocket support for frontend:
  - [ ] WebSocket endpoint for real-time connection status updates
  - [ ] Live metrics streaming (bandwidth, latency, packet loss)
  - [ ] Instant notifications for connection events and alerts
  - [ ] Connection health monitoring with auto-reconnect
- [ ] **Authentication**: Implement 2FA and OAuth2 integration.
- [ ] **DPI Bypass**: Implement automatic optimal method selection and ML-based effectiveness analysis.
- [ ] **Server Manager**: Integrate with external providers (AWS, GCP) and add webhook notifications.
- [ ] **Analytics**: Implement remaining data collectors, alert history, and notification triggers.
- [ ] **Notifications**: Implement real delivery adapters (SMTP, Twilio, etc.), retry logic, and user preferences.

### Testing
- [ ] Enhance Unit and Integration test coverage.
- [ ] Implement E2E and load testing.
- [ ] Automate testing for DPI circumvention.

### Documentation
- [ ] Create comprehensive API documentation (Swagger/OpenAPI).
- [ ] Write a detailed user guide and troubleshooting manual.

### Client Applications
- [ ] **Frontend Web Interface**: Complete web-based management interface:
  - [ ] Connection management dashboard with real-time status
  - [ ] Quick connect components for different VPN methods
  - [ ] Analytics and metrics visualization
  - [ ] User settings and preferences management
  - [ ] Connection history and logs viewer
- [ ] Develop mobile applications (iOS/Android).
- [ ] Develop desktop applications (Windows/macOS/Linux).

### General
- [ ] Centralize the `inappropriate ioctl for device` logger error filtering.
