package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/matt0x6f/hashpost/internal/config"
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

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email format is invalid")
	}

	// Additional checks
	if len(email) > 254 {
		return fmt.Errorf("email is too long (maximum 254 characters)")
	}

	if strings.Count(email, "@") != 1 {
		return fmt.Errorf("email must contain exactly one @ symbol")
	}

	parts := strings.Split(email, "@")
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return fmt.Errorf("email has invalid format")
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
