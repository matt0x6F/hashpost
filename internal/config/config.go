package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	PDS     PDSConfig     `yaml:"pds"`
	AppView AppViewConfig `yaml:"appview"`
	Logging LoggingConfig `yaml:"logging"`
}

// PDSConfig represents PDS-specific configuration
type PDSConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	NATS     NATSConfig     `yaml:"nats"`
	Atproto  AtprotoConfig  `yaml:"atproto"`
}

// AppViewConfig represents AppView-specific configuration
type AppViewConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	PDS      PDSURLConfig   `yaml:"pds"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Dev  bool   `yaml:"dev"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	URL      string `yaml:"url"` // Direct database URL for environment overrides
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL string `yaml:"url"`
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL string `yaml:"url"`
}

// AtprotoConfig represents atproto-specific configuration
type AtprotoConfig struct {
	DIDResolver string `yaml:"did_resolver"`
	HandleBase  string `yaml:"handle_base"`
}

// PDSURLConfig represents PDS URL configuration for AppView
type PDSURLConfig struct {
	URL string `yaml:"url"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables if present
	overrideFromEnv(&config)

	return &config, nil
}

// overrideFromEnv overrides configuration values with environment variables
func overrideFromEnv(config *Config) {
	if host := os.Getenv("PDS_HOST"); host != "" {
		config.PDS.Server.Host = host
	}
	if port := os.Getenv("PDS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.PDS.Server.Port = p
		}
	}
	if host := os.Getenv("APPVIEW_HOST"); host != "" {
		config.AppView.Server.Host = host
	}
	if port := os.Getenv("APPVIEW_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.AppView.Server.Port = p
		}
	}
	// Handle PDS database URL
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.PDS.Database.URL = dbURL
		// Clear individual fields to indicate URL is used
		config.PDS.Database.Host = ""
		config.PDS.Database.Port = 0
		config.PDS.Database.Database = ""
		config.PDS.Database.Username = ""
		config.PDS.Database.Password = ""
		config.PDS.Database.SSLMode = ""
	}

	// Handle AppView database URL
	if appViewDBURL := os.Getenv("APPVIEW_DATABASE_URL"); appViewDBURL != "" {
		config.AppView.Database.URL = appViewDBURL
		// Clear individual fields to indicate URL is used
		config.AppView.Database.Host = ""
		config.AppView.Database.Port = 0
		config.AppView.Database.Database = ""
		config.AppView.Database.Username = ""
		config.AppView.Database.Password = ""
		config.AppView.Database.SSLMode = ""
	}
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.PDS.Redis.URL = redisURL
	}
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		config.PDS.NATS.URL = natsURL
	}
	if pdsURL := os.Getenv("PDS_URL"); pdsURL != "" {
		config.AppView.PDS.URL = pdsURL
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	// If URL is set (from environment), use it directly
	if c.PDS.Database.URL != "" {
		return c.PDS.Database.URL
	}
	// Otherwise, construct from individual fields
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PDS.Database.Username,
		c.PDS.Database.Password,
		c.PDS.Database.Host,
		c.PDS.Database.Port,
		c.PDS.Database.Database,
		c.PDS.Database.SSLMode,
	)
}

// GetServerAddress returns the server address for PDS
func (c *Config) GetPDSServerAddress() string {
	return fmt.Sprintf("%s:%d", c.PDS.Server.Host, c.PDS.Server.Port)
}

// GetServerAddress returns the server address for AppView
func (c *Config) GetAppViewServerAddress() string {
	return fmt.Sprintf("%s:%d", c.AppView.Server.Host, c.AppView.Server.Port)
}

// GetAppViewDatabaseURL returns the AppView database connection URL
func (c *Config) GetAppViewDatabaseURL() string {
	// If URL is set (from environment), use it directly
	if c.AppView.Database.URL != "" {
		return c.AppView.Database.URL
	}
	// Otherwise, construct from individual fields
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.AppView.Database.Username,
		c.AppView.Database.Password,
		c.AppView.Database.Host,
		c.AppView.Database.Port,
		c.AppView.Database.Database,
		c.AppView.Database.SSLMode,
	)
}

// GetNATSURL returns the NATS connection URL with a default fallback
func (c *Config) GetNATSURL() string {
	if c.PDS.NATS.URL != "" {
		return c.PDS.NATS.URL
	}
	return "nats://localhost:4222"
}
