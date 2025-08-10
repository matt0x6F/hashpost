package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Logging  LoggingConfig
	IBE      IBEConfig
	JWT      JWTConfig
	Security SecurityConfig
	CORS     CORSConfig
	Email    EmailConfig
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	SiteURL      string // Base URL for the application (e.g., "https://hashpost.com")
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
}

// IBEConfig holds Identity-Based Encryption configuration
type IBEConfig struct {
	DomainKeysDir string // Directory containing domain-specific master keys
	KeyVersion    int32  // Current key version
	Salt          string // Salt for fingerprint generation (defaults to "fingerprint_salt_v1")
	KeyRotation   struct {
		Enabled     bool
		Interval    time.Duration
		GracePeriod time.Duration
	}
	// Role-to-domain mapping for encryption
	RoleDomainMapping map[string]string // Maps role names to domain names
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret      string
	Expiration  time.Duration
	Development bool // Controls cookie security settings
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableMFA bool // Controls whether MFA requirements are enforced

	// Password validation settings
	PasswordValidation PasswordValidationConfig
}

// PasswordValidationConfig holds password validation rules
type PasswordValidationConfig struct {
	MinLength          int  // Minimum password length
	RequireUppercase   bool // Require at least one uppercase letter
	RequireLowercase   bool // Require at least one lowercase letter
	RequireDigit       bool // Require at least one digit
	RequireSpecialChar bool // Require at least one special character
	DisallowCommon     bool // Disallow common passwords
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider    string // "mailgun" or other providers in the future
	FromAddress string
	FromName    string
	MailGun     MailGunConfig
	Validation  EmailValidationConfig
}

// MailGunConfig holds MailGun-specific configuration
type MailGunConfig struct {
	Domain        string
	SendingAPIKey string // API key for sending emails
	BaseURL       string
	Region        string // "us" or "eu"
}

// EmailValidationConfig holds email validation configuration
type EmailValidationConfig struct {
	Enabled            bool     // Enable email validation
	VerifierEmail      string   // Email address to use for SMTP verification
	VerifierPassword   string   // Password for the verifier email
	VerifierSMTPHost   string   // SMTP host for verification (e.g., "smtp.gmail.com")
	VerifierSMTPPort   int      // SMTP port for verification (e.g., 587)
	VerifierSMTPUser   string   // SMTP username (usually same as email)
	ConnectionTimeout  int      // Connection timeout in seconds
	ResponseTimeout    int      // Response timeout in seconds
	SmtpFailFast       bool     // Fail fast for better UX
	SmtpSafeCheck      bool     // Safe check to avoid false negatives
	ValidationLevel    string   // "basic", "mx", "smtp" - validation level to use
	BlacklistedDomains []string // Domains to blacklist (e.g., disposable email providers)
	BlacklistedMxIPs   []string // MX IP addresses to blacklist
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "hashpost"),
			Password:        getEnv("DB_PASSWORD", "hashpost_dev"),
			Database:        getEnv("DB_NAME", "hashpost"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8888),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			SiteURL:      getEnv("SERVER_SITE_URL", "https://hashpost.com"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", ""),
		},
		IBE: IBEConfig{
			DomainKeysDir: getEnv("IBE_DOMAIN_KEYS_DIR", "./keys/domains"),
			KeyVersion:    int32(getEnvAsInt("IBE_KEY_VERSION", 1)),
			Salt:          getEnv("IBE_SALT", "fingerprint_salt_v1"),
			KeyRotation: struct {
				Enabled     bool
				Interval    time.Duration
				GracePeriod time.Duration
			}{
				Enabled:     getEnvAsBool("IBE_KEY_ROTATION_ENABLED", false),
				Interval:    getEnvAsDuration("IBE_KEY_ROTATION_INTERVAL", 365*24*time.Hour),    // 1 year
				GracePeriod: getEnvAsDuration("IBE_KEY_ROTATION_GRACE_PERIOD", 30*24*time.Hour), // 30 days
			},
			RoleDomainMapping: map[string]string{
				"user":           "user_self_correlation_v1",
				"moderator":      "moderator_correlation_v1",
				"subforum_owner": "moderator_correlation_v1",
				"platform_admin": "admin_correlation_v1",
				"trust_safety":   "admin_correlation_v1",
				"legal_team":     "legal_correlation_v1",
			},
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
			Expiration:  getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			Development: getEnvAsBool("JWT_DEVELOPMENT", true),
		},
		Security: SecurityConfig{
			EnableMFA: getEnvAsBool("SECURITY_ENABLE_MFA", false),
			PasswordValidation: PasswordValidationConfig{
				MinLength:          getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
				RequireUppercase:   getEnvAsBool("PASSWORD_REQUIRE_UPPERCASE", true),
				RequireLowercase:   getEnvAsBool("PASSWORD_REQUIRE_LOWERCASE", true),
				RequireDigit:       getEnvAsBool("PASSWORD_REQUIRE_DIGIT", true),
				RequireSpecialChar: getEnvAsBool("PASSWORD_REQUIRE_SPECIAL", true),
				DisallowCommon:     getEnvAsBool("PASSWORD_DISALLOW_COMMON", true),
			},
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Authorization", "Content-Type"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 300),
		},
		Email: EmailConfig{
			Provider:    getEnv("EMAIL_PROVIDER", "mailgun"),
			FromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@hashpost.dev"),
			FromName:    getEnv("EMAIL_FROM_NAME", "HashPost"),
			MailGun: MailGunConfig{
				Domain:        getEnv("MAILGUN_DOMAIN", ""),
				SendingAPIKey: getEnv("MAILGUN_SENDING_API_KEY", ""),
				BaseURL:       getEnv("MAILGUN_BASE_URL", "https://api.mailgun.net"),
				Region:        getEnv("MAILGUN_REGION", "us"),
			},
			Validation: EmailValidationConfig{
				Enabled:            getEnvAsBool("EMAIL_VALIDATION_ENABLED", false),
				VerifierEmail:      getEnv("EMAIL_VALIDATION_VERIFIER_EMAIL", "noreply@hashpost.dev"),
				VerifierPassword:   getEnv("EMAIL_VALIDATION_VERIFIER_PASSWORD", ""),
				VerifierSMTPHost:   getEnv("EMAIL_VALIDATION_SMTP_HOST", "smtp.mailgun.org"),
				VerifierSMTPPort:   getEnvAsInt("EMAIL_VALIDATION_SMTP_PORT", 587),
				VerifierSMTPUser:   getEnv("EMAIL_VALIDATION_SMTP_USER", ""),
				ConnectionTimeout:  getEnvAsInt("EMAIL_VALIDATION_CONNECTION_TIMEOUT", 5),
				ResponseTimeout:    getEnvAsInt("EMAIL_VALIDATION_RESPONSE_TIMEOUT", 2),
				SmtpFailFast:       getEnvAsBool("EMAIL_VALIDATION_SMTP_FAIL_FAST", false),
				SmtpSafeCheck:      getEnvAsBool("EMAIL_VALIDATION_SMTP_SAFE_CHECK", false),
				ValidationLevel:    getEnv("EMAIL_VALIDATION_LEVEL", "basic"),
				BlacklistedDomains: getEnvAsSlice("EMAIL_VALIDATION_BLACKLISTED_DOMAINS", []string{}),
				BlacklistedMxIPs:   getEnvAsSlice("EMAIL_VALIDATION_BLACKLISTED_MX_IPS", []string{}),
			},
		},
	}

	// If DATABASE_URL is provided, parse it to override individual settings
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		if err := config.Database.ParseURL(dbURL); err != nil {
			return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
		}
	}

	return config, nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// ParseURL parses a PostgreSQL connection URL and updates the config
func (c *DatabaseConfig) ParseURL(dbURL string) error {
	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("invalid database URL: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	if u.User != nil {
		c.User = u.User.Username()
		if password, ok := u.User.Password(); ok {
			c.Password = password
		}
	}

	c.Host = u.Hostname()

	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}

	c.Database = ""
	if len(u.Path) > 1 {
		c.Database = u.Path[1:]
	}

	// Parse query parameters for sslmode
	q := u.Query()
	if sslmode := q.Get("sslmode"); sslmode != "" {
		c.SSLMode = sslmode
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
