"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/shadcn/button";
import { 
  User, 
  Settings, 
  LogOut,
  ChevronDown
} from "lucide-react";
import { useAuth } from "@/lib/auth-context";
import { toast } from "sonner";

export function UserAvatar() {
  const { user, logout, refreshUser } = useAuth();
  const router = useRouter();
  const [showDropdown, setShowDropdown] = useState(false);

  if (!user) {
    return null;
  }

  // Generate initials from handle
  const initials = user.handle
    .split(" ")
    .map(name => name.charAt(0))
    .join("")
    .toUpperCase()
    .slice(0, 2);

  return (
    <>
      <div className="relative">
        <Button
          variant="ghost"
          className="flex items-center space-x-2 px-3 py-2 rounded-md hover:bg-accent"
          onClick={() => setShowDropdown(!showDropdown)}
        >
          <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
            {initials}
          </div>
          <span className="hidden sm:block text-sm font-medium">
            {user.handle}
          </span>
          <ChevronDown className="h-4 w-4" />
        </Button>

        {showDropdown && (
          <div className="absolute right-0 top-full mt-2 w-56 rounded-md shadow-lg bg-background border border-border z-50">
            <div className="py-1">
              {/* User info */}
              <div className="px-4 py-2 border-b border-border">
                <div className="text-sm font-medium text-foreground">
                  {user.handle}
                </div>
                <div className="text-xs text-muted-foreground">
                  {user.email}
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  DID: {user.did}
                </div>
              </div>

              {/* Menu items */}
              <div className="py-1">
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                  onClick={() => {
                    setShowDropdown(false);
                    // Profile viewing not available in atproto system
                    toast.error('Profile viewing is not available in the atproto system');
                  }}
                >
                  <User className="w-4 h-4 mr-2" />
                  Profile
                </button>
                
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                  onClick={() => {
                    setShowDropdown(false);
                    router.push('/settings/account');
                  }}
                >
                  <Settings className="w-4 h-4 mr-2" />
                  Settings
                </button>
                
                <div className="border-t border-border my-1"></div>
                
                <button
                  className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                  onClick={() => {
                    setShowDropdown(false);
                    logout();
                    router.push("/");
                  }}
                >
                  <LogOut className="w-4 h-4 mr-2" />
                  Logout
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
}