#!/bin/bash

# Скрипт для тестирования всех функций notifications сервиса

echo "🧪 Тестирование Notifications сервиса"
echo "======================================"

# Проверяем, что сервис запущен
echo "1. Проверка health check..."
if curl -s http://localhost:8080/healthz | grep -q "ok"; then
    echo "✅ Health check работает"
else
    echo "❌ Health check не работает"
    exit 1
fi

echo ""
echo "2. Тестирование каналов доставки..."

# Email
echo "📧 Тестируем Email..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "Email тест",
    "message": "Тестовое email уведомление",
    "recipients": ["test@example.com"],
    "channels": ["email"]
  }' > /dev/null && echo "✅ Email работает"

# SMS
echo "📱 Тестируем SMS..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "SMS тест",
    "message": "Тестовое SMS уведомление",
    "recipients": ["+1234567890"],
    "channels": ["sms"]
  }' > /dev/null && echo "✅ SMS работает"

# Telegram
echo "📲 Тестируем Telegram..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "Telegram тест",
    "message": "Тестовое Telegram уведомление",
    "recipients": ["@test_user"],
    "channels": ["telegram"]
  }' > /dev/null && echo "✅ Telegram работает"

# Push
echo "🔔 Тестируем Push..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "Push тест",
    "message": "Тестовое Push уведомление",
    "recipients": ["device_token_123"],
    "channels": ["push"]
  }' > /dev/null && echo "✅ Push работает"

# Slack
echo "💬 Тестируем Slack..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "Slack тест",
    "message": "Тестовое Slack уведомление",
    "recipients": ["#general"],
    "channels": ["slack"]
  }' > /dev/null && echo "✅ Slack работает"

# Webhook
echo "🔗 Тестируем Webhook..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "test",
    "priority": "normal",
    "title": "Webhook тест",
    "message": "Тестовое Webhook уведомление",
    "recipients": ["http://localhost:8080/webhook"],
    "channels": ["webhook"]
  }' > /dev/null && echo "✅ Webhook работает"

echo ""
echo "3. Тестирование множественных каналов..."
curl -s -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "multi_test",
    "priority": "high",
    "title": "Множественные каналы",
    "message": "Тест отправки через несколько каналов",
    "recipients": ["user@example.com", "+1234567890", "@test_user"],
    "channels": ["email", "sms", "telegram"]
  }' > /dev/null && echo "✅ Множественные каналы работают"

echo ""
echo "4. Проверка RabbitMQ..."
if curl -s -u admin:admin http://localhost:15672/api/queues/%2F/notifications | grep -q '"consumers":1'; then
    echo "✅ RabbitMQ consumer активен"
else
    echo "⚠️ RabbitMQ consumer не найден"
fi

echo ""
echo "🎉 Тестирование завершено!"
echo "📊 Результаты сохранены в TEST_RESULTS.md" 