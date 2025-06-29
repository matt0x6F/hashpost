# CORS Configuration

The HashPost API supports configurable CORS (Cross-Origin Resource Sharing) settings to control which origins can access the API.

## Environment Variables

### CORS_ALLOWED_ORIGINS
Comma-separated list of allowed origins. Use `*` for development (allows all origins).

**Examples:**
```bash
# Development - allow all origins
CORS_ALLOWED_ORIGINS=*

# Production - specific origins only
CORS_ALLOWED_ORIGINS=https://app.hashpost.com,https://admin.hashpost.com

# Docker development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://ui:3000
```

### CORS_ALLOWED_METHODS
Comma-separated list of allowed HTTP methods.

**Default:** `GET,POST,PUT,DELETE,OPTIONS`

**Example:**
```bash
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS,PATCH
```

### CORS_ALLOWED_HEADERS
Comma-separated list of allowed request headers.

**Default:** `Authorization,Content-Type`

**Example:**
```bash
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With,Accept,Origin
```

### CORS_ALLOW_CREDENTIALS
Whether to allow credentials (cookies, authorization headers) in CORS requests.

**Default:** `true`

**Example:**
```bash
CORS_ALLOW_CREDENTIALS=true
```

### CORS_MAX_AGE
Maximum age (in seconds) for preflight requests to be cached.

**Default:** `300` (5 minutes)

**Example:**
```bash
CORS_MAX_AGE=86400  # 24 hours
```

## Docker Compose Configuration

The `docker-compose.yml` file includes CORS configuration for development:

```yaml
environment:
  CORS_ALLOWED_ORIGINS: http://localhost:3000,http://ui:3000
  CORS_ALLOWED_METHODS: GET,POST,PUT,DELETE,OPTIONS,PATCH
  CORS_ALLOWED_HEADERS: Content-Type,Authorization,X-Requested-With,Accept,Origin
  CORS_ALLOW_CREDENTIALS: true
  CORS_MAX_AGE: 86400
```

## Security Considerations

### Development
- Use `*` for `CORS_ALLOWED_ORIGINS` only in development
- This allows any origin to access the API

### Production
- Always specify exact origins in production
- Use HTTPS URLs only
- Consider using subdomain wildcards if needed (e.g., `https://*.hashpost.com`)

### Credentials
- When `CORS_ALLOW_CREDENTIALS` is `true`, the `Access-Control-Allow-Origin` header cannot be `*`
- The origin must be explicitly specified

## Testing CORS

You can test CORS configuration using curl:

```bash
# Test preflight request
curl -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:8888/health

# Test actual request
curl -X GET \
  -H "Origin: http://localhost:3000" \
  http://localhost:8888/health
```

## Browser Console Errors

Common CORS errors and solutions:

- **"No 'Access-Control-Allow-Origin' header"**: Check `CORS_ALLOWED_ORIGINS`
- **"Method not allowed"**: Check `CORS_ALLOWED_METHODS`
- **"Header not allowed"**: Check `CORS_ALLOWED_HEADERS`
- **"Credentials not supported"**: Check `CORS_ALLOW_CREDENTIALS` and ensure origin is not `*` 