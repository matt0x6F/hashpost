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
import { UserListTab } from "./admin/UserListTab";
// Removed PseudonymListTab - not part of atproto system
import { ContentModerationTab } from "./admin/ContentModerationTab";
import { SystemSettingsTab } from "./admin/SystemSettingsTab";
import { AnalyticsTab } from "./admin/AnalyticsTab";
import { PDSServersTab } from "./admin/PDSServersTab";
import { useAuth } from "@/lib/auth-context";
import { useCapabilities } from "@/lib/capabilities";

export function PlatformAdminDashboard() {
  const { user } = useAuth();
  const { hasRole, hasPermission } = useCapabilities();
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // Get initial tab and search context from URL params
  const initialTab = searchParams.get('tab') || 'users';
  const [activeTab, setActiveTab] = useState(initialTab);
  const [capabilities, setCapabilities] = useState({
    hasUserManagement: false,
    hasSystemAdmin: false,
    hasModeration: false,
    hasCompliance: false,
    hasLegalRequests: false,
  });

  // Load capabilities on mount
  useEffect(() => {
    const loadCapabilities = async () => {
      try {
        const [
          userManagement,
          systemAdmin,
          moderation,
          compliance,
          legalRequests,
        ] = await Promise.all([
          hasRole('platform_admin') || hasPermission('manage_users'),
          hasRole('platform_admin') || hasPermission('system_admin'),
          hasRole('platform_admin') || hasPermission('moderate_content'),
          hasRole('platform_admin') || hasPermission('compliance'),
          hasRole('platform_admin') || hasPermission('legal_requests'),
        ]);

        setCapabilities({
          hasUserManagement: userManagement,
          hasSystemAdmin: systemAdmin,
          hasModeration: moderation,
          hasCompliance: compliance,
          hasLegalRequests: legalRequests,
        });
      } catch (error) {
        console.error('Failed to load capabilities:', error);
      }
    };

    loadCapabilities();
  }, [hasRole, hasPermission]);

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

  // Use the loaded capabilities
  const { hasUserManagement, hasSystemAdmin, hasModeration, hasCompliance, hasLegalRequests } = capabilities;

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
          {/* Removed pseudonyms tab - not part of atproto system */}
          <TabsTrigger value="moderation" disabled={!hasModeration}>
            <Shield className="h-4 w-4 mr-2" />
            Moderation
          </TabsTrigger>
          <TabsTrigger value="analytics" disabled={!hasSystemAdmin}>
            <BarChart3 className="h-4 w-4 mr-2" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="pds" disabled={!hasSystemAdmin}>
            <Database className="h-4 w-4 mr-2" />
            PDS Servers
          </TabsTrigger>
          <TabsTrigger value="settings" disabled={!hasSystemAdmin}>
            <Settings className="h-4 w-4 mr-2" />
            Settings
          </TabsTrigger>
          <TabsTrigger value="system" disabled={!hasSystemAdmin}>
            <Key className="h-4 w-4 mr-2" />
            System
          </TabsTrigger>
        </TabsList>

        <TabsContent value="users" className="space-y-4">
          <UserListTab />
        </TabsContent>

        {/* Removed pseudonyms tab content - not part of atproto system */}

        <TabsContent value="moderation" className="space-y-4">
          <ContentModerationTab />
        </TabsContent>


        <TabsContent value="analytics" className="space-y-4">
          <AnalyticsTab />
        </TabsContent>

        <TabsContent value="pds" className="space-y-4">
          <PDSServersTab />
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
