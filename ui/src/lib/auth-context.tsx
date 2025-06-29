"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { UserLoginResponseBody, UserRegistrationResponseBody } from '@/generated/api/src/models';
import { authenticateUser, logoutUser } from './auth-utils';

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
          console.log('Authentication failed, checking localStorage...');
          // Check if we have user data in localStorage as fallback
          const storedUser = localStorage.getItem('hashpost_user');
          if (storedUser) {
            try {
              const userData = JSON.parse(storedUser);
              console.log('Using stored user data from localStorage');
              setUser(userData);
            } catch (error) {
              console.error('Error parsing stored user data:', error);
              localStorage.removeItem('hashpost_user');
            }
          } else {
            console.log('No stored user data found');
          }
        }
      } catch (error) {
        console.error('Error checking authentication:', error);
        // Clear invalid data
        localStorage.removeItem('hashpost_user');
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

  const logout = async () => {
    try {
      await logoutUser();
    } catch (error) {
      console.error('Error during logout:', error);
    } finally {
      setUser(null);
      localStorage.removeItem('hashpost_user');
    }
  };

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    logout,
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