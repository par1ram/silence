#!/bin/bash

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Генерация Swagger документации ===${NC}"

# Переходим в корень проекта gateway
cd "$(dirname "$0")/.."

# Проверяем наличие необходимых инструментов
echo -e "${BLUE}Проверка инструментов...${NC}"

if ! command -v protoc &> /dev/null; then
    echo -e "${RED}❌ protoc не найден. Установите Protocol Buffers compiler${NC}"
    echo -e "${YELLOW}macOS: brew install protobuf${NC}"
    echo -e "${YELLOW}Ubuntu: sudo apt install protobuf-compiler${NC}"
    exit 1
fi

if ! command -v protoc-gen-openapiv2 &> /dev/null; then
    echo -e "${RED}❌ protoc-gen-openapiv2 не найден.${NC}"
    echo -e "${YELLOW}Устанавливаем...${NC}"
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
fi

echo -e "${GREEN}✅ Все инструменты готовы${NC}"

# Создаем необходимые директории
echo -e "${BLUE}Создание директорий...${NC}"
mkdir -p ../../docs/swagger
mkdir -p ../../docs/api-specs

# Определяем пути
PROTO_PATH="api/proto"
SWAGGER_OUTPUT="../../docs/swagger"
GOOGLEAPIS_PATH="third_party/googleapis"

# Функция для генерации Swagger для одного сервиса
generate_swagger() {
    local service_name=$1
    local proto_file=$2
    local service_title=$3

    echo -e "${YELLOW}📝 Генерация Swagger для ${service_title}...${NC}"

    # Генерируем Swagger документацию
    protoc \
        --proto_path="${PROTO_PATH}" \
        --proto_path="${GOOGLEAPIS_PATH}" \
        --openapiv2_out="${SWAGGER_OUTPUT}" \
        --openapiv2_opt=logtostderr=true,allow_merge=true,merge_file_name="${service_name}" \
        "${proto_file}" 2>/dev/null && {
        echo -e "${GREEN}✅ Swagger для ${service_title} создан${NC}"
        return 0
    }

    # Если не удалось с googleapis, попробуем без них
    echo -e "${YELLOW}⚠️  Используем упрощенную генерацию для ${service_name}${NC}"
    protoc \
        --proto_path="${PROTO_PATH}" \
        --openapiv2_out="${SWAGGER_OUTPUT}" \
        --openapiv2_opt=logtostderr=true,allow_merge=true,merge_file_name="${service_name}" \
        "${proto_file}" 2>/dev/null && {
        echo -e "${GREEN}✅ Swagger для ${service_title} создан (упрощенная версия)${NC}"
        return 0
    }

    echo -e "${RED}❌ Ошибка генерации Swagger для ${service_name}${NC}"
    return 1
}

# Генерируем Swagger для каждого сервиса
echo -e "${BLUE}Генерация Swagger документации для всех сервисов...${NC}"

generate_swagger "auth" "${PROTO_PATH}/auth/auth.proto" "Authentication Service"
generate_swagger "notifications" "${PROTO_PATH}/notifications/notifications.proto" "Notifications Service"
generate_swagger "vpn-core" "${PROTO_PATH}/vpn-core/vpn.proto" "VPN Core Service"
generate_swagger "analytics" "${PROTO_PATH}/analytics/analytics.proto" "Analytics Service"
generate_swagger "server-manager" "${PROTO_PATH}/server-manager/server.proto" "Server Manager Service"
generate_swagger "dpi-bypass" "${PROTO_PATH}/dpi-bypass/dpi.proto" "DPI Bypass Service"

# Создаем объединенный index.html для всех API
echo -e "${BLUE}Создание объединенного API документа...${NC}"

cat > "${SWAGGER_OUTPUT}/index.html" << 'EOF'
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Silence VPN API Documentation</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .service-card {
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .service-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .service-title {
            color: #2c3e50;
            margin-bottom: 10px;
        }
        .service-description {
            color: #666;
            margin-bottom: 15px;
            font-size: 14px;
        }
        .service-link {
            display: inline-block;
            background: #3498db;
            color: white;
            padding: 8px 16px;
            text-decoration: none;
            border-radius: 4px;
            transition: background 0.2s;
            margin: 5px;
        }
        .service-link:hover {
            background: #2980b9;
        }
        .info-section {
            margin-top: 40px;
            padding-top: 30px;
            border-top: 1px solid #eee;
        }
        .info-section h2 {
            color: #2c3e50;
            margin-bottom: 15px;
        }
        .info-section p {
            color: #666;
            line-height: 1.6;
        }
        .auth-info {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
        }
        .auth-info h3 {
            color: #e74c3c;
            margin-bottom: 10px;
        }
        .code-block {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 14px;
            overflow-x: auto;
            margin-top: 10px;
        }
        .swagger-ui-link {
            background: #85c946;
        }
        .swagger-ui-link:hover {
            background: #6ba52b;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Silence VPN API Documentation</h1>

        <div class="services-grid">
            <div class="service-card">
                <h3 class="service-title">🔐 Authentication Service</h3>
                <p class="service-description">API для аутентификации и управления пользователями</p>
                <a href="auth.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=auth.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>

            <div class="service-card">
                <h3 class="service-title">📬 Notifications Service</h3>
                <p class="service-description">API для отправки и управления уведомлениями</p>
                <a href="notifications.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=notifications.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>

            <div class="service-card">
                <h3 class="service-title">🌐 VPN Core Service</h3>
                <p class="service-description">API для управления VPN туннелями и пирами</p>
                <a href="vpn-core.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=vpn-core.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>

            <div class="service-card">
                <h3 class="service-title">📊 Analytics Service</h3>
                <p class="service-description">API для сбора и анализа метрик</p>
                <a href="analytics.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=analytics.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>

            <div class="service-card">
                <h3 class="service-title">🖥️ Server Manager Service</h3>
                <p class="service-description">API для управления серверами и инфраструктурой</p>
                <a href="server-manager.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=server-manager.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>

            <div class="service-card">
                <h3 class="service-title">🔓 DPI Bypass Service</h3>
                <p class="service-description">API для обхода блокировок DPI</p>
                <a href="dpi-bypass.swagger.json" class="service-link" target="_blank">JSON Spec</a>
                <a href="https://petstore.swagger.io/?url=dpi-bypass.swagger.json" class="service-link swagger-ui-link" target="_blank">Swagger UI</a>
            </div>
        </div>

        <div class="info-section">
            <h2>📋 Информация об API</h2>
            <p>Добро пожаловать в документацию API Silence VPN. Все API endpoints используют JSON для запросов и ответов.</p>

            <div class="auth-info">
                <h3>🔑 Аутентификация</h3>
                <p>Большинство endpoints требуют аутентификации через Bearer токен в заголовке Authorization:</p>
                <div class="code-block">
Authorization: Bearer YOUR_JWT_TOKEN_HERE
                </div>
                <p>Токен можно получить через <strong>POST /api/v1/auth/login</strong> или <strong>POST /api/v1/auth/register</strong></p>
            </div>

            <h3>🌐 Base URL</h3>
            <div class="code-block">
Production: https://api.silence-vpn.com
Development: http://localhost:8080
            </div>

            <h3>📝 Форматы ответов</h3>
            <p>Все ответы возвращаются в формате JSON. Успешные ответы содержат данные, ошибки содержат поле <code>error</code>.</p>

            <h3>🔄 Статус коды</h3>
            <ul>
                <li><strong>200</strong> - Успешный запрос</li>
                <li><strong>201</strong> - Ресурс создан</li>
                <li><strong>400</strong> - Неверный запрос</li>
                <li><strong>401</strong> - Неавторизованный доступ</li>
                <li><strong>403</strong> - Доступ запрещен</li>
                <li><strong>404</strong> - Ресурс не найден</li>
                <li><strong>500</strong> - Внутренняя ошибка сервера</li>
            </ul>

            <h3>🛠️ Инструменты для тестирования</h3>
            <p>Рекомендуемые инструменты для работы с API:</p>
            <ul>
                <li><strong>Swagger UI</strong> - Интерактивная документация (ссылки выше)</li>
                <li><strong>Postman</strong> - Импортируйте OpenAPI спецификации</li>
                <li><strong>curl</strong> - Командная строка</li>
                <li><strong>HTTPie</strong> - Более удобная альтернатива curl</li>
            </ul>
        </div>
    </div>
</body>
</html>
EOF

echo -e "${GREEN}✅ Swagger документация успешно создана!${NC}"
echo -e "${GREEN}📁 Файлы сохранены в: ${SWAGGER_OUTPUT}${NC}"
echo -e "${GREEN}🌐 Откройте ${SWAGGER_OUTPUT}/index.html в браузере${NC}"
echo -e "${YELLOW}📝 Для интерактивного просмотра используйте Swagger UI${NC}"

# Показываем структуру созданных файлов
echo -e "\n${BLUE}📋 Созданные файлы:${NC}"
find "${SWAGGER_OUTPUT}" -type f -name "*.json" -o -name "*.html" | sort
