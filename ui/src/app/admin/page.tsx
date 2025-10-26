"use client";

import { useRequireAuth } from "@/lib/route-guards";
import { UnauthorizedPage } from "@/components/UnauthorizedPage";
import { PlatformAdminDashboard } from "@/components/PlatformAdminDashboard";

export default function AdminPage() {
  const { authorized, isLoading, error } = useRequireAuth();

  // Show loading state while auth is loading
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Show unauthorized page if not authenticated
  if (!authorized) {
    return (
      <UnauthorizedPage
        title="Admin Access Required"
        message={error || "You must be logged in to access the admin panel."}
        icon="lock"
      />
    );
  }

  return (
    <div className="w-max-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Platform Administration</h1>
        <p className="text-muted-foreground mt-2">
          Manage users, content, and platform-wide settings
        </p>
      </div>
      
      <PlatformAdminDashboard />
    </div>
  );
}
