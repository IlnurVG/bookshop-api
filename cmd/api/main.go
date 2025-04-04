package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bookshop/api/config"
	"github.com/bookshop/api/internal/app/cart"
	"github.com/bookshop/api/internal/app/checkout"
	"github.com/bookshop/api/internal/repository/postgres"
	"github.com/bookshop/api/internal/repository/redis"
	"github.com/bookshop/api/internal/server"
	"github.com/bookshop/api/internal/service"
	"github.com/bookshop/api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize logger
	l, err := logger.NewLogger(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer l.Sync()

	// Initialize PostgreSQL connection
	db, err := postgres.NewPostgresDB(cfg.Database)
	if err != nil {
		l.Fatal("Database connection error", err)
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		l.Fatal("Redis connection error", err)
	}
	defer redisClient.Close()

	// Initialize transaction manager
	txManager := postgres.NewTransactionManager(db)

	// Initialize repositories
	bookRepo := postgres.NewBookRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	userRepo := postgres.NewUserRepository(db)
	cartRepo := redis.NewCartRepository(redisClient)

	// Log wrapper for modules
	log := logger.Logger(*l)

	// Initialize profile cache service with worker pool
	profileCacheService := service.NewProfileCacheService(
		userRepo,
		orderRepo,
		redisClient,
		log,
	)

	// Initialize checkout module
	checkoutModule := checkout.NewModule(
		orderRepo,
		cartRepo,
		bookRepo,
		txManager,
		log,
		profileCacheService,
	)

	// Initialize cart module
	cartModule := cart.NewModule(
		cartRepo,
		bookRepo,
		txManager,
		log,
	)

	// Initialize server with dependencies
	srv, err := server.NewServer(
		&cfg,
		l,
		checkoutModule.Service,
		cartModule.Service,
		bookRepo,
		categoryRepo,
		txManager,
	)
	if err != nil {
		l.Fatal("Server initialization error", err)
	}

	// Start server in a goroutine
	go func() {
		l.Info("Starting HTTP server", "address", srv.Addr)
		if err := srv.Start(); err != nil {
			l.Error("Server start error", err)
		}
	}()

	// Wait for termination signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down the server...")

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the profile cache service
	profileCacheService.Shutdown()

	// Stop the server
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("Error during server shutdown", err)
	}

	l.Info("Server successfully stopped")
}
