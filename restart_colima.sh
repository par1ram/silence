#!/bin/bash

set -e

echo "[1/6] Останавливаю Colima..."
colima stop || true

sleep 2

echo "[2/6] Проверяю, нет ли ssh-процессов на 8086..."
PID=$(lsof -ti :8086 || true)
if [ -n "$PID" ]; then
  echo "Убиваю процесс $PID, занимающий порт 8086"
  kill $PID || true
fi

sleep 1

echo "[3/6] Запускаю Colima..."
colima start

sleep 2

echo "[4/6] Проверяю статус Docker..."
docker info && echo "Docker работает!" || { echo "Docker не запущен!"; exit 1; }

sleep 1

echo "[5/6] Запускаю все контейнеры (task docker:up)..."
task docker:up

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


echo "[6/6] Проверяю health всех сервисов:"
for svc in "${SERVICES[@]}"; do
  name="${svc%%|*}"
  url="${svc##*|}"
  echo -n "  $name: "
  if curl -s --max-time 3 "$url" | grep -q 'ok\|"ok"\|status'; then
    echo "OK"
  else
    echo "FAIL (см. $url)"
  fi
done 