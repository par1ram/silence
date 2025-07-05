#!/bin/bash

# Скрипт для развертывания Server Manager в Kubernetes
# Использование: ./scripts/deploy-k8s.sh [namespace]

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl не установлен"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Проверка подключения к кластеру
check_cluster() {
    log_info "Проверка подключения к Kubernetes кластеру..."
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Не удается подключиться к Kubernetes кластеру"
        exit 1
    fi
    
    log_success "Подключение к кластеру установлено"
}

# Сборка Docker образа
build_image() {
    log_info "Сборка Docker образа..."
    
    docker build -t silence/server-manager:latest -f rpc/server-manager/Dockerfile .
    
    log_success "Docker образ собран"
}

# Создание namespace
create_namespace() {
    local namespace=${1:-silence-vpn}
    
    log_info "Создание namespace: $namespace"
    
    if ! kubectl get namespace $namespace &> /dev/null; then
        kubectl create namespace $namespace
        log_success "Namespace $namespace создан"
    else
        log_warning "Namespace $namespace уже существует"
    fi
}

# Применение манифестов
apply_manifests() {
    local namespace=${1:-silence-vpn}
    
    log_info "Применение Kubernetes манифестов..."
    
    # Обновляем namespace в манифесте
    sed "s/namespace: silence-vpn/namespace: $namespace/g" deployments/kubernetes/server-manager.yaml | \
    sed "s/KUBERNETES_NAMESPACE: \"silence-vpn\"/KUBERNETES_NAMESPACE: \"$namespace\"/g" | \
    kubectl apply -f -
    
    log_success "Манифесты применены"
}

# Ожидание готовности подов
wait_for_pods() {
    local namespace=${1:-silence-vpn}
    
    log_info "Ожидание готовности подов..."
    
    kubectl wait --for=condition=ready pod -l app=server-manager -n $namespace --timeout=300s
    
    log_success "Поды готовы"
}

# Проверка статуса развертывания
check_deployment() {
    local namespace=${1:-silence-vpn}
    
    log_info "Проверка статуса развертывания..."
    
    kubectl get pods -n $namespace -l app=server-manager
    kubectl get services -n $namespace -l app=server-manager
    
    log_success "Развертывание завершено"
}

# Получение внешнего IP
get_external_ip() {
    local namespace=${1:-silence-vpn}
    
    log_info "Получение внешнего IP..."
    
    # Ждем назначения внешнего IP
    kubectl wait --for=condition=ready service/server-manager-service -n $namespace --timeout=300s
    
    local external_ip=$(kubectl get service server-manager-service -n $namespace -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    if [ -n "$external_ip" ]; then
        log_success "Внешний IP: $external_ip"
        log_info "API доступен по адресу: http://$external_ip/api/v1/servers"
    else
        log_warning "Внешний IP не назначен (возможно, используется NodePort или ClusterIP)"
        log_info "Используйте 'kubectl port-forward' для доступа к сервису"
    fi
}

# Основная функция
main() {
    local namespace=${1:-silence-vpn}
    
    log_info "Начало развертывания Server Manager в Kubernetes"
    log_info "Namespace: $namespace"
    
    check_dependencies
    check_cluster
    build_image
    create_namespace $namespace
    apply_manifests $namespace
    wait_for_pods $namespace
    check_deployment $namespace
    get_external_ip $namespace
    
    log_success "Развертывание завершено успешно!"
    log_info "Для проверки работы API выполните:"
    log_info "curl http://<external-ip>/health"
    log_info "curl http://<external-ip>/api/v1/servers"
}

# Обработка аргументов командной строки
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Использование: $0 [namespace]"
    echo ""
    echo "Аргументы:"
    echo "  namespace    Namespace для развертывания (по умолчанию: silence-vpn)"
    echo ""
    echo "Примеры:"
    echo "  $0                    # Развертывание в namespace silence-vpn"
    echo "  $0 my-namespace       # Развертывание в namespace my-namespace"
    exit 0
fi

# Запуск основной функции
main "$@" 