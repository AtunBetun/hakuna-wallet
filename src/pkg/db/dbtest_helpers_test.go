package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

func startTestPostgres(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()

	container, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("walletdb"),
		tcpostgres.WithUsername("wallet"),
		tcpostgres.WithPassword("walletpass"),
	)
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, container.Terminate(context.Background()))
	}

	rawConnStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	connStr := forceIPv4(t, rawConnStr)

	waitForDatabase(t, connStr)
	applyMigrations(t, connStr)

	return connStr, cleanup
}

func setupTestDatabase(t *testing.T, ctx context.Context) *gorm.DB {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	t.Cleanup(cancel)

	connStr, cleanup := startTestPostgres(t, ctx)
	t.Cleanup(cleanup)

	cfg := Config{
		DSN:             connStr,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
		LogLevel:        "warn",
	}

	conn, err := Open(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, Close(conn))
	})

	return conn
}

func applyMigrations(t *testing.T, connStr string) {
	t.Helper()

	migrationsPath := migrationsDir(t)
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), connStr)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = m.Close()
	})

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatalf("running migrations: %v", err)
	}
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "migrations")
	abs, err := filepath.Abs(path)
	require.NoError(t, err, "resolve migrations path")
	return abs
}

func forceIPv4(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err, "parse connection string")
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	if host == "localhost" || host == "::1" {
		u.Host = net.JoinHostPort("127.0.0.1", port)
	}
	return u.String()
}

func waitForDatabase(t *testing.T, connStr string) {
	t.Helper()

	require.Eventually(t, func() bool {
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			return false
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "reset") {
				time.Sleep(50 * time.Millisecond)
				return false
			}
			return false
		}
		return true
	}, 20*time.Second, 100*time.Millisecond, "database not ready in time")
}
