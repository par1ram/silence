# API Reference - Silence VPN

## 1. Gateway (http://localhost:8080)

**Единая точка входа. Все запросы к другим сервисам должны проходить через Gateway.**

### Основные эндпоинты

- `GET /health`: Проверка состояния.

### Аутентификация (`/api/v1/auth`)
- `POST /register`: Регистрация нового пользователя.
- `POST /login`: Вход и получение JWT токена.
- `GET /me`: Получение информации о текущем пользователе (требует JWT).

### VPN и Обфускация
- `POST /api/v1/connect`: **(Core-фича)** Создание VPN-соединения с обфускацией трафика (требует JWT).

### Прокси к другим сервисам (требуют JWT)
- `/api/v1/vpn/*`: Прокси к **VPN Core**.
- `/api/v1/dpi-bypass/*`: Прокси к **DPI Bypass**.
- `/api/v1/analytics/*`: Прокси к **Analytics**.
- `/api/v1/server-manager/*`: Прокси к **Server Manager**.
- `/api/v1/notifications/*`: Прокси к **Notifications**.

## 2. Auth Service

- **Отвечает за**: Регистрацию, аутентификацию и управление пользователями.
- **База данных**: PostgreSQL (`silence_auth`).

## 3. VPN Core Service

- **Отвечает за**: Управление туннелями WireGuard.
- **Эндпоинты**:
    - `POST /tunnels`: Создать туннель.
    - `GET /tunnels/list`: Список туннелей.
    - `POST /tunnels/start`: Запустить туннель.
    - `GET /tunnels/stats`: Получить статистику.
    - `POST /peers/add`: Добавить пира.

## 4. DPI Bypass Service

- **Отвечает за**: Обфускацию трафика для обхода DPI.
- **Методы**: Shadowsocks, V2Ray, Obfs4.
- **Эндпоинты**:
    - `POST /api/v1/bypass`: Создать конфигурацию обфускации.
    - `GET /api/v1/bypass`: Список конфигураций.
    - `POST /api/v1/bypass/{id}/start`: Запустить обфускацию.

## 5. Analytics Service

- **Отвечает за**: Сбор, хранение и предоставление метрик.
- **База данных**: InfluxDB.
- **Эндпоинты**:
    - `GET /metrics/connections`: Метрики подключений.
    - `GET /metrics/server-load`: Нагрузка на серверы.
    - `GET /dashboards`: Список дашбордов.
    - `POST /alerts`: Создать алерт.

## 6. Server Manager Service

- **Отвечает за**: Управление жизненным циклом серверов (Docker или Kubernetes).
- **Эндпоинты**:
    - `POST /api/v1/servers`: Создать сервер.
    - `GET /api/v1/servers`: Список серверов.
    - `POST /api/v1/servers/{id}/start`: Запустить сервер.
    - `GET /api/v1/servers/{id}/stats`: Статистика сервера.

## 7. Notifications Service

- **Отвечает за**: Отправку уведомлений по разным каналам.
- **Каналы**: Email, SMS, Push, Telegram, Slack, Webhook.
- **Интеграция**: RabbitMQ.
- **Эндпоинты**:
    - `POST /notifications`: Отправить уведомление.

## Пример: Полный цикл подключения

1.  **Регистрация**
    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/register -d '{"email":"user@test.com","password":"pass"}'
    ```
2.  **Вход**
    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/login -d '{"email":"user@test.com","password":"pass"}'
    # -> Получаем JWT_TOKEN
    ```
3.  **Подключение к VPN с обфускацией**
    ```bash
    curl -X POST http://localhost:8080/api/v1/connect \
      -H "Authorization: Bearer <JWT_TOKEN>" \
      -d '{
        "bypass_method": "shadowsocks",
        "bypass_config": { ... },
        "vpn_config": { ... }
      }'
    ```
