.PHONY: build run test lint clean docker-build docker-run migrate-up migrate-down deps help init

APP_NAME=bookshop-api
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

init: deps ## Инициализировать проект
	@echo "==> Инициализация проекта..."
	@mkdir -p $(BUILD_DIR)
	@if [ ! -d "$(MIGRATIONS_DIR)" ]; then mkdir -p $(MIGRATIONS_DIR); fi
	@echo "==> Проект успешно инициализирован. Для запуска: make run"

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Установить зависимости
	@echo "==> Установка зависимостей..."
	go mod download
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

build: ## Собрать приложение
	@echo "==> Сборка приложения..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) -ldflags="-s -w" ./cmd/api

run: ## Запустить приложение локально
	@echo "==> Запуск приложения..."
	go run ./cmd/api/main.go

swag: ## Сгенерировать Swagger документацию
	@echo "==> Генерация Swagger документации..."
	swag init -g cmd/api/main.go -o docs

test: ## Запустить тесты
	@echo "==> Запуск тестов..."
	go test -race -cover ./...

mock: ## Сгенерировать моки
	@echo "==> Генерация моков..."
	mockgen -source=internal/domain/repositories/user_repository.go -destination=internal/domain/repositories/mocks/user_repository_mock.go
	mockgen -source=internal/domain/repositories/book_repository.go -destination=internal/domain/repositories/mocks/book_repository_mock.go
	mockgen -source=internal/domain/repositories/category_repository.go -destination=internal/domain/repositories/mocks/category_repository_mock.go
	mockgen -source=internal/domain/repositories/cart_repository.go -destination=internal/domain/repositories/mocks/cart_repository_mock.go

lint: ## Запустить линтер
	@echo "==> Запуск линтера..."
	golangci-lint run

migrate-up: ## Выполнить миграции вверх
	@echo "==> Выполнение миграций вверх..."
	migrate -path $(MIGRATIONS_DIR) -database "postgres://bookshop:bookshop@localhost:5432/bookshop?sslmode=disable" up

migrate-down: ## Откатить миграции
	@echo "==> Откат миграций..."
	migrate -path $(MIGRATIONS_DIR) -database "postgres://bookshop:bookshop@localhost:5432/bookshop?sslmode=disable" down

docker-build: ## Собрать Docker образ
	@echo "==> Сборка Docker образа..."
	docker build -t $(APP_NAME) .

docker-up: ## Запустить приложение в Docker
	@echo "==> Запуск в Docker..."
	docker-compose up -d

docker-down: ## Остановить контейнеры Docker
	@echo "==> Остановка контейнеров..."
	docker-compose down

docker-logs: ## Показать логи контейнеров
	@echo "==> Логи контейнеров..."
	docker-compose logs -f

clean: ## Очистить артефакты сборки
	@echo "==> Очистка артефактов сборки..."
	rm -rf $(BUILD_DIR) 