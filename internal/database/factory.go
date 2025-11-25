package database

import (
	"fmt"

	"github.com/smotra-monitoring/server/internal/config"
)

// NewFromConfig creates a new database instance from application config
func NewFromConfig(cfg config.DatabaseConfig) (Database, error) {
	switch cfg.Type {
	case "postgres":
		pgConfig := PostgresConfig{
			Config: Config{
				Type:            cfg.Type,
				MaxOpenConns:    cfg.MaxOpenConns,
				MaxIdleConns:    cfg.MaxIdleConns,
				ConnMaxLifetime: cfg.ConnMaxLifetime,
				ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			},
			Host:     cfg.Host,
			Port:     cfg.Port,
			Username: cfg.Username,
			Password: cfg.Password,
			Database: cfg.Database,
			SSLMode:  cfg.SSLMode,
		}
		return NewWithPostgres(pgConfig), nil
	case "sqlite":
		sqliteConfig := SQLiteConfig{
			Config: Config{
				Type:            cfg.Type,
				MaxOpenConns:    cfg.MaxOpenConns,
				MaxIdleConns:    cfg.MaxIdleConns,
				ConnMaxLifetime: cfg.ConnMaxLifetime,
				ConnMaxIdleTime: cfg.ConnMaxIdleTime,
			},
			FilePath: cfg.FilePath,
		}
		return NewWithSQLite(sqliteConfig), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

// NewWithPostgres creates a new PostgreSQL database instance
func NewWithPostgres(config PostgresConfig) Database {
	return NewPostgresDB(config)
}

// NewWithSQLite creates a new SQLite database instance
func NewWithSQLite(config SQLiteConfig) Database {
	return NewSQLiteDB(config)
}
