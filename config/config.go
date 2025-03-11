package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config содержит все настройки приложения
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// AppConfig содержит общие настройки приложения
type AppConfig struct {
	Name        string
	Environment string
	LogLevel    string
}

// HTTPConfig содержит настройки HTTP сервера
type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig содержит настройки подключения к базе данных
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

// RedisConfig содержит настройки подключения к Redis
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
}

// JWTConfig содержит настройки для JWT токенов
type JWTConfig struct {
	Secret            string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CartExpirationTTL time.Duration
}

// LoadConfig загружает конфигурацию из файла и переменных окружения
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка десериализации конфигурации: %w", err)
	}

	return &cfg, nil
}

// GetDSN возвращает строку подключения к PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr возвращает адрес подключения к Redis
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
