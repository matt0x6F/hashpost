# Email Validation

HashPost includes comprehensive email validation capabilities using both basic RFC 5322 validation and advanced validation through the [truemail-go](https://github.com/truemail-rb/truemail-go) library.

## Features

### Basic Validation
- **RFC 5322 Compliance**: Full RFC 5322 email format validation
- **Length Validation**: Proper length checks for total email (254 chars), local part (64 chars), domain part (253 chars)
- **Domain Structure**: Validates domain parts, TLD requirements, and character restrictions
- **Special Characters**: Supports RFC 5322 compliant special characters in local part

### Advanced Validation (truemail-go)
- **Format Validation**: RFC 5322 compliant email format validation
- **DNS Validation**: MX record validation to verify mail servers exist
- **SMTP Validation**: Optional SMTP validation to check if email actually exists
- **Configurable Levels**: Choose between basic, MX, or SMTP validation levels
- **Fail-Fast Options**: Configurable timeouts and fail-fast behavior for better UX
- **Domain Blacklisting**: Block disposable email providers and suspicious domains
- **MX IP Blacklisting**: Block specific MX server IP addresses

## Configuration

Email validation is configured through environment variables:

```bash
# Enable email validation
EMAIL_VALIDATION_ENABLED=true

# Verifier email for SMTP validation
EMAIL_VALIDATION_VERIFIER_EMAIL=verifier@yourdomain.com

# SMTP configuration for validation
EMAIL_VALIDATION_SMTP_HOST=smtp.gmail.com
EMAIL_VALIDATION_SMTP_PORT=587
EMAIL_VALIDATION_SMTP_USER=verifier@yourdomain.com

# Timeouts
EMAIL_VALIDATION_CONNECTION_TIMEOUT=5
EMAIL_VALIDATION_RESPONSE_TIMEOUT=2

# Behavior options
EMAIL_VALIDATION_SMTP_FAIL_FAST=true
EMAIL_VALIDATION_SMTP_SAFE_CHECK=true

# Validation level: "basic", "mx", "smtp"
EMAIL_VALIDATION_LEVEL=mx

# Blacklist configuration (comma-separated lists)
EMAIL_VALIDATION_BLACKLISTED_DOMAINS=tempmail.com,10minutemail.com,guerrillamail.com
EMAIL_VALIDATION_BLACKLISTED_MX_IPS=192.168.1.1,10.0.0.1
```

## Usage

### Basic Validation
```go
import "github.com/matt0x6f/hashpost/internal/api/validation"

// Basic RFC 5322 validation
err := validation.ValidateEmailBasic("user@example.com")
if err != nil {
    // Handle validation error
}
```

### Advanced Validation
```go
import (
    "github.com/matt0x6f/hashpost/internal/api/validation"
    "github.com/matt0x6f/hashpost/internal/config"
)

cfg := config.Load()

// Comprehensive validation with format, DNS, and SMTP checks
err := validation.ValidateEmailStrict("user@example.com", cfg)
if err != nil {
    // Handle validation error
}
```

### Validation Levels

1. **Basic (`basic`)**: Only RFC 5322 format validation
2. **MX (`mx`)**: Format + DNS MX record validation
3. **SMTP (`smtp`)**: Format + DNS + SMTP validation

## Integration Points

### User Registration
Email validation is automatically applied during user registration:

```go
// In auth handler
if err := validation.ValidateEmailStrict(input.Email, cfg); err != nil {
    return huma.Error422UnprocessableEntity(err.Error())
}
```

### Admin Operations
For critical operations, use strict validation:

```go
// For admin user creation or sensitive operations
if err := validation.ValidateEmailStrict(email, cfg); err != nil {
    return fmt.Errorf("email validation failed: %w", err)
}
```

## Blacklisting

### Domain Blacklisting
Block disposable email providers and suspicious domains:

```bash
EMAIL_VALIDATION_BLACKLISTED_DOMAINS=tempmail.com,10minutemail.com,guerrillamail.com,mailinator.com
```

### MX IP Blacklisting
Block specific MX server IP addresses:

```bash
EMAIL_VALIDATION_BLACKLISTED_MX_IPS=192.168.1.1,10.0.0.1,172.16.0.1
```

## Error Handling

The validation functions return descriptive error messages:

- `"email is required"` - Empty email
- `"email is too long (maximum 254 characters)"` - Exceeds RFC limits
- `"email must contain exactly one @ symbol"` - Invalid format
- `"domain must contain at least one dot"` - Invalid domain structure
- `"email validation failed: domain does not have valid mail servers"` - DNS validation failed

## Security Considerations

1. **SMTP Credentials**: Store verifier email credentials securely
2. **Rate Limiting**: Implement rate limiting for validation requests
3. **Fail-Safe**: Use `SmtpSafeCheck` to avoid false negatives
4. **Timeouts**: Configure appropriate timeouts to prevent hanging

## Performance

- **Basic Validation**: ~1ms (regex only)
- **MX Validation**: ~100-500ms (DNS lookup)
- **SMTP Validation**: ~1-5s (SMTP connection)

## Troubleshooting

### Common Issues

1. **SMTP Connection Failed**
   - Check SMTP host and port configuration
   - Verify credentials are correct
   - Ensure firewall allows SMTP connections

2. **DNS Validation Failed**
   - Domain may not have MX records
   - Network connectivity issues
   - DNS server configuration

3. **Timeout Issues**
   - Increase `EMAIL_VALIDATION_CONNECTION_TIMEOUT`
   - Enable `EMAIL_VALIDATION_SMTP_FAIL_FAST`

### Debug Mode

Enable debug logging to troubleshoot validation issues:

```bash
LOG_LEVEL=debug
```

## Migration from Basic to Advanced

To migrate from basic to advanced validation:

1. **Configure Environment Variables**
   ```bash
   EMAIL_VALIDATION_ENABLED=true
   EMAIL_VALIDATION_VERIFIER_EMAIL=your-verifier@domain.com
   EMAIL_VALIDATION_LEVEL=mx
   ```

2. **Update Code**
   ```go
   // Before
   err := validation.ValidateEmailBasic(email)
   
   // After
   err := validation.ValidateEmailStrict(email, cfg)
   ```

3. **Test Thoroughly**
   - Test with various email formats
   - Verify DNS validation works
   - Test SMTP validation if enabled

## Best Practices

1. **Start with Basic**: Use basic validation for registration
2. **Gradual Rollout**: Enable advanced validation gradually
3. **Monitor Performance**: Watch for validation timeouts
4. **User Experience**: Provide clear error messages
5. **Fallback Strategy**: Have a fallback for validation failures 