'use client';

import { useAuth } from './auth-context';

export interface GuardResult {
  authorized: boolean;
  isLoading: boolean;
  error: string | null;
}

/**
 * Hook to require user authentication
 * Returns loading/error states for pages that require login
 */
export function useRequireAuth(): GuardResult {
  const { user, isLoading, isAuthenticated } = useAuth();

  if (isLoading) {
    return {
      authorized: false,
      isLoading: true,
      error: null,
    };
  }

  if (!isAuthenticated || !user) {
    return {
      authorized: false,
      isLoading: false,
      error: 'You must be logged in to access this page.',
    };
  }

  return {
    authorized: true,
    isLoading: false,
    error: null,
  };
}

/**
 * Hook to require user ownership of a resource
 * Compares the current user's handle with the required owner handle
 */
export function useRequireOwnership(ownerHandle: string): GuardResult {
  const { user, isLoading, isAuthenticated } = useAuth();

  if (isLoading) {
    return {
      authorized: false,
      isLoading: true,
      error: null,
    };
  }

  if (!isAuthenticated || !user) {
    return {
      authorized: false,
      isLoading: false,
      error: 'You must be logged in to access this page.',
    };
  }

  if (user.handle !== ownerHandle) {
    return {
      authorized: false,
      isLoading: false,
      error: 'You can only access your own resources.',
    };
  }

  return {
    authorized: true,
    isLoading: false,
    error: null,
  };
}
