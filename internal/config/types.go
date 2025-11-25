package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/smotra-monitoring/server/internal/database"
)

// Config holds the application configuration
type Config struct {
	Server         ServerConfig
	DatabaseType   string // postgres or sqlite
	PostgresConfig *database.PostgresConfig
	SQLiteConfig   *database.SQLiteConfig
	Logging        LoggingConfig
	Auth           AuthConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Environment     string // development, staging, production
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	// OAuth2 settings can be added here
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	dbType := getEnv("DB_TYPE", "sqlite")

	cfg := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
		DatabaseType: dbType,
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", ""),
			JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
	}

	// Initialize database config based on type
	if dbType == "postgres" {
		cfg.PostgresConfig = &database.PostgresConfig{
			Config: database.Config{
				Type:            dbType,
				MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
				ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
			},
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "smotra"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		}
	} else {
		cfg.SQLiteConfig = &database.SQLiteConfig{
			Config: database.Config{
				Type:            dbType,
				MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
				ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
			},
			FilePath: getEnv("DB_FILEPATH", "./data/smotra.db"),
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Database validation
	if c.DatabaseType != "postgres" && c.DatabaseType != "sqlite" {
		return fmt.Errorf("invalid database type: %s (must be 'postgres' or 'sqlite')", c.DatabaseType)
	}

	if c.DatabaseType == "postgres" {
		if c.PostgresConfig == nil {
			return fmt.Errorf("postgres config is required when database type is postgres")
		}
		if c.PostgresConfig.Host == "" {
			return fmt.Errorf("database host is required for postgres")
		}
		if c.PostgresConfig.Username == "" {
			return fmt.Errorf("database username is required for postgres")
		}
		if c.PostgresConfig.Database == "" {
			return fmt.Errorf("database name is required for postgres")
		}
	}

	if c.DatabaseType == "sqlite" {
		if c.SQLiteConfig == nil {
			return fmt.Errorf("sqlite config is required when database type is sqlite")
		}
		if c.SQLiteConfig.FilePath == "" {
			return fmt.Errorf("database file path is required for sqlite")
		}
	}

	// Logging validation
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	validLogFormats := map[string]bool{"json": true, "text": true}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	// Auth validation (warning only for development)
	if c.Auth.JWTSecret == "" && c.Server.Environment == "production" {
		return fmt.Errorf("JWT secret is required in production")
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration retrieves an environment variable as a duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
