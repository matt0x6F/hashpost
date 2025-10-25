import { Configuration } from '../generated/api/src/runtime';
import { ResponseError } from '../generated/api/src/runtime';
import { ModelError } from '../generated/api/src/models/ModelError';
import { AuthenticationApi } from '../generated/api/src/apis/AuthenticationApi';
import { SubforumsApi } from '../generated/api/src/apis/SubforumsApi';
import { PostsApi } from '../generated/api/src/apis/PostsApi';
import { CommentsApi } from '../generated/api/src/apis/CommentsApi';
import { SubscriptionsApi } from '../generated/api/src/apis/SubscriptionsApi';
import { VotingApi } from '../generated/api/src/apis/VotingApi';
import { UserManagementApi } from '../generated/api/src/apis/UserManagementApi';
import { SubforumManagementApi } from '../generated/api/src/apis/SubforumManagementApi';
import { PlatformAdminApi } from '../generated/api/src/apis/PlatformAdminApi';
import { RBACManagementApi } from '../generated/api/src/apis/RBACManagementApi';
import { withTokenRefresh } from './api-interceptor';

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';

// Token storage for Bearer authentication
let accessToken: string | null = null;
let refreshToken: string | null = null;

// Set the access token for Bearer authentication
export function setAccessToken(token: string | null) {
  console.log('[api-client] setAccessToken called with:', token ? `token present (length: ${token.length})` : 'null');
  accessToken = token;
  
  // Persist token to localStorage
  try {
    if (token) {
      localStorage.setItem('hashpost_access_token', token);
      console.log('[api-client] Token stored in localStorage');
    } else {
      localStorage.removeItem('hashpost_access_token');
      console.log('[api-client] Token removed from localStorage');
    }
  } catch (error) {
    console.error('Error storing access token to localStorage:', error);
  }
}

// Get the current access token
export function getAccessToken(): string | null {
  return accessToken;
}

// Set the refresh token for Bearer authentication
export function setRefreshToken(token: string | null) {
  console.log('[api-client] setRefreshToken called with:', token ? `token present (length: ${token.length})` : 'null');
  refreshToken = token;
  
  // Persist token to localStorage
  try {
    if (token) {
      localStorage.setItem('hashpost_refresh_token', token);
      console.log('[api-client] Refresh token stored in localStorage');
    } else {
      localStorage.removeItem('hashpost_refresh_token');
      console.log('[api-client] Refresh token removed from localStorage');
    }
  } catch (error) {
    console.error('Error storing refresh token to localStorage:', error);
  }
}

// Get the current refresh token
export function getRefreshToken(): string | null {
  return refreshToken;
}

// Restore tokens from localStorage on page load (client-side only)
export function restoreTokenFromStorage(): { accessToken: string | null; refreshToken: string | null } {
  // Only run on client-side where localStorage exists
  if (typeof window === 'undefined') {
    console.log('[api-client] restoreTokenFromStorage - running on server, skipping');
    return { accessToken: null, refreshToken: null };
  }
  
  try {
    const storedAccessToken = localStorage.getItem('hashpost_access_token');
    const storedRefreshToken = localStorage.getItem('hashpost_refresh_token');
    
    console.log('[api-client] restoreTokenFromStorage - checking for stored tokens:', {
      hasAccessToken: !!storedAccessToken,
      hasRefreshToken: !!storedRefreshToken
    });
    
    if (storedAccessToken) {
      accessToken = storedAccessToken;
      console.log('[api-client] Access token restored from localStorage, length:', storedAccessToken.length);
    } else {
      console.log('[api-client] No access token found in localStorage');
    }
    
    if (storedRefreshToken) {
      refreshToken = storedRefreshToken;
      console.log('[api-client] Refresh token restored from localStorage, length:', storedRefreshToken.length);
    } else {
      console.log('[api-client] No refresh token found in localStorage');
    }
    
    return { accessToken, refreshToken };
  } catch (error) {
    console.error('Error reading tokens from localStorage:', error);
    return { accessToken: null, refreshToken: null };
  }
}

// Helper to get a pre-configured API instance with authentication
export function getApi<T extends { new(config: Configuration): InstanceType<T> }>(ApiClass: T): InstanceType<T> {
  const config = new Configuration({
    basePath: API_BASE_URL,
    headers: {
      'Content-Type': 'application/json',
    },
    // Always provide a dynamic token getter that checks the current accessToken value
    accessToken: async () => {
      console.log('[api-client] getApi - providing access token:', !!accessToken);
      return accessToken || '';
    },
    // Add middleware to inject Authorization header
    middleware: [
      {
        pre: async (context) => {
          const token = await config.accessToken?.();
          if (token) {
            context.init.headers = {
              ...context.init.headers,
              'Authorization': `Bearer ${token}`
            };
            console.log('[api-client] Added Authorization header to request');
          }
          return context;
        }
      }
    ]
  });

  return new ApiClass(config);
}

// Helper to get a pre-configured API instance with automatic token refresh
export function getApiWithRefresh<T extends { new(config: Configuration): InstanceType<T> }>(ApiClass: T): InstanceType<T> {
  const config = new Configuration({
    basePath: API_BASE_URL,
    headers: {
      'Content-Type': 'application/json',
    },
    // Always provide a dynamic token getter that checks the current accessToken value
    accessToken: async () => {
      console.log('[api-client] getApiWithRefresh - providing access token:', !!accessToken);
      return accessToken || '';
    },
    // Add middleware to inject Authorization header
    middleware: [
      {
        pre: async (context) => {
          const token = await config.accessToken?.();
          if (token) {
            context.init.headers = {
              ...context.init.headers,
              'Authorization': `Bearer ${token}`
            };
            console.log('[api-client] Added Authorization header to request (with refresh)');
          }
          return context;
        }
      }
    ]
  });

  return new ApiClass(config);
}

// Helper to extract error messages from API responses
export async function extractApiErrorMessage(error: unknown): Promise<string> {
  if (error instanceof ResponseError) {
    try {
      const responseText = await error.response.text();
      
      // Check if response is JSON
      if (error.response.headers.get('content-type')?.includes('application/json')) {
        const errorData = JSON.parse(responseText);
        
        // Try to parse as ModelError
        if (errorData.error && errorData.message) {
          return errorData.message;
        }
        
        // Fallback to generic message
        if (errorData.message) {
          return errorData.message;
        }
      } else {
        // Handle plain text responses (like from http.Error)
        if (responseText && responseText.trim()) {
          return responseText.trim();
        }
      }
      
      return `Request failed with status ${error.response.status}`;
    } catch (parseError) {
      // If JSON parsing fails, try to return the raw text
      try {
        const responseText = await error.response.text();
        if (responseText && responseText.trim()) {
          return responseText.trim();
        }
      } catch (textError) {
        // Ignore text extraction errors
      }
      return `Request failed with status ${error.response.status}`;
    }
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unexpected error occurred';
}

// Export API instances directly
export const authApi = getApi(AuthenticationApi);
export const subforumsApi = getApi(SubforumsApi);
export const postsApi = getApi(PostsApi);
export const commentsApi = getApi(CommentsApi);
export const subscriptionsApi = getApi(SubscriptionsApi);
export const votingApi = getApi(VotingApi);
export const userManagementApi = getApi(UserManagementApi);
export const subforumManagementApi = getApi(SubforumManagementApi);
export const platformAdminApi = getApi(PlatformAdminApi);
export const rbacManagementApi = getApi(RBACManagementApi);

// Export API instances with automatic token refresh
export const authApiWithRefresh = getApiWithRefresh(AuthenticationApi);
export const subforumsApiWithRefresh = getApiWithRefresh(SubforumsApi);
export const postsApiWithRefresh = getApiWithRefresh(PostsApi);
export const commentsApiWithRefresh = getApiWithRefresh(CommentsApi);
export const subscriptionsApiWithRefresh = getApiWithRefresh(SubscriptionsApi);
export const votingApiWithRefresh = getApiWithRefresh(VotingApi);
export const userManagementApiWithRefresh = getApiWithRefresh(UserManagementApi);
export const subforumManagementApiWithRefresh = getApiWithRefresh(SubforumManagementApi);
export const platformAdminApiWithRefresh = getApiWithRefresh(PlatformAdminApi);
export const rbacManagementApiWithRefresh = getApiWithRefresh(RBACManagementApi);