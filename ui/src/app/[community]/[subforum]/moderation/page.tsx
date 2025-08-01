'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState, useRef } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Badge } from '@/components/shadcn/badge';
import { Button } from '@/components/shadcn/button';
import { Shield, Users, Flag, Settings, LayoutDashboard, ArrowLeft } from 'lucide-react';
import { DebugUserInfo } from '@/components/DebugUserInfo';
import { EngagementAnalytics } from '@/components/EngagementAnalytics';
import { ModerationStats } from '@/components/ModerationStats';
import { COMMUNITY_CONFIG, type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { authenticateUserForSubforum } from '@/lib/auth-utils';
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
  const { user, isAuthenticated, isLoading, updateUserWithSubforumData } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const hasLoadedContext = useRef(false);

  // Extract community type and subforum name from params
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;
  const communityConfig = COMMUNITY_CONFIG[communityType];

  // Load subforum-specific user context
  const loadSubforumUserContext = async () => {
    if (isAuthenticated && user && subforumName && !hasLoadedContext.current) {
      hasLoadedContext.current = true;
      try {
        const subforumUserData = await authenticateUserForSubforum(fullSubforumPath);
        if (subforumUserData) {
          // Update the user context with subforum-specific capabilities
          updateUserWithSubforumData(subforumUserData);
        }
      } catch (error) {
        console.error('Error loading subforum user context:', error);
      } finally {
        setSubforumContextLoaded(true);
      }
    } else if (!subforumContextLoaded) {
      setSubforumContextLoaded(true);
    }
  };

  useEffect(() => {
    loadSubforumUserContext();
  }, [isAuthenticated, user, subforumName]);

  // Show loading state while checking permissions or loading subforum context
  if (isLoading || !user || (isAuthenticated && !subforumContextLoaded)) {
    console.log('Loading state:', { isLoading, user: !!user, isAuthenticated, subforumContextLoaded });
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions - only after everything is loaded
  const hasModerateContent = user?.capabilities?.includes('moderate_content');
  const hasModeratorRole = user?.roles?.includes('moderator') || user?.roles?.includes('admin');
  const isModerator = hasModeratorRole || hasModerateContent;

  // Debug logging
  console.log('Permission check:', {
    user: user?.email,
    roles: user?.roles,
    capabilities: user?.capabilities,
    hasModerateContent,
    hasModeratorRole,
    isModerator,
    isAuthenticated,
    subforumContextLoaded,
    userExists: !!user
  });

  // Don't render the page if user doesn't have permissions
  if (!isAuthenticated || !isModerator) {
    console.log('Permission denied:', { isAuthenticated, isModerator, user: user?.email });
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-muted-foreground">You do not have permission to access this page.</p>
          <p className="text-sm text-muted-foreground mt-2">
            Debug: Auth={isAuthenticated}, Mod={isModerator}, Roles={user?.roles?.join(', ')}, Caps={user?.capabilities?.filter(c => c.includes('moderate')).join(', ')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Back to Forum Button */}
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${fullSubforumPath}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Forum
          </Button>
        </Link>
      </div>

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
              <NavigationMenuContent className="absolute top-full left-0 mt-1 bg-popover border border-border rounded-md shadow-lg">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2 text-popover-foreground">Report Management</div>
                  <div className="space-y-1 text-sm">
                    <Link href={`/${fullSubforumPath}/moderation/reports`}>
                      <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">View All Reports</div>
                    </Link>
                  </div>
                </div>
              </NavigationMenuContent>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuTrigger className="!flex !flex-row items-center gap-2">
                <Users className="w-4 h-4" />
                Users
              </NavigationMenuTrigger>
              <NavigationMenuContent className="absolute top-full left-0 mt-1 bg-popover border border-border rounded-md shadow-lg">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2 text-popover-foreground">User Management</div>
                  <div className="space-y-1 text-sm">
                    <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Active Users</div>
                    <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Banned Users</div>
                    <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">User Permissions</div>
                  </div>
                </div>
              </NavigationMenuContent>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuTrigger className="!flex !flex-row items-center gap-2">
                <Settings className="w-4 h-4" />
                Settings
              </NavigationMenuTrigger>
              <NavigationMenuContent className="absolute top-full left-0 mt-1 bg-popover border border-border rounded-md shadow-lg">
                <div className="p-4 w-48">
                  <div className="text-sm font-medium mb-2 text-popover-foreground">Moderation Settings</div>
                  <div className="space-y-1 text-sm">
                    <Link href={`/${fullSubforumPath}/moderation/settings`}>
                      <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Subforum Settings</div>
                    </Link>
                    <Link href={`/${fullSubforumPath}/moderation/moderators`}>
                      <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Moderator Team</div>
                    </Link>
                    <Link href={`/${fullSubforumPath}/moderation/content-rules`}>
                      <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Content Rules</div>
                    </Link>
                    <Link href={`/${fullSubforumPath}/moderation/moderation-settings`}>
                      <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Moderation Settings</div>
                    </Link>
                    <div className="p-2 hover:bg-accent hover:text-accent-foreground rounded cursor-pointer transition-colors text-popover-foreground">Auto-moderation</div>
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
              <Link href={`/${fullSubforumPath}/moderation/moderators`}>
                <Button variant="outline" className="justify-start w-full">
                  <Shield className="w-4 h-4 mr-2" />
                  Manage Moderators
                </Button>
              </Link>
              <Link href={`/${fullSubforumPath}/moderation/settings`}>
                <Button variant="outline" className="justify-start w-full">
                  <Settings className="w-4 h-4 mr-2" />
                  Subforum Settings
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Engagement Analytics */}
      <EngagementAnalytics subforumPath={fullSubforumPath} />
    </div>
  );
} 