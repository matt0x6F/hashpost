"use client";

import React, { useState } from "react";
import { Button } from "./shadcn/button";
import { useAuth } from "@/lib/auth-context";
import { LogOut, User, Settings, Users } from "lucide-react";
import { CreatePseudonymDialog } from "./CreatePseudonymDialog";
import { switchPseudonym } from "@/lib/pseudonym-utils";
import { usePathname } from "next/navigation";
import { authenticateUserForSubforum } from "@/lib/auth-utils";


export function UserAvatar() {
  const { user, logout, refreshUser, updateUserWithSubforumData } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [showPseudonymManager, setShowPseudonymManager] = useState(false);
  const [showCreatePseudonym, setShowCreatePseudonym] = useState(false);
  const pathname = usePathname();

  // Check if we're in a subforum context (any of the community type routes)
  const isInSubforumContext = pathname?.match(/^\/([tgbch])\/[^\/]+/);
  const subforumMatch = pathname?.match(/^\/([tgbch])\/([^\/]+)/);
  const subforumName = subforumMatch ? `${subforumMatch[1]}/${subforumMatch[2]}` : null;


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
    <>
      <div className="relative">
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 rounded-full p-0 hover:bg-foreground/10 dark:hover:bg-foreground/20"
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
            <div className="absolute right-0 top-full mt-2 w-56 rounded-md shadow-lg bg-background border border-border z-50">
              <div className="py-1">
                {/* User info */}
                <div className="px-4 py-2 border-b border-border">
                  <div className="text-sm font-medium text-foreground">
                    {user.displayName}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {user.email}
                  </div>
                  <div className="text-xs text-muted-foreground mt-1">
                    {user.pseudonyms.length} pseudonym{user.pseudonyms.length !== 1 ? 's' : ''}
                  </div>
                </div>

                {/* Menu items */}
                <div className="py-1">
                  <button
                    className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                    onClick={() => {
                      setShowPseudonymManager(true);
                      setShowDropdown(false);
                    }}
                  >
                    <Users className="w-4 h-4 mr-2" />
                    Manage Pseudonyms
                  </button>
                  
                  <button
                    className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                    onClick={() => {
                      setShowDropdown(false);
                      // TODO: Navigate to profile
                    }}
                  >
                    <User className="w-4 h-4 mr-2" />
                    Profile
                  </button>
                  
                  <button
                    className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
                    onClick={() => {
                      setShowDropdown(false);
                      // TODO: Navigate to settings
                    }}
                  >
                    <Settings className="w-4 h-4 mr-2" />
                    Settings
                  </button>
                  
                  <button
                    className="flex items-center w-full px-4 py-2 text-sm text-foreground hover:bg-foreground/10 dark:hover:bg-foreground/10 transition-colors"
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



      {/* Pseudonym Manager Modal Outside Dropdown */}
      {showPseudonymManager && (
        <>
          <div 
            className="fixed inset-0 bg-black/50 z-[9999]"
            onClick={() => setShowPseudonymManager(false)}
          />
          <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-background border border-border p-6 rounded-lg z-[10000] min-w-[400px] max-h-[80vh] overflow-y-auto shadow-lg">
            <h2 className="text-lg font-bold mb-4 text-foreground">Manage Pseudonyms</h2>
            <div className="space-y-4">
              <div className="p-3 bg-muted rounded-lg">
                <div className="text-sm font-medium text-muted-foreground mb-1">Active Pseudonym</div>
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium text-foreground">{user.displayName}</div>
                    <div className="text-sm text-muted-foreground">
                      {user.pseudonyms.find(p => p.pseudonymId === user.activePseudonymId)?.karmaScore || 0} karma
                    </div>
                  </div>
                  <div className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400/60">
                    <span className="text-sm">Active</span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <div className="text-sm font-medium text-foreground">Your Pseudonyms</div>
                {user.pseudonyms.map((pseudonym) => (
                  <div
                    key={pseudonym.pseudonymId}
                    className={`p-3 border rounded-lg ${
                      pseudonym.pseudonymId === user.activePseudonymId
                        ? "border-primary/50 bg-primary/5 dark:bg-primary/10"
                        : "border-border"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="font-medium text-foreground">{pseudonym.displayName}</div>
                        <div className="text-sm text-muted-foreground">
                          {pseudonym.karmaScore} karma • Created {new Date(pseudonym.createdAt).toLocaleDateString()}
                        </div>
                        {!pseudonym.isActive && (
                          <div className="text-sm text-red-600 mt-1">
                            Deactivated
                          </div>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        {pseudonym.isActive && pseudonym.pseudonymId !== user.activePseudonymId && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={async () => {
                              try {
                                const result = await switchPseudonym(pseudonym.pseudonymId);
                                if (result.success) {
                                  // Refresh user data to get updated active pseudonym
                                  // The new JWT token is automatically set as a cookie
                                  await refreshUser();
                                  
                                  // If we're in a subforum context, also refresh subforum-specific user context
                                  if (isInSubforumContext && subforumName) {
                                    try {
                                      const subforumUserData = await authenticateUserForSubforum(subforumName);
                                      if (subforumUserData) {
                                        // Update the user context with subforum-specific capabilities
                                        updateUserWithSubforumData(subforumUserData);
                                      }
                                    } catch (error) {
                                      console.error('Error refreshing subforum user context:', error);
                                      // Fallback to page reload if subforum context refresh fails
                                      window.location.reload();
                                    }
                                  }
                                  
                                  setShowPseudonymManager(false);
                                } else {
                                  console.error("Failed to switch pseudonym:", result.error);
                                }
                              } catch (error) {
                                console.error("Error switching pseudonym:", error);
                              }
                            }}
                          >
                            Switch
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <div className="pt-4 border-t border-border">
                <Button
                  onClick={() => {
                    setShowCreatePseudonym(true);
                    setShowPseudonymManager(false);
                  }}
                  className="w-full"
                >
                  Create New Pseudonym
                </Button>
              </div>
            </div>
            <div className="mt-4 flex justify-end">
              <Button onClick={() => setShowPseudonymManager(false)}>Close</Button>
            </div>
          </div>
        </>
      )}

      {/* Create Pseudonym Dialog */}
      <CreatePseudonymDialog
        isOpen={showCreatePseudonym}
        onClose={() => setShowCreatePseudonym(false)}
        onSuccess={() => {
          setShowPseudonymManager(true);
        }}
      />
    </>
  );
} 