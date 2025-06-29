"use client";

import React, { useState } from "react";
import { Button } from "./shadcn/button";
import { useAuth } from "@/lib/auth-context";
import { LogOut, User, Settings } from "lucide-react";

export function UserAvatar() {
  const { user, logout } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);

  const handleLogout = async () => {
    setIsLoading(true);
    try {
      await logout();
    } catch (error) {
      console.error("Logout failed:", error);
    } finally {
      setIsLoading(false);
      setShowDropdown(false);
    }
  };

  if (!user) {
    return null;
  }

  // Generate initials from display name
  const initials = user.displayName
    .split(" ")
    .map(name => name.charAt(0))
    .join("")
    .toUpperCase()
    .slice(0, 2);

  return (
    <div className="relative">
      <Button
        variant="ghost"
        size="sm"
        className="h-8 w-8 rounded-full p-0 hover:bg-zinc-800"
        onClick={() => setShowDropdown(!showDropdown)}
        disabled={isLoading}
      >
        <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center text-primary-foreground text-sm font-medium">
          {initials}
        </div>
      </Button>

      {showDropdown && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 z-40" 
            onClick={() => setShowDropdown(false)}
          />
          
          {/* Dropdown */}
          <div className="absolute right-0 top-full mt-2 w-56 rounded-md shadow-lg bg-background border border-zinc-800 z-50">
            <div className="py-1">
              {/* User info */}
              <div className="px-4 py-2 border-b border-zinc-800">
                <div className="text-sm font-medium text-foreground">
                  {user.displayName}
                </div>
                <div className="text-xs text-muted-foreground">
                  {user.email}
                </div>
              </div>

              {/* Menu items */}
              <div className="py-1">
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-zinc-800 transition-colors"
                  onClick={() => {
                    setShowDropdown(false);
                    // TODO: Navigate to profile
                  }}
                >
                  <User className="w-4 h-4 mr-2" />
                  Profile
                </button>
                
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-zinc-800 transition-colors"
                  onClick={() => {
                    setShowDropdown(false);
                    // TODO: Navigate to settings
                  }}
                >
                  <Settings className="w-4 h-4 mr-2" />
                  Settings
                </button>
                
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-red-500 hover:bg-zinc-800 transition-colors"
                  onClick={handleLogout}
                  disabled={isLoading}
                >
                  <LogOut className="w-4 h-4 mr-2" />
                  {isLoading ? "Signing out..." : "Sign out"}
                </button>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
} 