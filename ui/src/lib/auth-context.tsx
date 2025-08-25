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
  slug?: string;
}

interface AuthContextType {
  user: User | null | undefined;
  isLoading: boolean;
  login: (userData: UserLoginResponseBody | UserRegistrationResponseBody) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  updateUserWithSubforumData: (userData: UserLoginResponseBody) => void;
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
        // Try to authenticate using tokens in cookies
        const authResult = await authenticateUser();
        
        		// Debug: checkAuth result logged
        
        if (authResult) {
          // If we already have a user, update their data instead of calling login
          if (user) {
            // Update existing user with fresh data from server
            const updatedUser: User = {
              ...user,
              roles: authResult.roles || user.roles,
              capabilities: authResult.capabilities || user.capabilities,
              activePseudonymId: authResult.activePseudonymId || user.activePseudonymId,
              displayName: authResult.displayName || user.displayName,
              pseudonyms: authResult.pseudonyms || user.pseudonyms,
            };
            setUser(updatedUser);
            
            // Update localStorage
            const userDataToStore = {
              ...updatedUser,
              accessToken: undefined,
              refreshToken: undefined,
            };
            localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
          } else {
            // No existing user, use login function
            await login(authResult);
          }
        } else {
          // Clear any stale localStorage data when server says user is not authenticated
          localStorage.removeItem('hashpost_user');
          // Don't set user to null here - let it remain undefined until we're sure
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
            localStorage.removeItem('hashpost_user');
            // Don't set user to null here
          } else {
            // Refresh failed: log out and redirect
            await logout(getNearestUnprotectedPage());
          }
          return;
        }
        console.error('Error checking authentication:', error);
        // Clear invalid data
        localStorage.removeItem('hashpost_user');
        // Don't set user to null here
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (userData: UserLoginResponseBody | UserRegistrationResponseBody) => {
    // Handle both login and registration responses
    // Login response has pseudonyms array, registration response has pseudonymId
    const isLoginResponse = 'pseudonyms' in userData;
    
    // Debug logging for capabilities
    		// Debug: userData capabilities and roles logged
    
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
    
    		// Debug: normalized user capabilities logged
    
    setUser(normalizedUser);
    // Store user data in localStorage (excluding sensitive tokens)
    const userDataToStore = {
      ...normalizedUser,
      // Don't store tokens in localStorage - they're in cookies
      accessToken: undefined,
      refreshToken: undefined,
    };
    localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
    
    // After login, immediately fetch full user data to get capabilities
    // This is needed because login response has empty capabilities
    		// Debug: login complete, fetching full user data
    try {
      const fullUserData = await authenticateUser();
      if (fullUserData && fullUserData.capabilities && fullUserData.capabilities.length > 0) {
        		// Debug: got full user data with capabilities
        // Update user with full capabilities
        const updatedUser: User = {
          ...normalizedUser,
          roles: fullUserData.roles || normalizedUser.roles,
          capabilities: fullUserData.capabilities || normalizedUser.capabilities,
          activePseudonymId: fullUserData.activePseudonymID || normalizedUser.activePseudonymId,
          displayName: fullUserData.displayName || normalizedUser.displayName,
          pseudonyms: fullUserData.pseudonyms || normalizedUser.pseudonyms,
        };
        setUser(updatedUser);
        
        // Update localStorage
        const userDataToStore = {
          ...updatedUser,
          accessToken: undefined,
          refreshToken: undefined,
        };
        localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
      }
    } catch (error) {
      console.error('Failed to fetch full user data after login:', error);
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
      localStorage.removeItem('hashpost_user');
      if (redirectPath) {
        router.replace(redirectPath);
      }
    }
  };

  // Refresh user data from server
  const refreshUser = async () => {
    		// Debug: refreshUser called, current user status logged
    try {
      const authResult = await authenticateUser();
      if (authResult) {
        // Update existing user data instead of calling login to avoid infinite loop
        if (user) {
          			// Debug: updating existing user with fresh data
          const updatedUser: User = {
            ...user,
            roles: authResult.roles || user.roles,
            capabilities: authResult.capabilities || user.capabilities,
            activePseudonymId: authResult.activePseudonymID || user.activePseudonymId,
            displayName: authResult.displayName || user.displayName,
            pseudonyms: authResult.pseudonyms || user.pseudonyms,
            accessToken: user.accessToken, // Keep existing tokens
            refreshToken: user.refreshToken,
          };
          setUser(updatedUser);
          
          // Update localStorage
          const userDataToStore = {
            ...updatedUser,
            accessToken: undefined,
            refreshToken: undefined,
          };
          localStorage.setItem('hashpost_user', JSON.stringify(userDataToStore));
        } else {
          		// Debug: no existing user, calling login
          // No existing user, use login function
          login(authResult);
        }
      } else {
        		// Debug: no auth result, clearing user
        setUser(null);
        localStorage.removeItem('hashpost_user');
      }
    } catch (error) {
      console.error('Error refreshing user:', error);
      if (error instanceof AuthRefreshFailedError) {
        await logout();
      }
    }
  };

  const updateUserWithSubforumData = (userData: UserLoginResponseBody) => {
    if (!user) return; // Safety check
    
    const normalizedUser: User = {
      ...user, // Keep existing user data
      // Update with subforum-specific data
      roles: userData.roles || user.roles,
      capabilities: userData.capabilities || user.capabilities,
      activePseudonymId: userData.activePseudonymId,
      displayName: userData.displayName,
      pseudonyms: userData.pseudonyms || user.pseudonyms,
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

  const value: AuthContextType = {
    user,
    isLoading,
    login,
    logout,
    refreshUser,
    updateUserWithSubforumData,
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