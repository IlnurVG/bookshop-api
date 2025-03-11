package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bookshop/api/config"
	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/repository/postgres"
	"github.com/bookshop/api/internal/repository/redis"
	"github.com/bookshop/api/internal/server"
	"github.com/bookshop/api/pkg/logger"
)

func main() {
	// Парсинг флагов командной строки
	configPath := flag.String("config", "./config/config.yaml", "путь к файлу конфигурации")
	flag.Parse()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера
	l, err := logger.NewLogger(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer l.Sync()

	// Инициализация подключения к PostgreSQL
	db, err := postgres.NewPostgresDB(cfg.Database)
	if err != nil {
		l.Fatal("Ошибка подключения к базе данных", err)
	}
	defer db.Close()

	// Инициализация подключения к Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		l.Fatal("Ошибка подключения к Redis", err)
	}
	defer redisClient.Close()

	// Инициализация репозиториев
	bookRepo := postgres.NewBookRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)

	// Инициализация сервера с минимальными зависимостями
	srv, err := server.NewServer(cfg, l, nil, bookRepo, categoryRepo)
	if err != nil {
		l.Fatal("Ошибка инициализации сервера", err)
	}

	// Запуск сервера в горутине
	go func() {
		l.Info("Запуск HTTP сервера", "адрес", srv.Addr)
		if err := srv.Start(); err != nil {
			l.Error("Ошибка запуска сервера", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Завершение работы сервера...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("Ошибка при остановке сервера", err)
	}

	l.Info("Сервер успешно остановлен")
}

// cartRepositoryAdapter адаптер для CartRepository с поддержкой GetExpiredCarts
type cartRepositoryAdapter struct {
	*redis.CartRepository
	lockManager *redis.LockManager
}

// GetExpiredCarts реализация метода для интерфейса repositories.CartRepository
func (a *cartRepositoryAdapter) GetExpiredCarts(ctx context.Context) ([]models.Cart, error) {
	// Заглушка для метода GetExpiredCarts
	return []models.Cart{}, nil
}
