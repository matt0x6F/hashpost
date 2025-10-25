/**
 * JWT utilities for decoding and checking token expiration
 * No external dependencies - uses built-in base64 decoding
 */

export interface JWTPayload {
  exp: number;
  iat: number;
  sub: string;
  [key: string]: any;
}

/**
 * Decode a JWT token and return its payload
 * @param token JWT token string
 * @returns Decoded payload or null if invalid
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    // JWT format: header.payload.signature
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }

    // Decode the payload (middle part)
    const payload = parts[1];
    
    // Add padding if needed for base64 decoding
    const paddedPayload = payload + '='.repeat((4 - payload.length % 4) % 4);
    
    // Decode base64
    const decodedPayload = atob(paddedPayload);
    
    // Parse JSON
    return JSON.parse(decodedPayload);
  } catch (error) {
    console.error('Error decoding JWT:', error);
    return null;
  }
}

/**
 * Check if a JWT token is expired
 * @param token JWT token string
 * @returns true if expired, false if valid
 */
export function isTokenExpired(token: string): boolean {
  const payload = decodeJWT(token);
  if (!payload || !payload.exp) {
    console.log('[jwt-utils] Token is invalid or missing exp claim:', { hasPayload: !!payload, hasExp: !!payload?.exp });
    return true; // Consider invalid tokens as expired
  }

  const now = Math.floor(Date.now() / 1000);
  const isExpired = payload.exp <= now;
  
  console.log('[jwt-utils] Token expiration check:', {
    exp: payload.exp,
    now: now,
    isExpired: isExpired,
    timeUntilExpiry: payload.exp - now
  });
  
  return isExpired;
}

/**
 * Check if a JWT token is expiring soon (within specified minutes)
 * @param token JWT token string
 * @param minutesBeforeExpiry Minutes before expiry to consider "soon" (default: 5)
 * @returns true if expiring soon, false otherwise
 */
export function isTokenExpiringSoon(token: string, minutesBeforeExpiry: number = 5): boolean {
  const payload = decodeJWT(token);
  if (!payload || !payload.exp) {
    return true; // Consider invalid tokens as expiring soon
  }

  const now = Math.floor(Date.now() / 1000);
  const expiryTime = payload.exp;
  const bufferTime = minutesBeforeExpiry * 60; // Convert to seconds

  return expiryTime <= (now + bufferTime);
}

/**
 * Get the expiration time of a JWT token
 * @param token JWT token string
 * @returns Expiration timestamp in seconds, or null if invalid
 */
export function getTokenExpiration(token: string): number | null {
  const payload = decodeJWT(token);
  return payload?.exp || null;
}

/**
 * Get the time until a JWT token expires
 * @param token JWT token string
 * @returns Seconds until expiry, or null if invalid/expired
 */
export function getTimeUntilExpiry(token: string): number | null {
  const payload = decodeJWT(token);
  if (!payload || !payload.exp) {
    return null;
  }

  const now = Math.floor(Date.now() / 1000);
  const timeUntilExpiry = payload.exp - now;
  
  return timeUntilExpiry > 0 ? timeUntilExpiry : null;
}
