package pkg

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Ticket Tailor
	TicketTailorAPIKey  string `env:"TICKETTAILOR_API_KEY,required"`
	TicketTailorEventId string `env:"TT_EVENT_ID,required"`
	TicketTailorBaseUrl string `env:"TT_BASE_URL,required"`

	// Apple Pass
	AppleP12Path     string `env:"APPLE_P12_PATH"`
	AppleP12Password string `env:"APPLE_P12_PASSWORD,required"`
	AppleP12Base64   string `env:"APPLE_P12_BASE64"`

	AppleRootCertPath string `env:"APPLE_ROOT_CERT_PATH"`
	AppleRootBase64   string `env:"APPLE_ROOT_CERT_BASE64"`

	ApplePassTypeID string `env:"APPLE_PASS_TYPE_IDENTIFIER,required"`
	AppleTeamID     string `env:"APPLE_TEAM_IDENTIFIER,required"`

	// App
	BatchCron  string `env:"BATCH_CRON" envDefault:"@every 5m"`
	DataDir    string `env:"DATA_DIR" envDefault:"/app/data"`
	Port       int    `env:"PORT" envDefault:"8080"`
	TicketsDir string `env:"TICKETS_DIR" envDefault:"tickets"`

	// Database (raw inputs)
	DatabaseURL              string        `env:"DATABASE_URL,required"`
	DatabaseMaxOpenConns     int           `env:"DATABASE_MAX_OPEN_CONNS" envDefault:"10"`
	DatabaseMaxIdleConns     int           `env:"DATABASE_MAX_IDLE_CONNS" envDefault:"5"`
	DatabaseConnMaxLifetime  time.Duration `env:"DATABASE_CONN_MAX_LIFETIME" envDefault:"30m"`
	DatabaseConnMaxIdleTime  time.Duration `env:"DATABASE_CONN_MAX_IDLE_TIME" envDefault:"15m"`
	DatabaseLogLevel         string        `env:"DATABASE_LOG_LEVEL" envDefault:"warn"`
	DatabasePreferSimpleProt bool          `env:"DATABASE_PREFER_SIMPLE_PROTOCOL" envDefault:"false"`

	database DatabaseConfig
}

// TODO: could be better, this has mutation...
// Validate ensures required configuration is present and derives structured settings.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must use postgres scheme, got %q", parsed.Scheme)
	}

	query := parsed.Query()
	sslmode := query.Get("sslmode")
	if sslmode == "" {
		return fmt.Errorf("DATABASE_URL must include sslmode param (set to require for Neon)")
	}
	if strings.EqualFold(sslmode, "disable") {
		return fmt.Errorf("DATABASE_URL sslmode=disable is not supported for Neon")
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("DATABASE_URL must include host")
	}

	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("DATABASE_URL must include valid port, got %q", port)
	}

	maxOpen := c.DatabaseMaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := c.DatabaseMaxIdleConns
	if maxIdle < 0 {
		maxIdle = 0
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	logLevel := strings.ToLower(strings.TrimSpace(c.DatabaseLogLevel))
	switch logLevel {
	case "":
		logLevel = "warn"
	case "silent", "error", "warn", "info":
		// ok
	default:
		return fmt.Errorf("DATABASE_LOG_LEVEL must be one of silent|error|warn|info, got %q", c.DatabaseLogLevel)
	}

	c.database = DatabaseConfig{
		DSN:                  c.DatabaseURL,
		MaxOpenConns:         maxOpen,
		MaxIdleConns:         maxIdle,
		ConnMaxLifetime:      c.DatabaseConnMaxLifetime,
		ConnMaxIdleTime:      c.DatabaseConnMaxIdleTime,
		LogLevel:             logLevel,
		PreferSimpleProtocol: c.DatabasePreferSimpleProt,
		Host:                 host,
		Port:                 port,
		SSLMode:              sslmode,
	}

	return nil
}

// DatabaseConfig captures derived database settings after validation.
type DatabaseConfig struct {
	DSN                  string
	MaxOpenConns         int
	MaxIdleConns         int
	ConnMaxLifetime      time.Duration
	ConnMaxIdleTime      time.Duration
	LogLevel             string
	PreferSimpleProtocol bool
	Host                 string
	Port                 string
	SSLMode              string
}

// Database exposes validated database settings. Call Validate() first.
func (c Config) Database() DatabaseConfig {
	return c.database
}
