package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

func TestOpenAgainstMigratedDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	connStr, cleanup := startTestPostgres(t, ctx)
	t.Cleanup(cleanup)

	testCases := []struct {
		name string
		cfg  DatabaseConfig
	}{
		{
			name: "default",
			cfg: DatabaseConfig{
				DSN:             connStr,
				MaxOpenConns:    5,
				MaxIdleConns:    2,
				ConnMaxLifetime: 30 * time.Minute,
				ConnMaxIdleTime: 10 * time.Minute,
				LogLevel:        logger.Warn,
			},
		},
		{
			name: "prefer simple protocol",
			cfg: DatabaseConfig{
				DSN:                  connStr,
				MaxOpenConns:         3,
				MaxIdleConns:         3,
				ConnMaxLifetime:      time.Minute,
				ConnMaxIdleTime:      30 * time.Second,
				LogLevel:             logger.Info,
				PreferSimpleProtocol: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := Open(ctx, tc.cfg)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, Close(conn))
			})

			sqlDB, err := conn.DB()
			require.NoError(t, err)
			stats := sqlDB.Stats()
			require.Equal(t, tc.cfg.MaxOpenConns, stats.MaxOpenConnections)
		})
	}
}

func TestOpenRequiresDSN(t *testing.T) {
	_, err := Open(context.Background(), DatabaseConfig{})
	require.Error(t, err)
}

func TestCloseNilIsSafe(t *testing.T) {
	require.NoError(t, Close(nil))
}
