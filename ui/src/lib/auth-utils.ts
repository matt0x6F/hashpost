import { getApi, setAccessToken, setRefreshToken, getRefreshToken } from './api-client';
import { AuthenticationApi } from '@/generated/api/src/apis/AuthenticationApi';
import type { UserSession, UserLoginResponse, UserRegistrationResponse } from '@/generated/api/src/models';

export class AuthRefreshFailedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthRefreshFailedError';
  }
}

// Type guards for response types
function isLoginResponse(response: unknown): response is UserLoginResponse {
  return Boolean(response && typeof response === 'object' && response !== null && 'accessToken' in response);
}

function isRegistrationResponse(response: unknown): response is UserRegistrationResponse {
  return Boolean(response && typeof response === 'object' && response !== null && 'accessToken' in response);
}

// Authenticate user by calling /auth/me with Bearer token
export async function authenticateUser(): Promise<UserSession | null> {
  const authApi = getApi(AuthenticationApi);
  try {
    console.log('[auth-utils] Attempting to authenticate user...');
    const response = await authApi.getCurrentUserSession();
    console.log('[auth-utils] Authentication successful:', response);
    return response;
  } catch (error) {
    console.log('[auth-utils] Authentication failed:', error);
    // Check if it's a 401 Unauthorized response
    if (error && typeof error === 'object' && 'response' in error) {
      const response = (error as { response?: Response }).response;
      if (response && response.status === 401) {
        console.log('[auth-utils] 401 Unauthorized - no valid token');
        // No refresh token in atproto - just return null for unauthenticated
        return null;
      } else if (response) {
        console.warn('[auth-utils] Server returned unexpected status:', response.status);
      }
    }
    return null;
  }
}

// Login user with email/password
export async function loginUser(email: string, password: string): Promise<UserLoginResponse> {
  console.log('[auth-utils] loginUser called with email:', email);
  const authApi = getApi(AuthenticationApi);
  console.log('[auth-utils] About to call authApi.loginUser');
  const response = await authApi.loginUser({
    email,
    password
  });
  console.log('[auth-utils] Login response received:', {
    hasAccessToken: !!response.accessToken,
    accessTokenLength: response.accessToken?.length
  });
  
  // Set the access token for future API calls
  console.log('[auth-utils] About to call setAccessToken');
  setAccessToken(response.accessToken);
  console.log('[auth-utils] setAccessToken completed');
  
  // Set the refresh token for future token refreshes
  console.log('[auth-utils] About to call setRefreshToken');
  setRefreshToken(response.refreshToken);
  console.log('[auth-utils] setRefreshToken completed');
  
  return response;
}

// Register user with email/password/handle
export async function registerUser(email: string, password: string, handle: string, inviteCode?: string): Promise<UserRegistrationResponse> {
  const authApi = getApi(AuthenticationApi);
  const response = await authApi.registerUser({
    email,
    password,
    handle,
    inviteCode
  });
  
  // Set the access token for future API calls
  setAccessToken(response.accessToken);
  
  // Set the refresh token for future token refreshes
  setRefreshToken(response.refreshToken);
  
  return response;
}

// Refresh access token using refresh token
export async function refreshAccessToken(): Promise<string | null> {
  const currentRefreshToken = getRefreshToken();
  
  if (!currentRefreshToken) {
    console.log('[auth-utils] No refresh token available');
    throw new AuthRefreshFailedError('No refresh token available');
  }

  console.log('[auth-utils] Attempting to refresh access token');
  
  try {
    // Call the atproto refresh session endpoint (on PDS server, not AppView)
    const response = await fetch(`${process.env.NEXT_PUBLIC_PDS_URL || 'http://localhost:8080'}/xrpc/com.atproto.server.refreshSession`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refreshJwt: currentRefreshToken
      })
    });

    if (!response.ok) {
      console.log('[auth-utils] Refresh token request failed:', response.status, response.statusText);
      throw new AuthRefreshFailedError(`Token refresh failed: ${response.status} ${response.statusText}`);
    }

    const refreshResponse = await response.json();
    console.log('[auth-utils] Token refresh successful');

    // Update both tokens in storage
    setAccessToken(refreshResponse.accessJwt);
    setRefreshToken(refreshResponse.refreshJwt);

    return refreshResponse.accessJwt;
  } catch (error) {
    console.error('[auth-utils] Token refresh failed:', error);
    if (error instanceof AuthRefreshFailedError) {
      throw error;
    }
    throw new AuthRefreshFailedError(`Token refresh failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

// Logout user by calling backend logout endpoint
export async function logoutUser(): Promise<void> {
  try {
    const authApi = getApi(AuthenticationApi);
    await authApi.logoutUser();
  } catch (error) {
    console.error('Error during logout:', error);
  } finally {
    // Clear both tokens (this will also clear from localStorage)
    setAccessToken(null);
    setRefreshToken(null);
  }
}

// Store user data in localStorage (excluding sensitive tokens)
export function storeUserInLocalStorage(userData: UserLoginResponse | UserRegistrationResponse): void {
  if (typeof window === 'undefined') return;
  try {
    const userDataToStore = {
      ...userData,
      // Don't store tokens in localStorage - they're in memory
      accessToken: undefined,
      refreshToken: undefined,
    };
    localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
  } catch (error) {
    console.error('Error storing user data in localStorage:', error);
  }
}

// Clear user data from localStorage
export function clearUserFromLocalStorage(): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem('hashpost_user');
  } catch (error) {
    console.error('Error clearing user data from localStorage:', error);
  }
}

// Get user data from localStorage (fallback)
export function getUserFromLocalStorage(): (UserLoginResponse | UserRegistrationResponse) | null {
  if (typeof window === 'undefined') return null;
  try {
    const stored = localStorage.getItem('hashpost_user');
    if (stored) {
      const userData = JSON.parse(stored);
      if (isLoginResponse(userData) || isRegistrationResponse(userData)) {
        return userData;
      }
    }
  } catch (error) {
    console.error('Error reading from localStorage:', error);
  }
  return null;
}