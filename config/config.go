package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config contains all application settings
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string
	Environment string
	LogLevel    string
}

// HTTPConfig contains HTTP server settings
type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	Timeout  time.Duration
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
}

// JWTConfig contains JWT token settings
type JWTConfig struct {
	Secret            string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CartExpirationTTL time.Duration
}

// LoadConfig loads configuration from environment variables
// For local development, it will try to load .env file first
func LoadConfig() (Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	return Config{
		App:      loadAppConfig(),
		HTTP:     loadHTTPConfig(),
		Database: loadDatabaseConfig(),
		Redis:    loadRedisConfig(),
		JWT:      loadJWTConfig(),
	}, nil
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Name:        getEnv("APP_NAME", "bookshop-api"),
		Environment: getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func loadHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Host:         getEnv("HTTP_HOST", "0.0.0.0"),
		Port:         getEnvAsInt("HTTP_PORT", 8080),
		ReadTimeout:  time.Duration(getEnvAsInt("HTTP_READ_TIMEOUT_SECONDS", 5)) * time.Second,
		WriteTimeout: time.Duration(getEnvAsInt("HTTP_WRITE_TIMEOUT_SECONDS", 5)) * time.Second,
		IdleTimeout:  time.Duration(getEnvAsInt("HTTP_IDLE_TIMEOUT_SECONDS", 120)) * time.Second,
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "bookshop"),
		Password: getEnv("DB_PASSWORD", "bookshop"),
		DBName:   getEnv("DB_NAME", "bookshop"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		MaxConns: getEnvAsInt("DB_MAX_CONNS", 10),
		Timeout:  time.Duration(getEnvAsInt("DB_TIMEOUT_SECONDS", 5)) * time.Second,
	}
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnvAsInt("REDIS_PORT", 6379),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
		Timeout:  time.Duration(getEnvAsInt("REDIS_TIMEOUT_SECONDS", 5)) * time.Second,
	}
}

func loadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:            getEnv("JWT_SECRET", "app-secret-key-change-in-production"),
		AccessTokenTTL:    time.Duration(getEnvAsInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:   time.Duration(getEnvAsInt("JWT_REFRESH_TOKEN_TTL_DAYS", 7)) * 24 * time.Hour,
		CartExpirationTTL: time.Duration(getEnvAsInt("JWT_CART_EXPIRATION_TTL_HOURS", 24)) * time.Hour,
	}
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetDSN returns PostgreSQL connection string
func (c DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr returns Redis connection address
func (c RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
