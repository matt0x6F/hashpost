"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/shadcn/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Badge } from "@/components/shadcn/badge";
import { 
  Users, 
  Shield, 
  Settings, 
  Search, 
  BarChart3, 
  Flag, 
  Gavel,
  Database,
  Key,
  Activity
} from "lucide-react";
import { UserManagementTab } from "./admin/UserManagementTab";
import { ContentModerationTab } from "./admin/ContentModerationTab";
import { SystemSettingsTab } from "./admin/SystemSettingsTab";
import { AnalyticsTab } from "./admin/AnalyticsTab";
import { CorrelationTab } from "./admin/CorrelationTab";
import { useAuth } from "@/lib/auth-context";

export function PlatformAdminDashboard() {
  const { user } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // Get initial tab and search context from URL params
  const initialTab = searchParams.get('tab') || 'users';
  const [activeTab, setActiveTab] = useState(initialTab);

  // Update URL when tab changes
  const handleTabChange = (value: string) => {
    setActiveTab(value);
    
    // Update URL with new tab while preserving other params
    const newParams = new URLSearchParams(searchParams);
    newParams.set('tab', value);
    
    // Remove tab param if it's the default 'users' tab
    if (value === 'users') {
      newParams.delete('tab');
    }
    
    const newUrl = `/admin${newParams.toString() ? `?${newParams.toString()}` : ''}`;
    router.replace(newUrl, { scroll: false });
  };

  // Sync tab state with URL params
  useEffect(() => {
    const tabFromUrl = searchParams.get('tab') || 'users';
    if (tabFromUrl !== activeTab) {
      setActiveTab(tabFromUrl);
    }
  }, [searchParams, activeTab]);

  // Check specific capabilities
  const hasUserManagement = user?.capabilities?.includes("user_management");
  const hasSystemAdmin = user?.capabilities?.includes("system_admin");
  const hasModeration = user?.capabilities?.includes("moderation");
  const hasCompliance = user?.capabilities?.includes("compliance");
  const hasLegalRequests = user?.capabilities?.includes("legal_requests");
  const hasCrossUserCorrelation = user?.capabilities?.includes("cross_user_correlation");

  return (
    <div className="space-y-6">
      {/* Capability Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Admin Capabilities
          </CardTitle>
          <CardDescription>
            Your current platform administration permissions
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {hasUserManagement && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Users className="h-3 w-3" />
                User Management
              </Badge>
            )}
            {hasSystemAdmin && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Settings className="h-3 w-3" />
                System Admin
              </Badge>
            )}
            {hasModeration && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Shield className="h-3 w-3" />
                Platform Moderation
              </Badge>
            )}
            {hasCompliance && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Gavel className="h-3 w-3" />
                Compliance
              </Badge>
            )}
            {hasLegalRequests && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Flag className="h-3 w-3" />
                Legal Requests
              </Badge>
            )}
            {hasCrossUserCorrelation && (
              <Badge variant="secondary" className="flex items-center gap-1">
                <Search className="h-3 w-3" />
                Cross-User Correlation
              </Badge>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Main Admin Interface */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="space-y-4">
        <TabsList className="grid w-full grid-cols-6">
          <TabsTrigger value="users" disabled={!hasUserManagement}>
            <Users className="h-4 w-4 mr-2" />
            Users
          </TabsTrigger>
          <TabsTrigger value="moderation" disabled={!hasModeration}>
            <Shield className="h-4 w-4 mr-2" />
            Moderation
          </TabsTrigger>
          <TabsTrigger value="correlation" disabled={!hasCrossUserCorrelation}>
            <Search className="h-4 w-4 mr-2" />
            Correlation
          </TabsTrigger>
          <TabsTrigger value="analytics" disabled={!hasSystemAdmin}>
            <BarChart3 className="h-4 w-4 mr-2" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="settings" disabled={!hasSystemAdmin}>
            <Settings className="h-4 w-4 mr-2" />
            Settings
          </TabsTrigger>
          <TabsTrigger value="system" disabled={!hasSystemAdmin}>
            <Database className="h-4 w-4 mr-2" />
            System
          </TabsTrigger>
        </TabsList>

        <TabsContent value="users" className="space-y-4">
          <UserManagementTab />
        </TabsContent>

        <TabsContent value="moderation" className="space-y-4">
          <ContentModerationTab />
        </TabsContent>

        <TabsContent value="correlation" className="space-y-4">
          <CorrelationTab />
        </TabsContent>

        <TabsContent value="analytics" className="space-y-4">
          <AnalyticsTab />
        </TabsContent>

        <TabsContent value="settings" className="space-y-4">
          <SystemSettingsTab />
        </TabsContent>

        <TabsContent value="system" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Database Status</CardTitle>
                <Database className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">Healthy</div>
                <p className="text-xs text-muted-foreground">
                  All systems operational
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">IBE Keys</CardTitle>
                <Key className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">Active</div>
                <p className="text-xs text-muted-foreground">
                  Key rotation: 30 days
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Performance</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">98.5%</div>
                <p className="text-xs text-muted-foreground">
                  Uptime this month
                </p>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
