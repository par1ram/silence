#!/bin/bash
set -euo pipefail

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

step() {
  echo -e "${GREEN}[step]${NC} $1"
}

fail() {
  echo -e "${RED}[fail]${NC} $1" >&2
  exit 1
}

step "Останавливаю Colima..."
colima stop || true
sleep 2

step "Освобождаю порт 8086 (если занят)..."
PID=$(lsof -ti :8086 || true)
if [ -n "$PID" ]; then
  echo "Убиваю процесс $PID, занимающий порт 8086"
  kill $PID || true
fi
sleep 1

step "Запускаю Colima..."
colima start
sleep 2

step "Проверяю статус Docker..."
docker info >/dev/null 2>&1 || fail "Docker не запущен!"
echo "Docker работает!"
sleep 1

step "Запускаю контейнеры через docker-compose..."
docker-compose up -d || fail "Не удалось поднять контейнеры!"
sleep 5

# Health-check endpoints
SERVICES=(
  "gateway|http://localhost:8080/health"
  "auth|http://localhost:8081/health"
  "vpn-core|http://localhost:8082/health"
  "dpi-bypass|http://localhost:8083/health"
  "analytics|http://localhost:8084/health"
  "server-manager|http://localhost:8085/health"
  "notifications|http://localhost:8086/healthz"
)

step "Проверяю health всех сервисов:"
for svc in "${SERVICES[@]}"; do
  name="${svc%%|*}"
  url="${svc##*|}"
  printf "  %-15s: " "$name"
  if curl -s --max-time 4 "$url" | grep -q 'ok\|"ok"\|status'; then
    echo -e "${GREEN}OK${NC}"
  else
    echo -e "${RED}FAIL${NC} (см. $url)"
  fi
done 