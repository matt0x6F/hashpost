import React from "react";
import { Menu } from "lucide-react";
import Image from "next/image";
import { Button } from "./shadcn/button";
import { LoginDialog } from "./LoginDialog";
import { UserAvatar } from "./UserAvatar";
import { useAuth } from "@/lib/auth-context";

interface TopBarProps {
  onMenuClick?: () => void;
}

export default function TopBar({ onMenuClick }: TopBarProps) {
  const { isAuthenticated, isLoading } = useAuth();

  const handleLoginSuccess = () => {
    // TODO: Handle successful login (e.g., update user state, redirect, etc.)
    console.log("Login successful!");
  };

  const handleSignupSuccess = () => {
    // TODO: Handle successful signup (e.g., update user state, redirect, etc.)
    console.log("Signup successful!");
  };

  return (
    <header className="w-full h-14 flex items-center justify-between px-6 bg-background border-b border-zinc-800 shadow-sm z-50">
      <div className="flex items-center gap-2">
        {/* Hamburger for mobile */}
        <button
          className="md:hidden flex items-center justify-center w-9 h-9 rounded-full hover:bg-zinc-800 transition-colors mr-2"
          aria-label="Open sidebar"
          onClick={onMenuClick}
        >
          <Menu className="w-5 h-5" />
        </button>
        <Image src="/logo.svg" alt="HashPost Logo" height={32} width={32} className="mr-2" />
        <span className="font-bold text-xl tracking-tight">HashPost</span>
        <span className="text-xs text-muted-foreground ml-2">alpha</span>
      </div>
      <div className="flex items-center gap-4">
        {/* Show login button or user avatar based on auth state */}
        {!isLoading && (
          <>
            {isAuthenticated ? (
              <UserAvatar />
            ) : (
              <LoginDialog onLoginSuccess={handleLoginSuccess} onSignupSuccess={handleSignupSuccess}>
                <Button variant="outline" size="sm">
                  Login
                </Button>
              </LoginDialog>
            )}
          </>
        )}
      </div>
    </header>
  );
} 