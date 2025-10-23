package db

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg"
	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	DSN                  string           `validate:"required"`
	MaxOpenConns         int              `validate:"required"`
	MaxIdleConns         int              `validate:"required"`
	ConnMaxLifetime      time.Duration    `validate:"required"`
	ConnMaxIdleTime      time.Duration    `validate:"required"`
	LogLevel             gormLog.LogLevel `validate:"required"`
	PreferSimpleProtocol bool
	Host                 string `validate:"required"`
	Port                 string `validate:"required"`
	SSLMode              string `validate:"required"`
}

func FromAppConfig(cfg pkg.AppConfig) (DatabaseConfig, error) {
	maxOpen := cfg.DatabaseMaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := cfg.DatabaseMaxIdleConns
	if maxIdle < 0 {
		maxIdle = 0
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	logLevel := parseLogLevel(cfg.DatabaseLogLevel)

	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must use postgres scheme, got %q", parsed.Scheme)
	}

	query := parsed.Query()
	sslmode := query.Get("sslmode")
	if sslmode == "" {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must include sslmode param (set to require for Neon)")
	}
	if strings.EqualFold(sslmode, "disable") {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL sslmode=disable is not supported for Neon")
	}

	host := parsed.Hostname()
	if host == "" {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must include host")
	}

	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL must include valid port, got %q", port)
	}

	fmt.Println("HERE")

	database := DatabaseConfig{
		DSN:                  cfg.DatabaseURL,
		MaxOpenConns:         maxOpen,
		MaxIdleConns:         maxIdle,
		ConnMaxLifetime:      cfg.DatabaseConnMaxLifetime,
		ConnMaxIdleTime:      cfg.DatabaseConnMaxIdleTime,
		LogLevel:             logLevel,
		PreferSimpleProtocol: cfg.DatabasePreferSimpleProtocol,
		Host:                 host,
		Port:                 port,
		SSLMode:              sslmode,
	}
	fmt.Println("THERE")
	fmt.Printf("%+v", database)

	err = validate.Struct(database)
	if err != nil {
		return DatabaseConfig{}, err
	}
	fmt.Println("AFTER")

	return database, nil

}

// Database exposes validated database settings. Call Validate() first.

var validate = validator.New(validator.WithRequiredStructEnabled())

// Open establishes a gorm.DB connection using the provided settings.
func Open(ctx context.Context, cfg DatabaseConfig) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	gormCfg := &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormLog.Default.LogMode(cfg.LogLevel),
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

func parseLogLevel(level string) gormLog.LogLevel {
	switch level {
	case "silent":
		return gormLog.Silent
	case "error":
		return gormLog.Error
	case "info":
		return gormLog.Info
	default:
		return gormLog.Warn
	}
}
