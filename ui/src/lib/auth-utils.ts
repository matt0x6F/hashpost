import { getApi } from './api-client';
import { AuthenticationApi } from '@/generated/api/src/apis/AuthenticationApi';
import type { UserLoginResponseBody } from '@/generated/api/src/models';

export class AuthRefreshFailedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthRefreshFailedError';
  }
}

// Type guards for response types
function isLoginResponse(response: unknown): response is UserLoginResponseBody {
  return Boolean(response && typeof response === 'object' && response !== null && 'pseudonyms' in response);
}

// Authenticate user by calling /auth/me and letting the browser send cookies
export async function authenticateUser(): Promise<UserLoginResponseBody | null> {
  const authApi = getApi(AuthenticationApi);
  try {
    const response = await authApi.getCurrentUserSession();
    // Convert response to UserLoginResponseBody shape (tokens not available in JS)
    return {
      userId: response.userId,
      email: response.email,
      createdAt: response.createdAt,
      lastActiveAt: response.lastActiveAt,
      isActive: response.isActive,
      isSuspended: response.isSuspended,
      roles: response.roles || [],
      capabilities: response.capabilities || [],
      activePseudonymId: response.activePseudonymId,
      displayName: response.displayName,
      pseudonyms: response.pseudonyms || [],
      accessToken: '',
      refreshToken: ''
    };
  } catch (error) {
    // Check if it's a 401 Unauthorized response
    if (error && typeof error === 'object' && 'response' in error) {
      const response = (error as { response?: Response }).response;
      if (response && response.status === 401) {
        // Try to refresh the token
        try {
          await authApi.refreshToken({ refreshToken: '' });
          
          // After successful refresh, try /auth/me again
          const retryResponse = await authApi.getCurrentUserSession();
          
          return {
            userId: retryResponse.userId,
            email: retryResponse.email,
            createdAt: retryResponse.createdAt,
            lastActiveAt: retryResponse.lastActiveAt,
            isActive: retryResponse.isActive,
            isSuspended: retryResponse.isSuspended,
            roles: retryResponse.roles || [],
            capabilities: retryResponse.capabilities || [],
            activePseudonymId: retryResponse.activePseudonymId,
            displayName: retryResponse.displayName,
            pseudonyms: retryResponse.pseudonyms || [],
            accessToken: '',
            refreshToken: ''
          };
        } catch {
          // If refresh fails, throw AuthRefreshFailedError
          throw new AuthRefreshFailedError('Token refresh failed');
        }
      } else if (response) {
        console.warn('[auth-utils] Server returned unexpected status:', response.status);
      }
    }
    return null;
  }
}

// Authenticate user with subforum-specific capabilities
export async function authenticateUserForSubforum(subforumName: string): Promise<UserLoginResponseBody | null> {
  const authApi = getApi(AuthenticationApi);
  try {
    const response = await authApi.getCurrentUserSessionForSubforum(subforumName);
    // Convert response to UserLoginResponseBody shape (tokens not available in JS)
    return {
      userId: response.userId,
      email: response.email,
      createdAt: response.createdAt,
      lastActiveAt: response.lastActiveAt,
      isActive: response.isActive,
      isSuspended: response.isSuspended,
      roles: response.roles || [],
      capabilities: response.capabilities || [], // This now includes subforum-specific capabilities
      activePseudonymId: response.activePseudonymId,
      displayName: response.displayName,
      pseudonyms: response.pseudonyms || [],
      accessToken: '',
      refreshToken: ''
    };
  } catch (error) {
    // Check if it's a 401 Unauthorized response
    if (error && typeof error === 'object' && 'response' in error) {
      const response = (error as { response?: Response }).response;
      if (response && response.status === 401) {
        // Try to refresh the token
        try {
          await authApi.refreshToken({ refreshToken: '' });
          
          // After successful refresh, try /auth/me/subforum again
          const retryResponse = await authApi.getCurrentUserSessionForSubforum(subforumName);
          
          return {
            userId: retryResponse.userId,
            email: retryResponse.email,
            createdAt: retryResponse.createdAt,
            lastActiveAt: retryResponse.lastActiveAt,
            isActive: retryResponse.isActive,
            isSuspended: retryResponse.isSuspended,
            roles: retryResponse.roles || [],
            capabilities: retryResponse.capabilities || [],
            activePseudonymId: retryResponse.activePseudonymId,
            displayName: retryResponse.displayName,
            pseudonyms: retryResponse.pseudonyms || [],
            accessToken: '',
            refreshToken: ''
          };
        } catch {
          // If refresh fails, throw AuthRefreshFailedError
          throw new AuthRefreshFailedError('Token refresh failed');
        }
      } else if (response) {
        console.warn('[auth-utils] Server returned unexpected status:', response.status);
      }
    }
    return null;
  }
}

// Logout user by calling backend logout endpoint
export async function logoutUser(): Promise<void> {
  try {
    const authApi = getApi(AuthenticationApi);
    await authApi.logoutUser({ refreshToken: '' });
  } catch (error) {
    console.error('Error during logout:', error);
  }
}

// Store user data in localStorage (excluding sensitive tokens)
export function storeUserInLocalStorage(userData: UserLoginResponseBody): void {
  if (typeof window === 'undefined') return;
  try {
    const userDataToStore = {
      ...userData,
      // Don't store tokens in localStorage - they're in cookies
      accessToken: undefined,
      refreshToken: undefined,
    };
    localStorage.setItem('user', JSON.stringify(userDataToStore));
  } catch (error) {
    console.error('Error storing user data in localStorage:', error);
  }
}

// Clear user data from localStorage
export function clearUserFromLocalStorage(): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem('user');
  } catch (error) {
    console.error('Error clearing user data from localStorage:', error);
  }
}

// Get user data from localStorage (fallback)
export function getUserFromLocalStorage(): UserLoginResponseBody | null {
  if (typeof window === 'undefined') return null;
  try {
    const stored = localStorage.getItem('user');
    if (stored) {
      const userData = JSON.parse(stored);
      if (isLoginResponse(userData)) {
        return userData;
      }
    }
  } catch (error) {
    console.error('Error reading from localStorage:', error);
  }
  return null;
} 