import { Configuration } from '../generated/api/src/runtime';

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';

// Centralized configuration for all generated API clients
export const apiConfig = new Configuration({
  basePath: API_BASE_URL,
  credentials: 'include', // Always send cookies (for JWT)
  headers: {
    'Content-Type': 'application/json',
  },
});

// Helper to get a pre-configured API instance
// Usage: const usersApi = getApi(UsersApi);
export function getApi<T extends { new(config: Configuration): InstanceType<T> }>(ApiClass: T): InstanceType<T> {
  return new ApiClass(apiConfig);
}