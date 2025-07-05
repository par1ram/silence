# Интеграция Server Manager с Kubernetes

## Обзор

Server Manager может работать как с Docker, так и с Kubernetes для управления серверами. Это позволяет гибко выбирать подходящую платформу оркестрации в зависимости от требований проекта.

## Архитектура

### Варианты развертывания

1. **Server Manager + Docker** - для простых сценариев и разработки
2. **Server Manager + Kubernetes** - для продакшн окружений
3. **Server Manager поверх Kubernetes** - для управления приложениями в K8s кластере

### Компоненты

- **Orchestrator Interface** - абстракция для управления серверами
- **Docker Adapter** - реализация для Docker
- **Kubernetes Adapter** - реализация для Kubernetes
- **Orchestrator Factory** - фабрика для создания нужного оркестратора

## Конфигурация

### Переменные окружения

```bash
# Тип оркестратора: "docker" или "kubernetes"
ORCHESTRATOR_TYPE=docker

# Kubernetes конфигурация (только для kubernetes режима)
KUBECONFIG=/path/to/kubeconfig
KUBERNETES_NAMESPACE=default

# Docker конфигурация (только для docker режима)
DOCKER_HOST=unix:///var/run/docker.sock
DOCKER_API_VERSION=1.41
DOCKER_TIMEOUT=30s
```

### Примеры конфигурации

#### Docker режим

```bash
export ORCHESTRATOR_TYPE=docker
export DOCKER_HOST=unix:///var/run/docker.sock
```

#### Kubernetes режим

```bash
export ORCHESTRATOR_TYPE=kubernetes
export KUBECONFIG=/home/user/.kube/config
export KUBERNETES_NAMESPACE=silence-vpn
```

## Развертывание в Kubernetes

### 1. Подготовка кластера

```bash
# Создаем namespace
kubectl create namespace silence-vpn

# Создаем ConfigMap для конфигурации
kubectl create configmap server-manager-config \
  --from-literal=ORCHESTRATOR_TYPE=kubernetes \
  --from-literal=KUBERNETES_NAMESPACE=silence-vpn \
  -n silence-vpn
```

### 2. Развертывание Server Manager

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: server-manager
  namespace: silence-vpn
spec:
  replicas: 1
  selector:
    matchLabels:
      app: server-manager
  template:
    metadata:
      labels:
        app: server-manager
    spec:
      serviceAccountName: server-manager-sa
      containers:
        - name: server-manager
          image: silence/server-manager:latest
          ports:
            - containerPort: 8085
          env:
            - name: ORCHESTRATOR_TYPE
              value: 'kubernetes'
            - name: KUBERNETES_NAMESPACE
              value: 'silence-vpn'
            - name: DB_HOST
              value: 'postgres-service'
            - name: DB_NAME
              value: 'silence_server_manager'
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: username
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: password
          resources:
            requests:
              memory: '128Mi'
              cpu: '100m'
            limits:
              memory: '256Mi'
              cpu: '200m'
```

### 3. Service Account и RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: server-manager-sa
  namespace: silence-vpn
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: server-manager-role
rules:
  - apiGroups: ['']
    resources: ['pods', 'services', 'configmaps', 'secrets']
    verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']
  - apiGroups: ['apps']
    resources: ['deployments', 'replicasets']
    verbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']
  - apiGroups: ['']
    resources: ['events']
    verbs: ['create', 'patch']
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: server-manager-binding
subjects:
  - kind: ServiceAccount
    name: server-manager-sa
    namespace: silence-vpn
roleRef:
  kind: ClusterRole
  name: server-manager-role
  apiGroup: rbac.authorization.k8s.io
```

### 4. Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: server-manager-service
  namespace: silence-vpn
spec:
  selector:
    app: server-manager
  ports:
    - name: http
      port: 80
      targetPort: 8085
  type: LoadBalancer
```

## Управление серверами

### Создание VPN сервера

```bash
curl -X POST http://server-manager-service/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "vpn-server-1",
    "type": "vpn",
    "region": "us-east-1"
  }'
```

### Масштабирование

```bash
# Масштабирование до 3 реплик (только в Kubernetes режиме)
curl -X POST http://server-manager-service/api/v1/servers/vpn-server-1/scale \
  -H "Content-Type: application/json" \
  -d '{"replicas": 3}'
```

### Мониторинг

```bash
# Получение статистики
curl http://server-manager-service/api/v1/servers/vpn-server-1/stats

# Проверка здоровья
curl http://server-manager-service/api/v1/servers/vpn-server-1/health
```

## Преимущества Kubernetes режима

### 1. Автоматическое масштабирование

- Horizontal Pod Autoscaler (HPA)
- Vertical Pod Autoscaler (VPA)
- Custom metrics

### 2. Высокая доступность

- ReplicaSets для отказоустойчивости
- Rolling updates
- Health checks и readiness probes

### 3. Управление ресурсами

- Resource quotas
- Limit ranges
- Namespace isolation

### 4. Мониторинг и логирование

- Prometheus метрики
- Centralized logging
- Distributed tracing

### 5. Безопасность

- RBAC
- Network policies
- Pod security policies

## Миграция с Docker на Kubernetes

### 1. Подготовка

```bash
# Экспорт конфигурации серверов
curl http://localhost:8085/api/v1/servers > servers-backup.json
```

### 2. Изменение конфигурации

```bash
# Остановка Server Manager
docker-compose stop server-manager

# Изменение переменных окружения
export ORCHESTRATOR_TYPE=kubernetes
export KUBECONFIG=/path/to/kubeconfig
export KUBERNETES_NAMESPACE=silence-vpn
```

### 3. Запуск в Kubernetes

```bash
# Применение манифестов
kubectl apply -f k8s/

# Проверка статуса
kubectl get pods -n silence-vpn
```

## Ограничения

### Docker режим

- Нет автоматического масштабирования
- Ограниченная отказоустойчивость
- Простое управление ресурсами

### Kubernetes режим

- Сложность настройки
- Требует дополнительных ресурсов
- Необходимость понимания K8s концепций

## Мониторинг и алертинг

### Prometheus метрики

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: server-manager-monitor
  namespace: silence-vpn
spec:
  selector:
    matchLabels:
      app: server-manager
  endpoints:
    - port: http
      interval: 30s
```

### Grafana дашборд

- Количество активных серверов
- Использование ресурсов
- Статус здоровья серверов
- Метрики производительности

## Troubleshooting

### Проблемы с подключением к Kubernetes

```bash
# Проверка конфигурации
kubectl config view

# Проверка прав доступа
kubectl auth can-i create deployments --namespace silence-vpn

# Проверка логов
kubectl logs -f deployment/server-manager -n silence-vpn
```

### Проблемы с созданием серверов

```bash
# Проверка событий
kubectl get events --sort-by='.lastTimestamp' -n silence-vpn

# Проверка статуса подов
kubectl describe pod -l app=vpn-server-1 -n silence-vpn
```

## Заключение

Интеграция Server Manager с Kubernetes предоставляет мощные возможности для управления VPN серверами в продакшн окружениях. Выбор между Docker и Kubernetes режимами зависит от требований к масштабируемости, отказоустойчивости и сложности инфраструктуры.
