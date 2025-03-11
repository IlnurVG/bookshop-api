#!/bin/bash

# Установка разрешений на выполнение
chmod +x scripts/setup.sh
chmod +x scripts/build.sh

# Проверка наличия Docker
if ! command -v docker &> /dev/null; then
    echo "Docker не установлен. Установите Docker для запуска зависимостей."
    exit 1
fi

# Проверка наличия Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose не установлен. Установите Docker Compose для запуска зависимостей."
    exit 1
fi

# Проверка и установка зависимостей Go
if ! command -v golangci-lint &> /dev/null; then
    echo "Установка golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

if ! command -v swag &> /dev/null; then
    echo "Установка swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

if ! command -v mockgen &> /dev/null; then
    echo "Установка mockgen..."
    go install github.com/golang/mock/mockgen@latest
fi

if ! command -v migrate &> /dev/null; then
    echo "Установка migrate..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Создание директории для сборки
mkdir -p bin

echo "==> Создание go.sum..."
go mod tidy

echo "==> Настройка окружения завершена."
echo "Для инициализации проекта выполните: make init"
echo "Для запуска приложения выполните: make run"


