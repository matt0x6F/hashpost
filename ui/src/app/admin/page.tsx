"use client";

import { useAuth } from "@/lib/auth-context";
import { redirect } from "next/navigation";
import { PlatformAdminDashboard } from "@/components/PlatformAdminDashboard";

export default function AdminPage() {
  const { isAuthenticated, isLoading } = useAuth();

  // Show loading state while auth is loading
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Redirect if not authenticated
  if (!isAuthenticated) {
    redirect("/");
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
