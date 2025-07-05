#!/bin/bash

# Генерируем gRPC код из proto файлов
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/vpn.proto

echo "gRPC код сгенерирован" 