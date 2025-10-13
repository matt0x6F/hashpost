package appview

import (
	"testing"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRBACService_Structure(t *testing.T) {
	// Test that RBACService can be created and has expected fields
	logger := testutil.CreateMockLogger()
	rbacService := &RBACService{
		logger: logger,
	}

	// Test that the struct has the expected fields
	require.NotNil(t, rbacService.logger)
	assert.Equal(t, logger, rbacService.logger)
}
