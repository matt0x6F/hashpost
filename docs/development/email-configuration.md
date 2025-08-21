# Email Configuration

HashPost uses MailGun for email delivery in both testing and production environments.

## MailGun Setup

### 1. Create MailGun Account
1. Sign up at [mailgun.com](https://mailgun.com)
2. Verify your account and add payment method if needed
3. Create a domain for your environment:
   - **Testing**: `mg.testing.hashpost.dev`
   - **Production**: `mg.hashpost.dev`

### 2. Domain Configuration
1. Add your domain in MailGun dashboard
2. Configure DNS records as instructed by MailGun:
   - TXT record for domain verification
   - MX records for receiving
   - CNAME for tracking (optional)
   - TXT records for SPF and DKIM

### 3. Get API Credentials
1. Go to MailGun dashboard → Settings → API Keys
2. Copy your **Private API key** (starts with `key-`)
3. Note your domain name (e.g., `mg.testing.hashpost.dev`)

## Kubernetes Configuration

### Testing Environment

1. **Create the secret:**
   ```bash
   kubectl create secret generic hashpost-email-secrets-testing \
     --from-literal=mailgun-domain="mg.testing.hashpost.dev" \
     --from-literal=mailgun-api-key="key-your-mailgun-api-key" \
     --namespace=hashpost-testing
   ```

2. **Or use the example file:**
   ```bash
   cp deploy/overlays/testing/email-secrets.yaml.example deploy/overlays/testing/email-secrets.yaml
   # Edit the file with your base64-encoded values
   kubectl apply -f deploy/overlays/testing/email-secrets.yaml
   ```

### Production Environment

1. **Create the secret:**
   ```bash
   kubectl create secret generic hashpost-email-secrets-production \
     --from-literal=mailgun-domain="mg.hashpost.dev" \
     --from-literal=mailgun-api-key="key-your-mailgun-api-key" \
     --namespace=hashpost-production
   ```

2. **Or use the example file:**
   ```bash
   cp deploy/overlays/production/email-secrets.yaml.example deploy/overlays/production/email-secrets.yaml
   # Edit the file with your base64-encoded values
   kubectl apply -f deploy/overlays/production/email-secrets.yaml
   ```

## Environment Variables

The following environment variables are automatically configured:

### Testing Environment
- `EMAIL_PROVIDER=mailgun`
- `EMAIL_FROM_ADDRESS=noreply@testing.hashpost.dev`
- `EMAIL_FROM_NAME=HashPost Testing`
- `MAILGUN_REGION=us`
- `MAILGUN_BASE_URL=https://api.mailgun.net`
- `SERVER_SITE_URL=https://testing.hashpost.dev`

### Production Environment
- `EMAIL_PROVIDER=mailgun`
- `EMAIL_FROM_ADDRESS=noreply@hashpost.dev`
- `EMAIL_FROM_NAME=HashPost`
- `MAILGUN_REGION=us`
- `MAILGUN_BASE_URL=https://api.mailgun.net`
- `SERVER_SITE_URL=https://hashpost.dev`

## Email Templates

HashPost includes the following email templates:

- **welcome** - Welcome new users
- **email_verification** - Email address verification
- **password_reset** - Password reset instructions
- **notification** - General notifications
- **moderation_alert** - Moderation alerts

Templates are located in:
- `internal/services/templates/email/` (Go service)
- `templates/email/` (Docker container)

## Testing Email Delivery

### 1. Local Testing
For local development, you can:
1. Use MailGun's sandbox domain (limited to authorized recipients)
2. Set up a test domain like `mg.dev.hashpost.dev`
3. Use email testing services like MailHog or Mailtrap

### 2. Production Testing
1. Send test emails using the admin CLI:
   ```bash
   ./main send-test-email --to="your-email@example.com" --template="welcome"
   ```

2. Monitor MailGun logs in the dashboard for delivery status

## Troubleshooting

### Common Issues

1. **"mailgun domain and sending API key are required"**
   - Ensure the Kubernetes secret exists and contains the correct keys
   - Verify the secret is in the correct namespace

2. **"failed to send email"**
   - Check MailGun dashboard for delivery logs
   - Verify DNS records are properly configured
   - Ensure API key has sending permissions

3. **DNS Configuration Issues**
   - Use MailGun's DNS checker tool
   - Wait for DNS propagation (up to 48 hours)
   - Verify SPF and DKIM records

### Debug Commands

```bash
# Check if secret exists
kubectl get secret hashpost-email-secrets-testing -n hashpost-testing

# View secret contents (base64 encoded)
kubectl get secret hashpost-email-secrets-testing -n hashpost-testing -o yaml

# Check pod environment variables
kubectl exec -it deployment/hashpost-backend-testing -n hashpost-testing -- env | grep -i mail

# View application logs
kubectl logs deployment/hashpost-backend-testing -n hashpost-testing | grep -i email
```

## Security Considerations

1. **Never commit actual API keys to git**
2. **Use separate domains for testing and production**
3. **Rotate API keys periodically**
4. **Monitor MailGun usage and billing**
5. **Set up proper DNS security (SPF, DKIM, DMARC)**

## Cost Optimization

1. **MailGun Pricing**: First 5,000 emails/month are free
2. **Use sandbox domain** for development testing
3. **Monitor usage** in MailGun dashboard
4. **Set up billing alerts** to avoid unexpected charges
