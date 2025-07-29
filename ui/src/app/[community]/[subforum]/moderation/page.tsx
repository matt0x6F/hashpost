'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Badge } from '@/components/shadcn/badge';
import { Button } from '@/components/shadcn/button';
import { Shield, Users, Flag, Settings, LayoutDashboard } from 'lucide-react';
import { DebugUserInfo } from '@/components/DebugUserInfo';
import { EngagementAnalytics } from '@/components/EngagementAnalytics';
import { ModerationStats } from '@/components/ModerationStats';
import { COMMUNITY_CONFIG, type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { authenticateUserForSubforum } from '@/lib/auth-utils';
import { toast } from 'sonner';
import Link from 'next/link';
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@/components/shadcn/navigation-menu';

export default function SubforumModerationPage() {
  const params = useParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;

  const communityConfig = COMMUNITY_CONFIG[communityType];
  const { login } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);

  // Load subforum-specific user context
  useEffect(() => {
    if (fullSubforumPath && isAuthenticated) {
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, isAuthenticated]);

  const loadSubforumUserContext = async () => {
    try {
      const userData = await authenticateUserForSubforum(fullSubforumPath);
      if (userData) {
        login(userData);
      }
      setSubforumContextLoaded(true);
    } catch (error) {
      console.error('Error loading subforum user context:', error);
      setSubforumContextLoaded(true); // Mark as loaded even on error
    }
  };

  // Check if user has moderator permissions and redirect if not
  useEffect(() => {
    if (!isLoading && isAuthenticated && user && subforumContextLoaded) {
      const hasModerateContent = user.capabilities?.includes('moderate_content');
      const hasModeratorRole = user.roles?.includes('moderator') || user.roles?.includes('admin');
      const isModerator = hasModeratorRole || hasModerateContent;
      
      if (!isModerator) {
        toast.error('You do not have moderator permissions for this subforum');
        router.push(`/${fullSubforumPath}`);
      }
    } else if (!isLoading && !isAuthenticated) {
      // Redirect unauthenticated users
      toast.error('You must be logged in to access the moderation dashboard');
      router.push(`/${fullSubforumPath}`);
    }
  }, [user, isAuthenticated, isLoading, subforumContextLoaded, fullSubforumPath, router]);

  // Show loading state while checking permissions or loading subforum context
  if (isLoading || (isAuthenticated && !subforumContextLoaded)) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Checking permissions...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions
  const hasModerateContent = user?.capabilities?.includes('moderate_content');
  const hasModeratorRole = user?.roles?.includes('moderator') || user?.roles?.includes('admin');
  const isModerator = hasModeratorRole || hasModerateContent;

  // Don't render the page if user doesn't have permissions (redirect will happen)
  if (!isAuthenticated || !isModerator) {
    return null;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Navigation Menu */}
      <div className="mb-6">
        <NavigationMenu viewport={false}>
          <NavigationMenuList>
            <NavigationMenuItem>
              <NavigationMenuLink className="!flex !flex-row items-center gap-2 px-4 py-2 bg-accent text-accent-foreground rounded-md font-medium">
                <LayoutDashboard className="w-4 h-4 text-accent-foreground" />
                Dashboard
              </NavigationMenuLink>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuTrigger className="!flex !flex-row items-center gap-2">
                <Shield className="w-4 h-4" />
                Reports
              </NavigationMenuTrigger>
              <NavigationMenuContent className="absolute top-full left-0 mt-1">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2">Report Management</div>
                  <div className="space-y-1 text-sm text-muted-foreground">
                    <Link href={`/${fullSubforumPath}/moderation/reports`}>
                      <div className="p-2 hover:bg-accent rounded cursor-pointer">View All Reports</div>
                    </Link>
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Report Settings</div>
                  </div>
                </div>
              </NavigationMenuContent>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuTrigger className="!flex !flex-row items-center gap-2">
                <Users className="w-4 h-4" />
                Users
              </NavigationMenuTrigger>
              <NavigationMenuContent className="absolute top-full left-0 mt-1">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2">User Management</div>
                  <div className="space-y-1 text-sm text-muted-foreground">
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Active Users</div>
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Banned Users</div>
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">User Permissions</div>
                  </div>
                </div>
              </NavigationMenuContent>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuTrigger className="!flex !flex-row items-center gap-2">
                <Settings className="w-4 h-4" />
                Settings
              </NavigationMenuTrigger>
              <NavigationMenuContent className="absolute top-full left-0 mt-1">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2">Moderation Settings</div>
                  <div className="space-y-1 text-sm text-muted-foreground">
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Content Rules</div>
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Auto-moderation</div>
                    <div className="p-2 hover:bg-accent rounded cursor-pointer">Notification Settings</div>
                  </div>
                </div>
              </NavigationMenuContent>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Moderation Dashboard</h1>
        <p className="text-muted-foreground">
          Manage content and users for{' '}
          <Badge variant="secondary" className={communityConfig.color}>
            {fullSubforumPath}
          </Badge>
        </p>
      </div>

      {/* Debug Info */}
      <DebugUserInfo />

      <div className="space-y-4">
        <ModerationStats subforumPath={fullSubforumPath} />

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>
              Common moderation tasks for {fullSubforumPath}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Link href={`/${fullSubforumPath}/moderation/reports`}>
                <Button variant="outline" className="justify-start w-full">
                  <Flag className="w-4 h-4 mr-2" />
                  Review Reports
                </Button>
              </Link>
              <Button variant="outline" className="justify-start">
                <Users className="w-4 h-4 mr-2" />
                Manage Bans
              </Button>
              <Button variant="outline" className="justify-start">
                <Shield className="w-4 h-4 mr-2" />
                Add Moderator
              </Button>
              <Button variant="outline" className="justify-start">
                <Settings className="w-4 h-4 mr-2" />
                Subforum Settings
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Engagement Analytics */}
      <EngagementAnalytics subforumPath={fullSubforumPath} />
    </div>
  );
} 