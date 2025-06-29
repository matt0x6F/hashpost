import { getApi } from './api-client';
import { AuthenticationApi } from '@/generated/api/src/apis/AuthenticationApi';
import { UserLoginResponseBody, UserRegistrationResponseBody } from '@/generated/api/src/models';

// Type guard to check if response is a login response
function isLoginResponse(response: unknown): response is UserLoginResponseBody {
  return Boolean(response && typeof response === 'object' && 'userId' in response && 'email' in response);
}

// Type guard to check if response is a registration response
function isRegistrationResponse(response: unknown): response is UserRegistrationResponseBody {
  return Boolean(response && typeof response === 'object' && 'userId' in response && 'email' in response);
}

// Authenticate user by calling /auth/me and letting the browser send cookies
export async function authenticateUser(): Promise<UserLoginResponseBody | null> {
  console.log('[auth-utils] authenticateUser called');
  try {
    const authApi = getApi(AuthenticationApi);
    const response = await authApi.getCurrentUserSession();
    console.log('[auth-utils] /auth/me response:', response);
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
    console.log('[auth-utils] /auth/me failed or not authenticated:', error);
    return null;
  }
}

// Logout user by calling backend logout endpoint and clearing localStorage
export async function logoutUser(): Promise<void> {
  console.log('[auth-utils] Logging out user...');
  
  try {
    // Call backend logout endpoint to clear cookies
    const authApi = getApi(AuthenticationApi);
    await authApi.logoutUser({ refreshToken: '' }); // Empty string since we can't read httpOnly cookies
    console.log('[auth-utils] Backend logout successful');
  } catch (error) {
    console.warn('[auth-utils] Backend logout failed, but continuing with local cleanup:', error);
  }
  
  // Clear local storage as fallback
  if (typeof window !== 'undefined') {
    localStorage.removeItem('user');
  }
  
  console.log('[auth-utils] Logout completed');
}

// Store user data in localStorage (fallback)
export function storeUserInLocalStorage(userData: UserLoginResponseBody | UserRegistrationResponseBody) {
  if (typeof window === 'undefined') return;
  if (isLoginResponse(userData) || isRegistrationResponse(userData)) {
    try {
      localStorage.setItem('user', JSON.stringify(userData));
      console.log('📦 Stored user data in localStorage');
    } catch (error) {
      console.error('Error writing to localStorage:', error);
    }
  }
}

// Clear user data from localStorage (fallback)
export function clearUserFromLocalStorage() {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem('user');
    console.log('📦 Cleared user data from localStorage');
  } catch (error) {
    console.error('Error clearing localStorage:', error);
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
        console.log('📦 Retrieved user data from localStorage');
        return userData;
      }
    }
  } catch (error) {
    console.error('Error reading from localStorage:', error);
  }
  return null;
} 