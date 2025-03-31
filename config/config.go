package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contains all application settings
type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	RateLimit RateLimiterConfig
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

// RateLimiterConfig contains rate limiter settings
type RateLimiterConfig struct {
	Enabled          bool
	GlobalIPLimit    int            // Global rate limit per IP per second
	DefaultPathLimit int            // Default rate limit per path per second
	CleanupInterval  time.Duration  // Interval for cleaning up stale limiters
	Endpoints        map[string]int // Custom rate limits for specific endpoints
}

// LoadConfig loads configuration from environment variables
// For local development, it will try to load .env file first
func LoadConfig() (Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	return Config{
		App:       loadAppConfig(),
		HTTP:      loadHTTPConfig(),
		Database:  loadDatabaseConfig(),
		Redis:     loadRedisConfig(),
		JWT:       loadJWTConfig(),
		RateLimit: loadRateLimiterConfig(),
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

func loadRateLimiterConfig() RateLimiterConfig {
	// Parse endpoint-specific rate limits
	endpointLimits := make(map[string]int)
	endpointConfig := getEnv("RATE_LIMIT_ENDPOINTS", "/api/v1/checkout=20,/api/v1/orders=50,/api/v1/admin/*=10")

	if endpointConfig != "" {
		pairs := strings.Split(endpointConfig, ",")
		for _, pair := range pairs {
			if kv := strings.Split(pair, "="); len(kv) == 2 {
				path := strings.TrimSpace(kv[0])
				if limit, err := strconv.Atoi(strings.TrimSpace(kv[1])); err == nil {
					endpointLimits[path] = limit
				}
			}
		}
	}

	return RateLimiterConfig{
		Enabled:          getEnvAsBool("RATE_LIMIT_ENABLED", true),
		GlobalIPLimit:    getEnvAsInt("RATE_LIMIT_GLOBAL_IP", 100),    // 100 requests per second per IP
		DefaultPathLimit: getEnvAsInt("RATE_LIMIT_DEFAULT_PATH", 200), // 200 requests per second per path
		CleanupInterval:  time.Duration(getEnvAsInt("RATE_LIMIT_CLEANUP_MINUTES", 5)) * time.Minute,
		Endpoints:        endpointLimits,
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
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
