"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { UserLoginResponseBody, UserRegistrationResponseBody } from '@/generated/api/src/models';
import { authenticateUser, logoutUser } from './auth-utils';
import { useRouter } from 'next/navigation';
import { AuthRefreshFailedError } from './auth-utils';

// User interface based on the login response structure
export interface User {
  userId: number;
  email: string;
  createdAt: string;
  lastActiveAt: string;
  isActive: boolean;
  isSuspended: boolean;
  roles: string[];
  capabilities: string[];
  activePseudonymId: string;
  displayName: string;
  pseudonyms: Pseudonym[];
  accessToken: string;
  refreshToken: string;
}

export interface Pseudonym {
  pseudonymId: string;
  displayName: string;
  karmaScore: number;
  createdAt: string;
  lastActiveAt: string;
  isActive: boolean;
}

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (userData: UserLoginResponseBody | UserRegistrationResponseBody) => void;
  logout: () => void;
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
        console.log('Checking authentication...');
        // Try to authenticate using tokens in cookies
        const authResult = await authenticateUser();
        console.log('Authentication result:', authResult);
        
        if (authResult) {
          console.log('User authenticated successfully:', authResult);
          login(authResult);
        } else {
          console.log('Authentication failed - user is not authenticated');
          // Clear any stale localStorage data when server says user is not authenticated
          localStorage.removeItem('hashpost_user');
          setUser(null);
        }
      } catch (error) {
        if (error instanceof AuthRefreshFailedError) {
          // Refresh failed: log out and redirect
          await logout(getNearestUnprotectedPage());
          return;
        }
        console.error('Error checking authentication:', error);
        // Clear invalid data
        localStorage.removeItem('hashpost_user');
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = (userData: UserLoginResponseBody | UserRegistrationResponseBody) => {
    // Handle both login and registration responses
    // Login response has pseudonyms array, registration response has pseudonymId
    const isLoginResponse = 'pseudonyms' in userData;
    
    const normalizedUser: User = {
      userId: userData.userId,
      email: userData.email,
      createdAt: userData.createdAt,
      lastActiveAt: userData.lastActiveAt,
      isActive: userData.isActive,
      isSuspended: userData.isSuspended,
      roles: userData.roles || [],
      capabilities: userData.capabilities || [],
      activePseudonymId: isLoginResponse ? userData.activePseudonymId : userData.pseudonymId,
      displayName: userData.displayName,
      pseudonyms: isLoginResponse ? userData.pseudonyms : [{
        pseudonymId: userData.pseudonymId,
        displayName: userData.displayName,
        karmaScore: userData.karmaScore || 0,
        createdAt: userData.createdAt,
        lastActiveAt: userData.lastActiveAt,
        isActive: userData.isActive,
      }],
      accessToken: userData.accessToken,
      refreshToken: userData.refreshToken,
    };
    
    setUser(normalizedUser);
    // Store user data in localStorage (excluding sensitive tokens)
    const userDataToStore = {
      ...normalizedUser,
      // Don't store tokens in localStorage - they're in cookies
      accessToken: undefined,
      refreshToken: undefined,
    };
    localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
  };

  // Enhance logout to accept optional redirect
  const logout = async (redirectPath?: string) => {
    try {
      await logoutUser();
    } catch (error) {
      console.error('Error during logout:', error);
    } finally {
      setUser(null);
      localStorage.removeItem('hashpost_user');
      if (redirectPath) {
        router.replace(redirectPath);
      }
    }
  };

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    logout,
    isAuthenticated: !!user,
  };

  // Debug logging
  console.log('[auth-context] State update:', { user: !!user, isLoading, isAuthenticated: !!user });

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