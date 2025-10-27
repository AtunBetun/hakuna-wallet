package pkg

import (
	"os"
	"time"

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
)

type AppConfig struct {
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

	ApplePassword string `env:"APPLE_PASSWORD,required"`

	TicketsDir string `env:"TICKETS_DIR" envDefault:"tickets"`

	// Database (raw inputs)
	DatabaseURL                  string        `env:"DATABASE_URL,required"`
	DatabaseMaxOpenConns         int           `env:"DATABASE_MAX_OPEN_CONNS" envDefault:"10"`
	DatabaseMaxIdleConns         int           `env:"DATABASE_MAX_IDLE_CONNS" envDefault:"5"`
	DatabaseConnMaxLifetime      time.Duration `env:"DATABASE_CONN_MAX_LIFETIME" envDefault:"30m"`
	DatabaseConnMaxIdleTime      time.Duration `env:"DATABASE_CONN_MAX_IDLE_TIME" envDefault:"15m"`
	DatabaseLogLevel             string        `env:"DATABASE_LOG_LEVEL" envDefault:"warn"`
	DatabasePreferSimpleProtocol bool          `env:"DATABASE_PREFER_SIMPLE_PROTOCOL" envDefault:"false"`

	// Aws
	S3Bucket           string `env:"S3_BUCKET,required" envDefault:"hakuna-use1"`
	AwsRegion          string `env:"AWS_REGION,required" envDefault:"us-east-1"`
	awsAccessKey       string `env:"AWS_ACCESS_KEY,required"`
	awsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY,required"`
}

func ShouldLoadDotenv() bool {
	env, found := os.LookupEnv("APP_ENV")
	if found && env == "prod" {
		logger.Logger.Info("Not loading .env")
		return false
	}
	return true
}
