import { Configuration } from '../generated/api/src/runtime';

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