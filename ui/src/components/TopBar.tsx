import React, { useEffect, useState } from "react";
import { Menu, Sun, Moon, Shield } from "lucide-react";
import Image from "next/image";
import { Button } from "./shadcn/button";
import { LoginDialog } from "./LoginDialog";
import { UserAvatar } from "./UserAvatar";
import { useAuth } from "@/lib/auth-context";
import { useTheme } from "next-themes";
import Link from "next/link";
import { usePathname } from "next/navigation";
// Removed authenticateUserForSubforum - not needed in atproto system

interface TopBarProps {
  onMenuClick?: () => void;
}

export default function TopBar({ onMenuClick }: TopBarProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const { theme, setTheme } = useTheme();
  const pathname = usePathname();

  // Check if we're in a subforum context (any of the community type routes)
  const isInSubforumContext = pathname?.match(/^\/([tgbch])\/[^\/]+/);
  const currentSubforumPath = isInSubforumContext ? pathname?.split('/').slice(1, 3).join('/') : null;
  
  // State to track if we've loaded subforum-specific context
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const [subforumUserCapabilities, setSubforumUserCapabilities] = useState<string[]>([]);

  // Load subforum-specific user context when in a subforum
  useEffect(() => {
    if (isAuthenticated && user && currentSubforumPath && !subforumContextLoaded) {
      const loadSubforumContext = async () => {
        try {
          // In atproto system, capabilities are handled globally via RBAC
          // No need for subforum-specific authentication
          console.log('Subforum context loading not needed in atproto system');
        } catch (error) {
          console.error('Error loading subforum user context:', error);
        } finally {
          setSubforumContextLoaded(true);
        }
      };
      
      loadSubforumContext();
    } else if (!isInSubforumContext) {
      // Not in subforum context, reset subforum capabilities
      setSubforumUserCapabilities([]);
      setSubforumContextLoaded(false);
    }
  }, [isAuthenticated, user, currentSubforumPath, subforumContextLoaded, isInSubforumContext]);

  // Reset subforum context when pathname changes (user navigates to different page)
  useEffect(() => {
    if (!isInSubforumContext) {
      setSubforumUserCapabilities([]);
      setSubforumContextLoaded(false);
    }
  }, [pathname, isInSubforumContext]);

  // Check if user has moderator permissions
  // For subforum moderation, only use subforum-specific capabilities loaded fresh for each subforum
  // This prevents capabilities from persisting across different subforums
  const hasSubforumModeration = isInSubforumContext && subforumContextLoaded 
    ? subforumUserCapabilities.includes('moderate_content')
    : false;
  
  // Show moderation link only when user has permissions for the current context
  const isModerator = hasSubforumModeration;

  // Debug logging for moderation permissions
  useEffect(() => {
    if (isAuthenticated && user) {
      console.log('Debug: TopBar moderation check logged', {
        pathname,
        isInSubforumContext,
        currentSubforumPath,
        subforumContextLoaded,
        hasSubforumModeration,
        subforumUserCapabilities,
        userCapabilities: [], // Capabilities not available in atproto system
        isModerator,
        moderationButtonShown: !isLoading && isAuthenticated && isModerator
      });
    }
  }, [isAuthenticated, user, pathname, isInSubforumContext, currentSubforumPath, subforumContextLoaded, hasSubforumModeration, subforumUserCapabilities, isModerator, isLoading]);



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
    <header className="w-full h-14 flex items-center justify-between px-3 sm:px-6 bg-background border-b border-border shadow-sm z-50">
      <div className="flex items-center gap-2 min-w-0">
        {/* Hamburger for mobile */}
        <button
          className="md:hidden flex items-center justify-center w-9 h-9 rounded-full hover:bg-muted transition-colors mr-2 flex-shrink-0"
          aria-label="Open sidebar"
          title="Open sidebar menu"
          onClick={onMenuClick}
        >
          <Menu className="w-5 h-5" />
        </button>
        <Image src="/logo.svg" alt="HashPost Logo" height={32} width={32} className="mr-2 flex-shrink-0" />
        <div className="flex items-center min-w-0">
          <span className="font-bold text-xl tracking-tight truncate">HashPost</span>
          <span className="text-xs text-muted-foreground ml-2 flex-shrink-0">alpha</span>
        </div>
      </div>
      <div className="flex items-center gap-2 sm:gap-4 flex-shrink-0">
        {/* Platform Rules Link */}
        <Link href="/platform-rules">
          <Button 
            variant="ghost" 
            size="sm" 
            className="flex items-center gap-2"
            title="Platform Rules"
          >
            <Shield className="w-4 h-4" />
            <span className="hidden sm:inline">Rules</span>
          </Button>
        </Link>

        {/* Moderator Dashboard Link */}
        {!isLoading && isAuthenticated && isModerator && currentSubforumPath && (
          <Link href={`/${currentSubforumPath}/moderation`}>
            <Button 
              variant="ghost" 
              size="sm" 
              className="flex items-center gap-2"
              title="Moderation Dashboard"
            >
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
          title={mounted ? (theme === "dark" ? "Switch to light mode" : "Switch to dark mode") : "Toggle theme"}
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