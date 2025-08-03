import { Configuration } from '../generated/api/src/runtime';
import { ResponseError } from '../generated/api/src/runtime';
import { ErrorModelFromJSON } from '../generated/api/src/models/ErrorModel';

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';

// Helper to get a pre-configured API instance with authentication
// Usage: const usersApi = getApi(UsersApi);
export function getApi<T extends { new(config: Configuration): InstanceType<T> }>(ApiClass: T): InstanceType<T> {
  const config = new Configuration({
    basePath: API_BASE_URL,
    credentials: 'include', // Always send cookies (for JWT)
    headers: {
      'Content-Type': 'application/json',
    },
  });

  return new ApiClass(config);
}

// Helper to extract error messages from API responses
export async function extractApiErrorMessage(error: unknown): Promise<string> {
  if (error instanceof ResponseError) {
    try {
      const responseText = await error.response.text();
      const errorData = JSON.parse(responseText);
      
      // Try to parse as ErrorModel
      if (errorData.detail) {
        return errorData.detail;
      }
      
      // Fallback to title or generic message
      if (errorData.title) {
        return errorData.title;
      }
      
      return `Request failed with status ${error.response.status}`;
    } catch (parseError) {
      return `Request failed with status ${error.response.status}`;
    }
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unexpected error occurred';
}