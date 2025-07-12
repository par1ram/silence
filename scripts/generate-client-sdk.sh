#!/bin/bash

# Скрипт для объединения swagger файлов и генерации клиентского SDK
# Использует OpenAPI React Query Codegen для генерации TypeScript hooks

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода логов
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."

    # Проверяем наличие Node.js
    if ! command -v node &> /dev/null; then
        error "Node.js не найден. Установите Node.js версии 18 или выше."
    fi

    # Проверяем наличие npm
    if ! command -v npm &> /dev/null; then
        error "npm не найден. Установите npm."
    fi

    # Проверяем наличие jq для работы с JSON
    if ! command -v jq &> /dev/null; then
        warn "jq не найден. Устанавливаю jq..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install jq
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            sudo apt-get install -y jq
        else
            error "Не удалось установить jq. Установите его вручную."
        fi
    fi

    log "✅ Все зависимости проверены"
}

# Создание объединенного swagger файла
merge_swagger_files() {
    log "Объединение swagger файлов..."

    SWAGGER_DIR="docs/swagger"
    OUTPUT_FILE="$SWAGGER_DIR/unified-api.json"

    # Создаем базовую структуру
    cat > "$OUTPUT_FILE" << 'EOF'
{
  "swagger": "2.0",
  "info": {
    "title": "Silence VPN API",
    "version": "1.0.0",
    "description": "Unified API documentation for Silence VPN platform",
    "contact": {
      "name": "Silence VPN Team",
      "email": "support@silence-vpn.com"
    },
    "license": {
      "name": "MIT",
      "url": "https://opensource.org/licenses/MIT"
    }
  },
  "host": "localhost:8080",
  "basePath": "/api/v1",
  "schemes": ["http", "https"],
  "consumes": ["application/json"],
  "produces": ["application/json"],
  "securityDefinitions": {
    "bearerAuth": {
      "type": "apiKey",
      "in": "header",
      "name": "Authorization",
      "description": "Bearer token authorization header"
    }
  },
  "security": [
    {
      "bearerAuth": []
    }
  ],
  "paths": {},
  "definitions": {}
}
EOF

    # Массив файлов для объединения
    SWAGGER_FILES=(
        "auth.swagger.json"
        "analytics.swagger.json"
        "server-manager.swagger.json"
        "vpn-core.swagger.json"
        "dpi-bypass.swagger.json"
        "notifications.swagger.json"
    )

    # Объединяем paths и definitions из всех файлов
    for file in "${SWAGGER_FILES[@]}"; do
        if [[ -f "$SWAGGER_DIR/$file" ]]; then
            log "Обрабатываю $file..."

            # Добавляем paths
            jq -s '.[0].paths as $base | .[1].paths as $new | .[0] | .paths = ($base + $new)' \
                "$OUTPUT_FILE" "$SWAGGER_DIR/$file" > "$OUTPUT_FILE.tmp" && mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"

            # Добавляем definitions
            jq -s '.[0].definitions as $base | .[1].definitions as $new | .[0] | .definitions = ($base + $new)' \
                "$OUTPUT_FILE" "$SWAGGER_DIR/$file" > "$OUTPUT_FILE.tmp" && mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"

            # Добавляем теги
            jq -s '.[0].tags as $base | .[1].tags as $new | .[0] | .tags = (($base // []) + ($new // []) | unique_by(.name))' \
                "$OUTPUT_FILE" "$SWAGGER_DIR/$file" > "$OUTPUT_FILE.tmp" && mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"
        else
            warn "Файл $file не найден, пропускаю..."
        fi
    done

    log "✅ Объединенный swagger файл создан: $OUTPUT_FILE"
}

# Установка зависимостей для генерации
install_codegen_deps() {
    log "Установка зависимостей для генерации клиентского SDK..."

    # Переходим в папку frontend
    cd frontend

    # Устанавливаем OpenAPI React Query Codegen
    if ! npm list @7nohe/openapi-react-query-codegen &> /dev/null; then
        log "Устанавливаю @7nohe/openapi-react-query-codegen..."
        npm install -D @7nohe/openapi-react-query-codegen
    fi

    # Проверяем наличие @tanstack/react-query
    if ! npm list @tanstack/react-query &> /dev/null; then
        log "Устанавливаю @tanstack/react-query..."
        npm install @tanstack/react-query
    fi

    # Возвращаемся в корневую папку
    cd ..

    log "✅ Зависимости установлены"
}

# Генерация клиентского SDK
generate_client_sdk() {
    log "Генерация клиентского SDK..."

    # Переходим в папку frontend
    cd frontend

    # Создаем директорию для сгенерированного API
    mkdir -p src/generated

    # Генерируем TypeScript SDK
    npx openapi-rq \
        --input ../docs/swagger/unified-api.json \
        --output src/generated

    # Возвращаемся в корневую папку
    cd ..

    log "✅ Клиентский SDK сгенерирован в frontend/src/generated"
}

# Создание конфигурационного файла для автогенерации
create_config_file() {
    log "Создание конфигурационного файла..."

    cat > "frontend/openapi-codegen.config.ts" << 'EOF'
import type { ConfigFile } from '@7nohe/openapi-react-query-codegen'

const config: ConfigFile = {
  schemaFile: '../docs/swagger/unified-api.json',
  apiFile: './src/generated/api.ts',
  outputDir: './src/generated',
  exportCore: true,
  exportServices: true,
  exportModels: true,
  exportSchemas: true,
  useOptions: true,
  useUnionTypes: true,
  client: 'axios',
  httpClient: 'axios',
  format: ['prettier', 'eslint'],
  lint: true,
  prettier: true,
  operationId: true,
  serviceResponse: 'response',
  base: 'Server',
  request: './src/lib/request.ts',
}

export default config
EOF

    log "✅ Конфигурационный файл создан"
}

# Создание кастомного HTTP клиента
create_http_client() {
    log "Создание кастомного HTTP клиента..."

    mkdir -p frontend/src/lib

    cat > "frontend/src/lib/request.ts" << 'EOF'
import axios, { AxiosError, AxiosResponse } from 'axios'

// Создаем экземпляр axios с базовой конфигурацией
const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Интерцептор для добавления токена авторизации
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Интерцептор для обработки ответов
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as any

    // Если получили 401 и это не повторный запрос
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = localStorage.getItem('refreshToken')
        if (refreshToken) {
          const response = await api.post('/auth/refresh', {
            refreshToken: refreshToken,
          })

          const { accessToken, refreshToken: newRefreshToken } = response.data
          localStorage.setItem('accessToken', accessToken)
          localStorage.setItem('refreshToken', newRefreshToken)

          // Повторяем оригинальный запрос с новым токеном
          originalRequest.headers.Authorization = `Bearer ${accessToken}`
          return api(originalRequest)
        }
      } catch (refreshError) {
        // Если обновление токена не удалось, очищаем хранилище и редиректим на логин
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default api
EOF

    log "✅ HTTP клиент создан"
}

# Обновление package.json с новыми скриптами
update_package_json() {
    log "Обновление package.json..."

    cd frontend

    # Добавляем скрипты для генерации API
    npm pkg set scripts.generate:api="openapi-rq --input ../docs/swagger/unified-api.json --output src/generated"
    npm pkg set scripts.generate:api:watch="npm run generate:api -- --watch"
    npm pkg set scripts.api:validate="openapi-rq --input ../docs/swagger/unified-api.json --validate"

    cd ..

    log "✅ package.json обновлен"
}

# Создание типов для API
create_api_types() {
    log "Создание дополнительных типов для API..."

    cat > "frontend/src/generated/types.ts" << 'EOF'
// Дополнительные типы для API
export interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
  timestamp: string
}

export interface PaginatedResponse<T = any> {
  data: T[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
    hasNext: boolean
    hasPrev: boolean
  }
}

export interface ErrorResponse {
  error: {
    code: string
    message: string
    details?: any
  }
  timestamp: string
}

// Типы для аутентификации
export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  firstName: string
  lastName: string
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: User
}

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: string
  createdAt: string
  updatedAt: string
}

// Типы для VPN
export interface VPNServer {
  id: string
  name: string
  region: string
  country: string
  city: string
  load: number
  ping: number
  bandwidth: number
  online: boolean
  protocols: string[]
}

export interface VPNConnection {
  id: string
  serverId: string
  userId: string
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
  connectedAt?: string
  disconnectedAt?: string
  bytesReceived: number
  bytesSent: number
  duration: number
}

// Типы для аналитики
export interface ConnectionMetrics {
  totalConnections: number
  activeConnections: number
  averageSessionDuration: number
  totalDataTransferred: number
  topCountries: { country: string; count: number }[]
  topServers: { serverId: string; count: number }[]
}

export interface UserActivityMetrics {
  userId: string
  totalSessions: number
  totalDuration: number
  totalDataTransferred: number
  favoriteServers: string[]
  lastActivity: string
}

// Типы для уведомлений
export interface Notification {
  id: string
  userId: string
  type: 'info' | 'warning' | 'error' | 'success'
  title: string
  message: string
  read: boolean
  createdAt: string
  expiresAt?: string
}

// Типы для системы управления серверами
export interface ServerStats {
  serverId: string
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  networkIn: number
  networkOut: number
  activeConnections: number
  timestamp: string
}

export interface ServerConfig {
  id: string
  name: string
  region: string
  specs: {
    cpu: string
    memory: string
    disk: string
    bandwidth: string
  }
  config: {
    maxConnections: number
    protocols: string[]
    features: string[]
  }
  status: 'active' | 'inactive' | 'maintenance'
}
EOF

    log "✅ Дополнительные типы созданы"
}

# Создание утилит для работы с API
create_api_utils() {
    log "Создание утилит для работы с API..."

    mkdir -p frontend/src/lib

    cat > "frontend/src/lib/api-utils.ts" << 'EOF'
import { AxiosError } from 'axios'
import { toast } from 'react-hot-toast'

// Утилита для обработки ошибок API
export const handleApiError = (error: AxiosError | Error) => {
  if (error instanceof AxiosError) {
    const message = error.response?.data?.error?.message || error.message
    const status = error.response?.status

    switch (status) {
      case 400:
        toast.error(`Неверные данные: ${message}`)
        break
      case 401:
        toast.error('Необходима авторизация')
        break
      case 403:
        toast.error('Доступ запрещен')
        break
      case 404:
        toast.error('Ресурс не найден')
        break
      case 500:
        toast.error('Внутренняя ошибка сервера')
        break
      default:
        toast.error(`Ошибка: ${message}`)
    }
  } else {
    toast.error(`Ошибка: ${error.message}`)
  }
}

// Утилита для форматирования данных
export const formatBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

// Утилита для форматирования времени
export const formatDuration = (seconds: number) => {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (hours > 0) {
    return `${hours}ч ${minutes}м ${remainingSeconds}с`
  } else if (minutes > 0) {
    return `${minutes}м ${remainingSeconds}с`
  } else {
    return `${remainingSeconds}с`
  }
}

// Утилита для валидации данных
export const validateEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

export const validatePassword = (password: string) => {
  return password.length >= 8 && /[A-Z]/.test(password) && /[a-z]/.test(password) && /\d/.test(password)
}

// Утилита для работы с токенами
export const getAuthToken = () => {
  return localStorage.getItem('accessToken')
}

export const setAuthTokens = (accessToken: string, refreshToken: string) => {
  localStorage.setItem('accessToken', accessToken)
  localStorage.setItem('refreshToken', refreshToken)
}

export const clearAuthTokens = () => {
  localStorage.removeItem('accessToken')
  localStorage.removeItem('refreshToken')
}

export const isAuthenticated = () => {
  return !!getAuthToken()
}

// Утилита для работы с WebSocket
export const createWebSocketConnection = (url: string, token?: string) => {
  const wsUrl = new URL(url)
  if (token) {
    wsUrl.searchParams.append('token', token)
  }

  return new WebSocket(wsUrl.toString())
}

// Утилита для дебаунса
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) => {
  let timeoutId: NodeJS.Timeout

  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => func(...args), delay)
  }
}

// Утилита для throttle
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean

  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}
EOF

    log "✅ Утилиты для работы с API созданы"
}

# Главная функция
main() {
    log "🚀 Запуск генерации клиентского SDK для Silence VPN..."

    # Проверяем, что мы в корневой папке проекта
    if [[ ! -f "go.work" ]]; then
        error "Скрипт должен запускаться из корневой папки проекта"
    fi

    # Выполняем все этапы
    check_dependencies
    merge_swagger_files
    install_codegen_deps
    create_config_file
    create_http_client
    create_api_types
    create_api_utils
    update_package_json
    generate_client_sdk

    log "✅ Генерация клиентского SDK завершена успешно!"
    log ""
    log "📁 Структура сгенерированных файлов:"
    log "   └── frontend/src/generated/     - Сгенерированные API хуки"
    log "   └── frontend/src/lib/          - Утилиты для работы с API"
    log "   └── docs/swagger/unified-api.json - Объединенная API схема"
    log ""
    log "🔧 Доступные команды:"
    log "   npm run generate:api          - Пересгенерировать API"
    log "   npm run generate:api:watch    - Автоматическая регенерация"
    log "   npm run api:validate          - Валидация API схемы"
    log ""
    log "📖 Документация API доступна по адресу: http://localhost:8080/swagger"
}

# Запуск скрипта
main "$@"
