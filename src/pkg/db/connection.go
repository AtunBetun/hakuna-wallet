package db

import (
	"context"
	"fmt"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config represents the minimal settings required to open a database connection.
type Config struct {
	DSN                  string
	MaxOpenConns         int
	MaxIdleConns         int
	ConnMaxLifetime      time.Duration
	ConnMaxIdleTime      time.Duration
	LogLevel             string
	PreferSimpleProtocol bool
}

// Open establishes a gorm.DB connection using the provided settings.
func Open(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	gormCfg := &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(parseLogLevel(cfg.LogLevel)),
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN,
		PreferSimpleProtocol: cfg.PreferSimpleProtocol,
	}), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("opening database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("retrieving sql.DB handle: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns >= 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db.WithContext(ctx), nil
}

// Close closes the underlying sql.DB connection.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("retrieving sql.DB handle: %w", err)
	}
	return sqlDB.Close()
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

// FromAppConfig converts the application config into a database config.
func FromAppConfig(cfg pkg.Config) Config {
	dbCfg := cfg.Database()
	return Config{
		DSN:                  dbCfg.DSN,
		MaxOpenConns:         dbCfg.MaxOpenConns,
		MaxIdleConns:         dbCfg.MaxIdleConns,
		ConnMaxLifetime:      dbCfg.ConnMaxLifetime,
		ConnMaxIdleTime:      dbCfg.ConnMaxIdleTime,
		LogLevel:             dbCfg.LogLevel,
		PreferSimpleProtocol: dbCfg.PreferSimpleProtocol,
	}
}
