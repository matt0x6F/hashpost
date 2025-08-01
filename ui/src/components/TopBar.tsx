import React, { useEffect, useState } from "react";
import { Menu, Sun, Moon, Shield, Bell } from "lucide-react";
import Image from "next/image";
import { Button } from "./shadcn/button";
import { LoginDialog } from "./LoginDialog";
import { UserAvatar } from "./UserAvatar";
import { useAuth } from "@/lib/auth-context";
import { useTheme } from "next-themes";
import Link from "next/link";
import { usePathname } from "next/navigation";

interface TopBarProps {
  onMenuClick?: () => void;
}

export default function TopBar({ onMenuClick }: TopBarProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const { theme, setTheme } = useTheme();
  const pathname = usePathname();

  // Check if we're in a subforum context (any of the community type routes)
  const isInSubforumContext = pathname?.match(/^\/([tgbch])\/[^\/]+/);

  // Check if user has moderator permissions
  const hasModerateContent = user?.capabilities?.includes('moderate_content');
  const hasModeratorRole = user?.roles?.includes('moderator') || user?.roles?.includes('admin');
  
  // Show moderation link if user has moderator role or moderate_content capability
  // AND we're in a subforum context where the role would be assigned
  const isModerator = (hasModeratorRole || hasModerateContent) && isInSubforumContext;



  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    setMounted(true);
  }, []);

  const handleLoginSuccess = () => {
    // TODO: Handle successful login (e.g., update user state, redirect, etc.)
  };

  const handleSignupSuccess = () => {
    // TODO: Handle successful signup (e.g., update user state, redirect, etc.)
  };

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  return (
    <header className="w-full h-14 flex items-center justify-between px-6 bg-background border-b border-border shadow-sm z-50">
      <div className="flex items-center gap-2">
        {/* Hamburger for mobile */}
        <button
          className="md:hidden flex items-center justify-center w-9 h-9 rounded-full hover:bg-muted transition-colors mr-2"
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
        {/* Subscriptions Link */}
        {!isLoading && isAuthenticated && (
          <Link href="/subscriptions">
            <Button variant="ghost" size="sm" className="flex items-center gap-2">
              <Bell className="w-4 h-4" />
              <span className="hidden sm:inline">Subscriptions</span>
            </Button>
          </Link>
        )}

        {/* Moderator Dashboard Link */}
        {!isLoading && isAuthenticated && isModerator && (
          <Link href={`/${pathname?.split('/').slice(1, 3).join('/')}/moderation`}>
            <Button variant="ghost" size="sm" className="flex items-center gap-2">
              <Shield className="w-4 h-4" />
              <span className="hidden sm:inline">Moderation</span>
            </Button>
          </Link>
        )}

        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleTheme}
          className="w-9 h-9 p-0"
        >
          {mounted && (
            theme === "dark" ? (
              <Sun className="w-4 h-4" />
            ) : (
              <Moon className="w-4 h-4" />
            )
          )}
          <span className="sr-only">Toggle theme</span>
        </Button>
        
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