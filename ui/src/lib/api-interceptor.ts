/**
 * API interceptor to handle 401 errors with automatic token refresh
 */

import { refreshAccessToken } from './auth-utils';
import { AuthRefreshFailedError } from './auth-utils';

/**
 * Wraps an API call to automatically handle 401 errors with token refresh
 * @param apiCall Function that makes the API call
 * @returns Promise that resolves to the API response
 */
export async function withTokenRefresh<T>(apiCall: () => Promise<T>): Promise<T> {
  try {
    // Try the original API call
    return await apiCall();
  } catch (error) {
    // Check if it's a 401 Unauthorized error
    if (isUnauthorizedError(error)) {
      console.log('[api-interceptor] 401 error detected, attempting token refresh');
      
      try {
        // Attempt to refresh the token
        await refreshAccessToken();
        console.log('[api-interceptor] Token refreshed successfully, retrying original request');
        
        // Retry the original API call with the new token
        return await apiCall();
      } catch (refreshError) {
        console.error('[api-interceptor] Token refresh failed:', refreshError);
        
        // If refresh fails, throw the original error
        // The auth context will handle logout
        throw error;
      }
    }
    
    // For non-401 errors, just re-throw
    throw error;
  }
}

/**
 * Check if an error is a 401 Unauthorized response
 * @param error The error to check
 * @returns true if it's a 401 error
 */
function isUnauthorizedError(error: unknown): boolean {
  if (error && typeof error === 'object' && 'response' in error) {
    const response = (error as { response?: Response }).response;
    return response?.status === 401;
  }
  return false;
}

/**
 * Wraps multiple API calls to handle token refresh for any 401 errors
 * @param apiCalls Array of functions that make API calls
 * @returns Promise that resolves to an array of results
 */
export async function withTokenRefreshAll<T>(apiCalls: (() => Promise<T>)[]): Promise<T[]> {
  const results: T[] = [];
  
  for (const apiCall of apiCalls) {
    try {
      const result = await withTokenRefresh(apiCall);
      results.push(result);
    } catch (error) {
      // If any call fails after refresh, we need to handle it
      // For now, we'll re-throw the error to let the caller handle it
      throw error;
    }
  }
  
  return results;
}

