#!/bin/bash

# Скрипт для генерации gRPC кода из proto файлов

set -e

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/.."

# Создаем директорию для сгенерированных файлов если её нет
mkdir -p api/proto

# Устанавливаем пути
PROTO_DIR="api/proto"
OUTPUT_DIR="api/proto"

# Проверяем наличие protoc
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install Protocol Buffers compiler."
    exit 1
fi

# Проверяем наличие protoc-gen-go
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Error: protoc-gen-go is not installed. Please install it with:"
    echo "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# Проверяем наличие protoc-gen-go-grpc
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Error: protoc-gen-go-grpc is not installed. Please install it with:"
    echo "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

echo "Generating gRPC code for analytics service..."

# Генерируем Go код из proto файлов
protoc \
    --go_out=$OUTPUT_DIR \
    --go_opt=paths=source_relative \
    --go-grpc_out=$OUTPUT_DIR \
    --go-grpc_opt=paths=source_relative \
    --proto_path=$PROTO_DIR \
    $PROTO_DIR/analytics.proto

echo "gRPC code generation completed successfully!"

# Показываем сгенерированные файлы
echo "Generated files:"
ls -la $OUTPUT_DIR/*.pb.go 2>/dev/null || echo "No .pb.go files found"
