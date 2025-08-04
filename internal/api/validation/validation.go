package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/truemail-rb/truemail-go"
)

// Common passwords to disallow (if enabled)
var commonPasswords = map[string]bool{
	"password":    true,
	"123456":      true,
	"12345678":    true,
	"qwerty":      true,
	"abc123":      true,
	"password123": true,
	"admin":       true,
	"letmein":     true,
	"welcome":     true,
	"monkey":      true,
	"dragon":      true,
	"master":      true,
	"football":    true,
	"superman":    true,
	"trustno1":    true,
	"butterfly":   true,
	"baseball":    true,
	"shadow":      true,
	"michael":     true,
	"jennifer":    true,
	"hunter":      true,
	"joshua":      true,
}

// ValidateEmail validates an email address with configurable validation levels
func ValidateEmail(email string, cfg *config.Config) error {
	// Skip validation if email validation is disabled
	if !cfg.Email.Validation.Enabled {
		return nil
	}

	// Configure truemail for comprehensive validation
	verifierEmail := cfg.Email.Validation.VerifierEmail

	// Use custom strict regex for validation
	customRegex := `^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	configuration, err := truemail.NewConfiguration(
		truemail.ConfigurationAttr{
			VerifierEmail:            verifierEmail,
			ConnectionTimeout:        cfg.Email.Validation.ConnectionTimeout,
			ResponseTimeout:          cfg.Email.Validation.ResponseTimeout,
			SmtpFailFast:             cfg.Email.Validation.SmtpFailFast,
			SmtpSafeCheck:            cfg.Email.Validation.SmtpSafeCheck,
			BlacklistedDomains:       cfg.Email.Validation.BlacklistedDomains,
			BlacklistedMxIpAddresses: cfg.Email.Validation.BlacklistedMxIPs,
			EmailPattern:             customRegex,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to configure email validator: %w", err)
	}

	// Determine validation level and method
	validationLevel := cfg.Email.Validation.ValidationLevel
	if validationLevel == "" {
		validationLevel = "basic"
	}

	// Perform validation based on level
	var result *truemail.ValidatorResult
	switch validationLevel {
	case "basic":
		result, err = truemail.Validate(email, configuration, "regex")
	case "mx":
		result, err = truemail.Validate(email, configuration, "mx")
	case "smtp":
		result, err = truemail.Validate(email, configuration, "smtp")
	default:
		result, err = truemail.Validate(email, configuration, "regex")
	}

	if err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			// Return the first meaningful error
			for _, validationErr := range result.Errors {
				if validationErr != "" {
					return fmt.Errorf("email validation failed: %s", validationErr)
				}
			}
		}
		return fmt.Errorf("email validation failed: domain does not have valid mail servers")
	}

	return nil
}

// ValidatePassword validates a password against the configured rules
func ValidatePassword(password string, config config.PasswordValidationConfig) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Check minimum length
	if len(password) < config.MinLength {
		return fmt.Errorf("password must be at least %d characters long", config.MinLength)
	}

	// Check for required character types
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Validate character requirements
	if config.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if config.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if config.RequireDigit && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if config.RequireSpecialChar && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check against common passwords
	if config.DisallowCommon && commonPasswords[strings.ToLower(password)] {
		return fmt.Errorf("password is too common, please choose a more unique password")
	}

	return nil
}

// ValidateDisplayName validates a display name
func ValidateDisplayName(displayName string) error {
	if displayName == "" {
		return fmt.Errorf("display_name is required")
	}

	if len(displayName) < 3 {
		return fmt.Errorf("display_name must be at least 3 characters long")
	}

	if len(displayName) > 50 {
		return fmt.Errorf("display_name must be no more than 50 characters long")
	}

	// Check for valid characters (alphanumeric, spaces, hyphens, underscores)
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	if !validNameRegex.MatchString(displayName) {
		return fmt.Errorf("display_name can only contain letters, numbers, spaces, hyphens, and underscores")
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(displayName) != displayName {
		return fmt.Errorf("display_name cannot start or end with whitespace")
	}

	return nil
}
