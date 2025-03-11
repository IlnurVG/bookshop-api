#!/bin/bash

set -e

# Получение текущей директории
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$DIR/.." && pwd)"

# Установка переменных среды
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# Создание директории для сборки, если не существует
if [ ! -d "$PROJECT_ROOT/bin" ]; then
    mkdir -p "$PROJECT_ROOT/bin"
fi

# Очистка предыдущей сборки
rm -f "$PROJECT_ROOT/bin/bookshop-api"

echo "==> Загрузка зависимостей..."
cd "$PROJECT_ROOT" && go mod download

echo "==> Сборка приложения..."
cd "$PROJECT_ROOT" && go build -o "$PROJECT_ROOT/bin/bookshop-api" -ldflags="-s -w" ./cmd/api

echo "==> Сборка успешно завершена!"
echo "Исполняемый файл: $PROJECT_ROOT/bin/bookshop-api"


