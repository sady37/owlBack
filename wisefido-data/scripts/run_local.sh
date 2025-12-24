#!/bin/bash

# 使用 8081 端口运行（避免与 Docker 冲突）
cd "$(dirname "$0")/.."

export HTTP_ADDR=:8081
export DB_ENABLED=true
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=
export REDIS_DB=0

export DOCTOR_ENABLED=true
export LOG_LEVEL=info

echo "Starting wisefido-data on port 8081..."
echo "Note: Docker service is running on port 8080"
echo ""
go run cmd/wisefido-data/main.go
