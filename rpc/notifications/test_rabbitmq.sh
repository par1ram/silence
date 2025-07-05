#!/bin/bash

# Тестовый скрипт для отправки события через RabbitMQ

echo "Отправка тестового события через RabbitMQ..."

# Создаем тестовое событие
cat > /tmp/test_event.json << EOF
{
  "id": "test-event-$(date +%s)",
  "type": "alert",
  "priority": "high",
  "title": "Тестовое событие из RabbitMQ",
  "message": "Это тестовое событие, отправленное через RabbitMQ",
  "source": "test-script",
  "recipients": ["test@example.com"],
  "channels": ["email"],
  "metadata": {
    "test": true,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF

# Отправляем событие через RabbitMQ Management API
curl -X POST http://localhost:15672/api/exchanges/%2F/notifications/publish \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "properties": {
      "content_type": "application/json",
      "delivery_mode": 2
    },
    "routing_key": "notifications.alert",
    "payload": "'$(cat /tmp/test_event.json | tr -d '\n' | sed 's/"/\\"/g')'",
    "payload_encoding": "string"
  }'

echo ""
echo "Событие отправлено. Проверьте логи notifications сервиса."

# Очищаем временный файл
rm -f /tmp/test_event.json 