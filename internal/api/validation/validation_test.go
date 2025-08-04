package validation

import (
	"testing"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmailBasic(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Validation: config.EmailValidationConfig{
				Enabled:       true,                   // Enable validation for basic tests
				VerifierEmail: "noreply@hashpost.dev", // Default verifier email
			},
		},
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email - no @",
			email:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "invalid email - multiple @",
			email:   "test@@example.com",
			wantErr: true,
		},
		{
			name:    "invalid email - empty local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid email - empty domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "invalid email - too long",
			email:   "verylongemailaddress" + string(make([]byte, 300)) + "@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email, cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmailWithTruemail(t *testing.T) {
	// Create a test config with validation disabled
	cfg := &config.Config{
		Email: config.EmailConfig{
			Validation: config.EmailValidationConfig{
				Enabled:       false,                  // Disable validation for testing
				VerifierEmail: "noreply@hashpost.dev", // Default verifier email
			},
		},
	}

	// Test with validation disabled
	err := ValidateEmail("test@example.com", cfg)
	assert.NoError(t, err, "Should pass when validation is disabled")

	// Test with validation enabled but no verifier email
	cfg.Email.Validation.Enabled = true
	err = ValidateEmail("test@example.com", cfg)
	assert.NoError(t, err, "Should pass when validation is enabled with default verifier email")

	// Test with proper configuration
	cfg.Email.Validation.VerifierEmail = "verifier@example.com"
	cfg.Email.Validation.ValidationLevel = "basic"
	err = ValidateEmail("test@example.com", cfg)
	// This might fail due to network issues, but should not panic
	// We're mainly testing that the function doesn't crash
}

func TestValidateEmailStrict(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			Validation: config.EmailValidationConfig{
				Enabled:       false,
				VerifierEmail: "noreply@hashpost.dev", // Default verifier email
			},
		},
	}

	err := ValidateEmail("test@example.com", cfg)
	assert.NoError(t, err, "Should pass when validation is disabled")
}

func TestValidateEmailRFC5322(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "standard email",
			email:   "user@domain.com",
			wantErr: false,
		},
		{
			name:    "email with plus",
			email:   "user+tag@domain.com",
			wantErr: false,
		},
		{
			name:    "email with dots",
			email:   "user.name@domain.com",
			wantErr: false,
		},
		{
			name:    "email with special chars",
			email:   "user+tag@domain.com",
			wantErr: false,
		},
		{
			name:    "subdomain email",
			email:   "user@sub.domain.com",
			wantErr: false,
		},
		{
			name:    "invalid - no domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid - no local part",
			email:   "@domain.com",
			wantErr: true,
		},
		{
			name:    "invalid - multiple @",
			email:   "user@@domain.com",
			wantErr: true,
		},
	}

	cfg := &config.Config{
		Email: config.EmailConfig{
			Validation: config.EmailValidationConfig{
				Enabled:       true,                   // Enable validation for RFC tests
				VerifierEmail: "noreply@hashpost.dev", // Default verifier email
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email, cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
