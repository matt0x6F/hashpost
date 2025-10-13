package pds

import (
	"context"
	"testing"

	"github.com/matt0x6f/hashpost/internal/config"
	pds "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		PDS: config.PDSConfig{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
				Dev:  true,
			},
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "hashpost_test",
				Username: "hashpost",
				Password: "password",
				SSLMode:  "disable",
			},
			Redis: config.RedisConfig{
				URL: "redis://localhost:6379",
			},
			Atproto: config.AtprotoConfig{
				DIDResolver: "https://plc.directory",
				HandleBase:  "hashpost.local",
			},
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
	}

	// Create mock database queries
	// In a real test, you would use a test database
	var queries *pds.Queries

	// Test server creation
	server, err := NewServer(cfg, queries)
	require.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, queries, server.db)
}

func TestServerConfig(t *testing.T) {
	cfg := &config.Config{
		PDS: config.PDSConfig{
			Server: config.ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
		},
	}

	server := &Server{
		config: cfg,
	}

	// Test server address generation
	expected := "0.0.0.0:8080"
	actual := server.config.GetPDSServerAddress()
	assert.Equal(t, expected, actual)
}

func TestServerLifecycle(t *testing.T) {
	// This is a basic test to ensure the server can be created and stopped
	// In a real test environment, you would test with actual HTTP requests

	cfg := &config.Config{
		PDS: config.PDSConfig{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 0, // Use port 0 for testing
				Dev:  true,
			},
			Atproto: config.AtprotoConfig{
				DIDResolver: "https://plc.directory",
				HandleBase:  "hashpost.local",
			},
		},
	}

	var queries *pds.Queries
	server, err := NewServer(cfg, queries)
	require.NoError(t, err)

	// Test that server can be stopped (even if not started)
	ctx := context.Background()
	err = server.Stop(ctx)
	assert.NoError(t, err)
}
