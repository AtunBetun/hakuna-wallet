package pkg

type Config struct {
	// Ticket Tailor
	TicketTailorAPIKey  string `env:"TICKETTAILOR_API_KEY,required"`
	TicketTailorEventId string `env:"TT_EVENT_ID,required"`
	TicketTailorBaseUrl string `env:"TT_BASE_URL,required"`

	// Database
	DatabaseURL string `env:"DATABASE_URL,required"`

	// Apple Pass
	AppleP12Path     string `env:"APPLE_P12_PATH,required"`
	AppleP12Password string `env:"APPLE_P12_PASSWORD,required"`
	AppleRootCert    string `env:"APPLE_ROOT_CERT_PATH,required"`
	ApplePassTypeID  string `env:"APPLE_PASS_TYPE_IDENTIFIER,required"`
	AppleTeamID      string `env:"APPLE_TEAM_IDENTIFIER,required"`

	// Google Wallet
	GoogleServiceAccountJSON string `env:"GOOGLE_SERVICE_ACCOUNT_JSON,required"`
	GoogleIssuerEmail        string `env:"GOOGLE_ISSUER_EMAIL,required"`

	// App
	BatchCron  string `env:"BATCH_CRON" envDefault:"@every 5m"`
	DataDir    string `env:"DATA_DIR" envDefault:"/app/data"`
	Port       int    `env:"PORT" envDefault:"8080"`
	TicketsDir string `env:"TICKETS_DIR" envDefault:"tickets"`
}
