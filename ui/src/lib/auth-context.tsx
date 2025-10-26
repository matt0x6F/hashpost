"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { UserSession, UserLoginResponse, UserRegistrationResponse } from '@/generated/api/src/models';
import { authenticateUser, logoutUser, loginUser, registerUser, storeUserInLocalStorage, clearUserFromLocalStorage, getUserFromLocalStorage, refreshAccessToken } from './auth-utils';
import { setAccessToken, setRefreshToken, restoreTokenFromStorage } from './api-client';
import { isTokenExpired } from './jwt-utils';
import { useRouter } from 'next/navigation';
import { AuthRefreshFailedError } from './auth-utils';
import { capabilitiesService } from './capabilities';

// User interface based on the atproto UserSession structure
export interface User {
  handle: string;
  did: string;
  email: string;
  displayName?: string | null;
  accessToken?: string;
  refreshToken?: string;
}

interface AuthContextType {
  user: User | null | undefined;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, handle: string, inviteCode?: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Utility to determine nearest unprotected page (expand as needed)
  const getNearestUnprotectedPage = () => '/forums';

  // Check for existing authentication on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // First, restore any stored tokens
        const { accessToken, refreshToken } = restoreTokenFromStorage();
        
        // Set the tokens in the API client
        if (accessToken) {
          setAccessToken(accessToken);
        }
        if (refreshToken) {
          setRefreshToken(refreshToken);
        }
        
        // Check if access token is expired and we have a refresh token
        if (accessToken && refreshToken) {
          const isExpired = isTokenExpired(accessToken);
          console.log('[auth-context] Token expiration check:', { 
            hasAccessToken: !!accessToken, 
            hasRefreshToken: !!refreshToken, 
            isExpired 
          });
          
          if (isExpired) {
            console.log('[auth-context] Access token expired, attempting refresh');
            try {
              await refreshAccessToken();
              console.log('[auth-context] Token refreshed successfully');
            } catch (refreshError) {
              console.error('[auth-context] Token refresh failed:', refreshError);
              // If refresh fails, clear everything and continue to auth check
              clearUserFromLocalStorage();
              setAccessToken(null);
            }
          } else {
            console.log('[auth-context] Access token is still valid, proceeding with auth check');
          }
        }
        
        // Try to authenticate using Bearer token
        const authResult = await authenticateUser();
        
        if (authResult) {
          // Convert UserSession to User interface
          const normalizedUser: User = {
            handle: authResult.handle,
            did: authResult.did,
            email: authResult.email,
            displayName: authResult.displayName,
            accessToken: undefined, // Tokens are handled by the API client
            refreshToken: undefined,
          };
          setUser(normalizedUser);
          
          // Store user data in localStorage (excluding sensitive tokens)
          storeUserInLocalStorage({
            handle: authResult.handle,
            did: authResult.did,
            email: authResult.email,
            displayName: authResult.displayName,
            accessToken: '',
            refreshToken: ''
          });
        } else {
          // Don't immediately clear token on first auth failure - might be server issue
          // Only clear user state, keep token for retry
          clearUserFromLocalStorage();
          // Don't call setAccessToken(null) here - keep token for potential retry
        }
      } catch (error) {
        if (error instanceof AuthRefreshFailedError) {
          // Check if we're on a public page that shouldn't redirect
          const currentPath = window.location.pathname;
          const isPublicPage = currentPath.startsWith('/reset-password') || 
                              currentPath.startsWith('/verify-email') ||
                              currentPath === '/';
          
          if (isPublicPage) {
            // Don't redirect on public pages, just clear the user state
            clearUserFromLocalStorage();
            setAccessToken(null);
          } else {
            // Refresh failed: log out and redirect
            await logout(getNearestUnprotectedPage());
          }
          return;
        }
        console.error('Error checking authentication:', error);
        // Don't clear token on general errors - might be network/server issues
        // Only clear user state, keep token for retry
        clearUserFromLocalStorage();
        // Don't call setAccessToken(null) here
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (email: string, password: string) => {
    try {
      const loginResponse = await loginUser(email, password);
      
      // Convert login response to User interface
      const normalizedUser: User = {
        handle: loginResponse.handle,
        did: loginResponse.did,
        email: loginResponse.email,
        displayName: loginResponse.displayName,
        accessToken: loginResponse.accessToken,
        refreshToken: loginResponse.refreshToken,
      };
      
      setUser(normalizedUser);
      
      // Store user data in localStorage (excluding sensitive tokens)
      storeUserInLocalStorage(loginResponse);
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const register = async (email: string, password: string, handle: string, inviteCode?: string) => {
    try {
      const registerResponse = await registerUser(email, password, handle, inviteCode);
      
      // Convert registration response to User interface
      const normalizedUser: User = {
        handle: registerResponse.handle,
        did: registerResponse.did,
        email: registerResponse.email,
        displayName: registerResponse.displayName,
        accessToken: registerResponse.accessToken,
        refreshToken: registerResponse.refreshToken,
      };
      
      setUser(normalizedUser);
      
      // Store user data in localStorage (excluding sensitive tokens)
      storeUserInLocalStorage(registerResponse);
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  };

  // Enhance logout to accept optional redirect
  const logout = async (redirectPath?: string) => {
    try {
      await logoutUser();
    } catch (error) {
      console.error('Error during logout:', error);
    } finally {
      setUser(null);
      clearUserFromLocalStorage();
      // Clear capabilities cache on logout
      capabilitiesService.clearCache();
      if (redirectPath) {
        router.replace(redirectPath);
      }
    }
  };

  // Refresh user data from server
  const refreshUser = async () => {
    try {
      const authResult = await authenticateUser();
      if (authResult) {
        // Convert UserSession to User interface
        const normalizedUser: User = {
          handle: authResult.handle,
          did: authResult.did,
          email: authResult.email,
          displayName: authResult.displayName,
          accessToken: undefined, // Tokens are handled by the API client
          refreshToken: undefined,
        };
        setUser(normalizedUser);
        
        // Store user data in localStorage (excluding sensitive tokens)
        storeUserInLocalStorage({
          handle: authResult.handle,
          did: authResult.did,
          email: authResult.email,
          displayName: authResult.displayName,
          accessToken: '',
          refreshToken: ''
        });
      } else {
        setUser(null);
        clearUserFromLocalStorage();
        setAccessToken(null);
      }
    } catch (error) {
      console.error('Error refreshing user:', error);
      if (error instanceof AuthRefreshFailedError) {
        await logout();
      }
    }
  };

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    register,
    logout,
    refreshUser,
    isAuthenticated: !!user,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}