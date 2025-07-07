# Silence VPN Project

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.23-blue.svg" alt="Go Version">
  <img src="https://img.shields.io/badge/Docker-20.10+-blue.svg" alt="Docker Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Status-Development-yellow.svg" alt="Status">
</div>

A comprehensive VPN management system built with Go microservices, featuring authentication, analytics, server management, and advanced networking capabilities.

## 🚀 Quick Start

### Prerequisites
- Docker (20.10+) & Docker Compose (2.0+)
- Git
- 8GB+ RAM (16GB recommended)

### 1. Clone and Setup
```bash
git clone <repository-url>
cd silence
cp .env.example .env
```

### 2. Deploy
```bash
# Start all services using the management script (recommended)
./manage.sh start

# Or use Docker Compose directly
# docker-compose up -d
```

### 3. Verify
```bash
# Check the health of all services
./manage.sh health

# View the status of all services
./manage.sh status
```

## 🌐 Service Access Points

| Service | URL | Purpose |
|---|---|---|
| 🌐 **API Gateway** | http://localhost:8080 | Main API endpoint |
| 🔐 **Auth Service** | http://localhost:8081 | Authentication |
| 📊 **Analytics** | http://localhost:8082 | Metrics & reporting |
| 🔒 **DPI Bypass** | http://localhost:8083 | Traffic obfuscation |
| 🔑 **VPN Core** | http://localhost:8084 | VPN management |
| ⚙️ **Server Manager** | http://localhost:8085 | Infrastructure |
| 📢 **Notifications** | http://localhost:8087 | Event system |

### Management UIs
- **RabbitMQ**: http://localhost:15672 (admin/admin)
- **InfluxDB**: http://localhost:8086

## 🏗️ Architecture

The system is composed of 7 Go-based microservices and 4 infrastructure services (PostgreSQL, Redis, RabbitMQ, InfluxDB), all running in Docker containers.

```
Client → API Gateway (8080) → [Auth, Analytics, VPN Core, etc.]
                                  ↓
                      [PostgreSQL, Redis, RabbitMQ, InfluxDB]
```

## 🛠️ Management Script (`manage.sh`)

The `manage.sh` script is the primary tool for interacting with the system.

```bash
# Start/stop/restart all services
./manage.sh start
./manage.sh stop
./manage.sh restart

# View logs for all or a specific service
./manage.sh logs
./manage.sh logs gateway

# Check system health and status
./manage.sh health
./manage.sh status

# Clean up containers and volumes
./manage.sh clean-all

# Backup data
./manage.sh backup
```

## 🚢 Deployment

The project is designed for Docker Compose deployment. For production environments, consider the following:

### Security Checklist
- [ ] Change default passwords and secrets in `.env`.
- [ ] Configure SSL/TLS certificates for the Gateway.
- [ ] Set up firewall rules to limit port exposure.
- [ ] Configure a robust backup and recovery strategy.
- [ ] Set up external monitoring and alerting (e.g., Prometheus/Grafana).

### Production Configuration
Create a `docker-compose.prod.yml` to override development settings, such as removing exposed ports and adding resource limits.

## 🐛 Troubleshooting

| Problem | Solution |
|---|---|
| **Port Conflicts** | Check `sudo netstat -tlnp | grep <port>` and change the port in `.env`. |
| **Memory Issues** | Check `docker stats`. Increase memory allocated to Docker. |
| **DB Connection** | Check logs with `./manage.sh logs postgres`. Ensure the service is healthy. |
| **Service Fails** | Check logs with `./manage.sh logs <service-name>`. |

## 📚 Development & Contributing

- The project follows a standard Go microservice structure.
- New features should be developed in separate branches.
- Pull Requests are welcome. Please ensure tests and documentation are updated.

## 🎯 Roadmap

See the [ROADMAP.md](ROADMAP.md) file for a detailed list of completed and planned features.

## 📄 License

This project is licensed under the MIT License.
