FROM golang:1.24-alpine as builder

WORKDIR /app

# Кэширование зависимостей
COPY go.mod go.sum* ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookshop-api ./cmd/api

# Финальный образ
FROM alpine:3.19

WORKDIR /app

# Установка зависимостей
RUN apk --no-cache add ca-certificates tzdata

# Копирование бинарного файла из builder
COPY --from=builder /app/bookshop-api .
COPY --from=builder /app/config /app/config

# Экспорт порта
EXPOSE 8080

# Запуск приложения
CMD ["./bookshop-api"]
