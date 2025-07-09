# TODO

## Исправление линтеров
- [x] Исправить errcheck ошибки в vpn-core тестах
- [x] Исправить errcheck ошибки в auth сервисе
- [x] Исправить errcheck ошибки в остальных RPC сервисах
- [x] Запустить общий линтер для всех сервисов

## Миграция на gRPC
- [x] Создать proto файлы для всех RPC сервисов
- [x] Заменить HTTP адаптеры на gRPC в analytics сервисе
- [x] Заменить HTTP адаптеры на gRPC в dpi-bypass сервисе
- [x] Заменить HTTP адаптеры на gRPC в server-manager сервисе
- [x] Заменить HTTP адаптеры на gRPC в notifications сервисе
- [x] Обновить конфигурацию сервисов для gRPC

## Рефакторинг пользователей
- [x] Перенести CRUD пользователей из gateway в auth сервис
- [x] Создать gRPC методы для управления пользователями в auth
- [x] Обновить gateway для вызова auth сервиса вместо прямого CRUD
- [x] Обновить тесты после рефакторинга

## gRPC Gateway и документация
- [x] Добавить grpc-gateway в gateway сервис
- [x] Настроить генерацию Swagger документации из proto файлов
- [x] Обновить proto файлы с аннотациями для HTTP API
- [x] Создать единую точку входа для HTTP и gRPC в gateway
- [x] Вынести все локальные состояние в Redis cluster, реализовать балансировку нагрузки на уровне api Gateway
  - [x] RateLimiter в gateway - перенесен на Redis с поддержкой sliding window
  - [x] WebSocketHandler в gateway - создан Redis-based manager для сессий с TTL
  - [x] GRPCClients в gateway - создан Redis-based manager с health check и load balancing
  - [x] AlertService в analytics - пропускаем, будем переделывать на OpenTelemetry
  - [x] CustomAdapter в dpi-bypass - пропускаем, локальное состояние соединений оправдано
  - [x] Создан comprehensive Redis client с HIncrBy, ZIncrBy методами
  - [x] Добавлены интеграционные тесты для Redis компонентов
  - [x] Создана конфигурация для Redis в production-ready формате
- [ ] Переделать текущую аналитику на OpenTelemetry (Prometheus, Grafana, Loki, Jaeger, Zipkin, Tempo)
- [ ] Проверить соответвие моделей данных и миграций с текущим API


## Финальная проверка
- [ ] Проверить все ручки напрямую
- [ ] Доработать Docker файлы и docker-compose
- [ ] Проверить работу всех сервисов в контейнерах
- [ ] Обновить документацию по API

- [ ] Интегрировать frontend с backend
