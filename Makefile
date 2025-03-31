.PHONY: build run test lint clean docker-build docker-run migrate-up migrate-down deps help init

APP_NAME=bookshop-api
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

init: deps ## Initialize project
	@echo "==> Initializing project..."
	@mkdir -p $(BUILD_DIR)
	@if [ ! -d "$(MIGRATIONS_DIR)" ]; then mkdir -p $(MIGRATIONS_DIR); fi
	@echo "==> Project successfully initialized. To run: make run"

help: ## Show command help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "==> Installing dependencies..."
	go mod download
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

build: ## Build application
	@echo "==> Building application..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) -ldflags="-s -w" ./cmd/api

run: ## Run application locally
	@echo "==> Starting application..."
	go run ./cmd/api/main.go

swag: ## Generate Swagger documentation
	@echo "==> Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs

test: ## Run tests
	@echo "==> Running tests..."
	go test -race -cover ./...

mock: ## Generate mocks
	@echo "==> Generating mocks..."
	mockgen -source=internal/domain/repositories/user_repository.go -destination=internal/domain/repositories/mocks/user_repository_mock.go
	mockgen -source=internal/domain/repositories/book_repository.go -destination=internal/domain/repositories/mocks/book_repository_mock.go
	mockgen -source=internal/domain/repositories/category_repository.go -destination=internal/domain/repositories/mocks/category_repository_mock.go
	mockgen -source=internal/domain/repositories/cart_repository.go -destination=internal/domain/repositories/mocks/cart_repository_mock.go

lint: ## Run linter
	@echo "==> Running linter..."
	golangci-lint run

migrate-up: ## Run migrations up
	@echo "==> Running migrations up..."
	migrate -path $(MIGRATIONS_DIR) -database "postgres://bookshop:bookshop@localhost:5432/bookshop?sslmode=disable" up

migrate-down: ## Rollback migrations
	@echo "==> Rolling back migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "postgres://bookshop:bookshop@localhost:5432/bookshop?sslmode=disable" down

docker-build: ## Build Docker image
	@echo "==> Building Docker image..."
	docker build -t $(APP_NAME) .

docker-up: ## Run application in Docker
	@echo "==> Starting in Docker..."
	docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "==> Stopping containers..."
	docker-compose down

docker-logs: ## Show container logs
	@echo "==> Container logs..."
	docker-compose logs -f

clean: ## Clean build artifacts
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) 