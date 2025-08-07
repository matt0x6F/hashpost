package services

import (
	"context"
	"html/template"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMailGun is a mock implementation of the MailGun client
type MockMailGun struct {
	sentMessages []MockMessage
}

type MockMessage struct {
	From    string
	To      string
	Subject string
	Text    string
	HTML    string
}

func (m *MockMailGun) NewMessage(from, subject, text, to string) interface{} {
	return &MockMessage{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    text,
	}
}

func (m *MockMailGun) Send(ctx context.Context, message interface{}) (string, string, error) {
	if msg, ok := message.(*MockMessage); ok {
		m.sentMessages = append(m.sentMessages, *msg)
	}
	return "test-message-id", "test-message", nil
}

// TestEmailService tests the email service functionality
func TestEmailService(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Email: config.EmailConfig{
			Provider:    "mailgun",
			FromAddress: "test@hashpost.com",
			FromName:    "HashPost Test",
			MailGun: config.MailGunConfig{
				Domain:        "test.hashpost.com",
				SendingAPIKey: "test-api-key",
				BaseURL:       "https://api.mailgun.net",
				Region:        "us",
			},
		},
	}

	t.Run("NewEmailService", func(t *testing.T) {
		// Test with valid configuration
		service, err := NewEmailService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, "mailgun", service.config.Provider)
		assert.Equal(t, "test@hashpost.com", service.config.FromAddress)
	})

	t.Run("NewEmailServiceInvalidProvider", func(t *testing.T) {
		invalidCfg := *cfg
		invalidCfg.Email.Provider = "invalid"

		service, err := NewEmailService(&invalidCfg)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "unsupported email provider")
	})

	t.Run("NewEmailServiceMissingCredentials", func(t *testing.T) {
		invalidCfg := *cfg
		invalidCfg.Email.MailGun.Domain = ""

		service, err := NewEmailService(&invalidCfg)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "mailgun domain and sending API key are required")
	})
}

// TestEmailTemplates tests template loading and rendering
func TestEmailTemplates(t *testing.T) {
	// Create temporary directory for templates
	tempDir, err := os.MkdirTemp("", "email_templates")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test configuration
	cfg := &config.Config{
		Email: config.EmailConfig{
			Provider:    "mailgun",
			FromAddress: "test@hashpost.com",
			FromName:    "HashPost Test",
			MailGun: config.MailGunConfig{
				Domain:        "test.hashpost.com",
				SendingAPIKey: "test-api-key",
				BaseURL:       "https://api.mailgun.net",
				Region:        "us",
			},
		},
	}

	service := &EmailService{
		config:      &cfg.Email,
		templateDir: tempDir,
		templates:   make(map[string]*template.Template),
	}

	t.Run("LoadTemplates", func(t *testing.T) {
		// Create test template files
		testTemplates := []struct {
			name     string
			htmlFile string
			textFile string
		}{
			{"welcome", "welcome.html", "welcome.txt"},
			{"password_reset", "password_reset.html", "password_reset.txt"},
		}

		for _, tmpl := range testTemplates {
			// Create HTML template
			htmlContent := `<!DOCTYPE html>
<html>
<head><title>{{.Subject}}</title></head>
<body>
<h1>Hello {{.ToName}}</h1>
<p>{{.Data.message}}</p>
</body>
</html>`
			err := os.WriteFile(filepath.Join(tempDir, tmpl.htmlFile), []byte(htmlContent), 0644)
			require.NoError(t, err)

			// Create text template
			textContent := `Hello {{.ToName}},

{{.Data.message}}

Best regards,
HashPost`
			err = os.WriteFile(filepath.Join(tempDir, tmpl.textFile), []byte(textContent), 0644)
			require.NoError(t, err)
		}

		err := service.loadTemplates()
		require.NoError(t, err)

		// Check that templates were loaded
		templates := service.GetTemplateList()
		assert.Contains(t, templates, "welcome")
		assert.Contains(t, templates, "password_reset")
	})

	t.Run("GetTemplateSubject", func(t *testing.T) {
		subjects := map[string]string{
			"welcome":            "Welcome to HashPost!",
			"password_reset":     "Reset Your HashPost Password",
			"email_verification": "Verify Your Email Address",
			"notification":       "New Notification from HashPost",
			"moderation_alert":   "Moderation Alert",
			"unknown":            "Message from HashPost",
		}

		for template, expected := range subjects {
			subject := service.getTemplateSubject(template)
			assert.Equal(t, expected, subject)
		}
	})
}

// TestEmailValidation tests email validation functionality
func TestEmailValidation(t *testing.T) {
	service := &EmailService{}

	t.Run("ValidateEmailValid", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
		}

		for _, email := range validEmails {
			assert.True(t, service.ValidateEmail(email), "Email should be valid: %s", email)
		}
	})

	t.Run("ValidateEmailInvalid", func(t *testing.T) {
		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"user@",
			"",
			"user@.com",
		}

		for _, email := range invalidEmails {
			assert.False(t, service.ValidateEmail(email), "Email should be invalid: %s", email)
		}
	})
}

// TestEmailData tests the EmailData structure
func TestEmailData(t *testing.T) {
	data := EmailData{
		ToName:         "John Doe",
		ToEmail:        "john@example.com",
		Subject:        "Test Subject",
		FromName:       "HashPost",
		FromEmail:      "noreply@hashpost.com",
		SiteURL:        "https://hashpost.com",
		UnsubscribeURL: "https://hashpost.com/unsubscribe",
		Data: map[string]interface{}{
			"message": "Test message",
			"url":     "https://example.com",
		},
	}

	assert.Equal(t, "John Doe", data.ToName)
	assert.Equal(t, "john@example.com", data.ToEmail)
	assert.Equal(t, "Test Subject", data.Subject)
	assert.Equal(t, "HashPost", data.FromName)
	assert.Equal(t, "noreply@hashpost.com", data.FromEmail)
	assert.Equal(t, "https://hashpost.com", data.SiteURL)
	assert.Equal(t, "https://hashpost.com/unsubscribe", data.UnsubscribeURL)
	assert.Equal(t, "Test message", data.Data["message"])
	assert.Equal(t, "https://example.com", data.Data["url"])
}

// TestEmailTemplateStructure tests the EmailTemplate structure
func TestEmailTemplateStructure(t *testing.T) {
	template := EmailTemplate{
		Name:        "test_template",
		Subject:     "Test Subject",
		HTMLFile:    "test.html",
		TextFile:    "test.txt",
		Description: "Test template description",
	}

	assert.Equal(t, "test_template", template.Name)
	assert.Equal(t, "Test Subject", template.Subject)
	assert.Equal(t, "test.html", template.HTMLFile)
	assert.Equal(t, "test.txt", template.TextFile)
	assert.Equal(t, "Test template description", template.Description)
}

// TestEmailStats tests email statistics functionality
func TestEmailStats(t *testing.T) {
	service := &EmailService{
		config: &config.EmailConfig{
			MailGun: config.MailGunConfig{
				Domain: "test.hashpost.com",
			},
		},
	}

	ctx := context.Background()
	stats, err := service.GetEmailStats(ctx, 24*time.Hour)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "mailgun", stats["provider"])
	assert.Equal(t, "test.hashpost.com", stats["domain"])
	assert.Equal(t, "operational", stats["status"])
}
