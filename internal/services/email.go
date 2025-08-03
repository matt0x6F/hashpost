package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/rs/zerolog/log"
)

// EmailService handles email sending with template support
type EmailService struct {
	config       *config.EmailConfig
	serverConfig *config.ServerConfig
	mailgun      mailgun.Mailgun
	templates    map[string]*template.Template
	templateMu   sync.RWMutex
	templateDir  string
}

// EmailData represents the data passed to email templates
type EmailData struct {
	ToName         string
	ToEmail        string
	Subject        string
	FromName       string
	FromEmail      string
	SiteURL        string
	UnsubscribeURL string
	Data           map[string]interface{}
}

// EmailTemplate represents a template configuration
type EmailTemplate struct {
	Name        string
	Subject     string
	HTMLFile    string
	TextFile    string
	Description string
}

// NewEmailService creates a new email service with MailGun integration
func NewEmailService(cfg *config.Config) (*EmailService, error) {
	if cfg.Email.Provider != "mailgun" {
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Email.Provider)
	}

	if cfg.Email.MailGun.Domain == "" || cfg.Email.MailGun.SendingAPIKey == "" {
		return nil, fmt.Errorf("mailgun domain and sending API key are required")
	}

	// Create MailGun client
	var mg mailgun.Mailgun
	if cfg.Email.MailGun.Region == "eu" {
		mg = mailgun.NewMailgun(cfg.Email.MailGun.Domain, cfg.Email.MailGun.SendingAPIKey)
		mg.SetAPIBase("https://api.eu.mailgun.net")
	} else {
		mg = mailgun.NewMailgun(cfg.Email.MailGun.Domain, cfg.Email.MailGun.SendingAPIKey)
	}

	service := &EmailService{
		config:       &cfg.Email,
		serverConfig: &cfg.Server,
		mailgun:      mg,
		templates:    make(map[string]*template.Template),
		templateDir:  "./templates/email",
	}

	// Load templates
	if err := service.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	log.Info().Str("provider", "mailgun").Str("domain", cfg.Email.MailGun.Domain).Msg("Email service initialized")
	return service, nil
}

// loadTemplates loads all email templates from disk
func (s *EmailService) loadTemplates() error {
	s.templateMu.Lock()
	defer s.templateMu.Unlock()

	// Ensure template directory exists
	if err := os.MkdirAll(s.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Define available templates
	templates := []EmailTemplate{
		{
			Name:        "welcome",
			Subject:     "Welcome to HashPost!",
			HTMLFile:    "welcome.html",
			TextFile:    "welcome.txt",
			Description: "Welcome email for new users",
		},
		{
			Name:        "password_reset",
			Subject:     "Reset Your HashPost Password",
			HTMLFile:    "password_reset.html",
			TextFile:    "password_reset.txt",
			Description: "Password reset email",
		},
		{
			Name:        "email_verification",
			Subject:     "Verify Your Email Address",
			HTMLFile:    "email_verification.html",
			TextFile:    "email_verification.txt",
			Description: "Email verification email",
		},
		{
			Name:        "notification",
			Subject:     "New Notification from HashPost",
			HTMLFile:    "notification.html",
			TextFile:    "notification.txt",
			Description: "General notification email",
		},
		{
			Name:        "moderation_alert",
			Subject:     "Moderation Alert",
			HTMLFile:    "moderation_alert.html",
			TextFile:    "moderation_alert.txt",
			Description: "Moderation alert email",
		},
	}

	// Load each template
	for _, tmpl := range templates {
		if err := s.loadTemplate(tmpl); err != nil {
			log.Warn().Str("template", tmpl.Name).Err(err).Msg("Failed to load template")
			continue
		}
	}

	log.Info().Int("template_count", len(s.templates)).Msg("Email templates loaded")
	return nil
}

// loadTemplate loads a single template from disk
func (s *EmailService) loadTemplate(tmpl EmailTemplate) error {
	// Load HTML template
	htmlPath := filepath.Join(s.templateDir, tmpl.HTMLFile)
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to load HTML template %s: %w", tmpl.Name, err)
	}

	// Load text template
	textPath := filepath.Join(s.templateDir, tmpl.TextFile)
	textContent, err := os.ReadFile(textPath)
	if err != nil {
		return fmt.Errorf("failed to load text template %s: %w", tmpl.Name, err)
	}

	// Parse templates
	htmlTmpl, err := template.New(tmpl.Name + "_html").Parse(string(htmlContent))
	if err != nil {
		return fmt.Errorf("failed to parse HTML template %s: %w", tmpl.Name, err)
	}

	textTmpl, err := template.New(tmpl.Name + "_text").Parse(string(textContent))
	if err != nil {
		return fmt.Errorf("failed to parse text template %s: %w", tmpl.Name, err)
	}

	// Store templates
	s.templates[tmpl.Name+"_html"] = htmlTmpl
	s.templates[tmpl.Name+"_text"] = textTmpl

	return nil
}

// SendEmail sends an email using a template
func (s *EmailService) SendEmail(ctx context.Context, templateName, toEmail, toName string, data map[string]interface{}) error {
	s.templateMu.RLock()
	htmlTmpl, htmlExists := s.templates[templateName+"_html"]
	textTmpl, textExists := s.templates[templateName+"_text"]
	s.templateMu.RUnlock()

	if !htmlExists || !textExists {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Prepare email data
	emailData := EmailData{
		ToName:         toName,
		ToEmail:        toEmail,
		Subject:        s.getTemplateSubject(templateName),
		FromName:       s.config.FromName,
		FromEmail:      s.config.FromAddress,
		SiteURL:        s.serverConfig.SiteURL,
		UnsubscribeURL: fmt.Sprintf("%s/unsubscribe?email=%s", s.serverConfig.SiteURL, toEmail),
		Data:           data,
	}

	// Add template name to data for conditional rendering
	emailData.Data["TemplateName"] = templateName

	// Render templates
	var htmlBody, textBody bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBody, emailData); err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}
	if err := textTmpl.Execute(&textBody, emailData); err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Create message
	message := s.mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", emailData.FromName, emailData.FromEmail),
		emailData.Subject,
		textBody.String(),
		toEmail,
	)

	// Set HTML content
	message.SetHtml(htmlBody.String())

	// Add custom headers if needed (MailGun Go SDK doesn't support SetHeader)
	// message.SetHeader("X-Mailgun-Variables", fmt.Sprintf(`{"template":"%s"}`, templateName))

	// Send message
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, _, err := s.mailgun.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().
		Str("template", templateName).
		Str("to", toEmail).
		Msg("Email sent successfully")

	return nil
}

// SendRawEmail sends a raw email without using templates
func (s *EmailService) SendRawEmail(ctx context.Context, toEmail, toName, subject, htmlBody, textBody string) error {
	message := s.mailgun.NewMessage(
		fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress),
		subject,
		textBody,
		toEmail,
	)

	message.SetHtml(htmlBody)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, _, err := s.mailgun.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send raw email: %w", err)
	}

	log.Info().
		Str("to", toEmail).
		Str("subject", subject).
		Msg("Raw email sent successfully")

	return nil
}

// ReloadTemplates reloads all templates from disk
func (s *EmailService) ReloadTemplates() error {
	return s.loadTemplates()
}

// GetTemplateList returns a list of available templates
func (s *EmailService) GetTemplateList() []string {
	s.templateMu.RLock()
	defer s.templateMu.RUnlock()

	var templates []string
	for name := range s.templates {
		if strings.HasSuffix(name, "_html") {
			templates = append(templates, strings.TrimSuffix(name, "_html"))
		}
	}
	return templates
}

// getTemplateSubject returns the subject for a template
func (s *EmailService) getTemplateSubject(templateName string) string {
	subjects := map[string]string{
		"welcome":            "Welcome to HashPost!",
		"password_reset":     "Reset Your HashPost Password",
		"email_verification": "Verify Your Email Address",
		"notification":       "New Notification from HashPost",
		"moderation_alert":   "Moderation Alert",
	}

	if subject, exists := subjects[templateName]; exists {
		return subject
	}
	return "Message from HashPost"
}

// ValidateEmail validates an email address format
func (s *EmailService) ValidateEmail(email string) bool {
	// Basic email validation
	if email == "" {
		return false
	}

	// Check for @ symbol
	if !strings.Contains(email, "@") {
		return false
	}

	// Split by @ and check parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	// Check local part (before @)
	localPart := parts[0]
	if localPart == "" {
		return false
	}

	// Check domain part (after @)
	domainPart := parts[1]
	if domainPart == "" {
		return false
	}

	// Check for domain separator
	if !strings.Contains(domainPart, ".") {
		return false
	}

	// Check that domain has valid structure
	domainParts := strings.Split(domainPart, ".")
	if len(domainParts) < 2 {
		return false
	}

	// Check that TLD is not empty and domain parts are not empty
	for _, part := range domainParts {
		if part == "" {
			return false
		}
	}

	return true
}

// GetEmailStats returns email sending statistics (if available from MailGun)
func (s *EmailService) GetEmailStats(ctx context.Context, duration time.Duration) (map[string]interface{}, error) {
	// This would require additional MailGun API calls to get statistics
	// For now, return a basic structure
	return map[string]interface{}{
		"provider": "mailgun",
		"domain":   s.config.MailGun.Domain,
		"status":   "operational",
	}, nil
}
