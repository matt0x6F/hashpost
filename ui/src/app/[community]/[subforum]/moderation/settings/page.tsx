'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Badge } from '@/components/shadcn/badge';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Textarea } from '@/components/shadcn/textarea';
import { Switch } from '@/components/shadcn/switch';
import { Separator } from '@/components/shadcn/separator';
import { Settings, Save, ArrowLeft } from 'lucide-react';
import { DebugUserInfo } from '@/components/DebugUserInfo';
import { COMMUNITY_CONFIG, type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { authenticateUserForSubforum } from '@/lib/auth-utils';
import { toast } from 'sonner';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import { SubforumSettings } from '@/generated/api/src/models';
import { SubforumRulesManager } from '@/components/SubforumRulesManager';

export default function SubforumSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;

  const communityConfig = COMMUNITY_CONFIG[communityType];
  const { login } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const [settings, setSettings] = useState<SubforumSettings | null>(null);
  const [subforumId, setSubforumId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

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

  // Load settings
  useEffect(() => {
    if (isAuthenticated && subforumContextLoaded) {
      loadSettings();
    }
  }, [isAuthenticated, subforumContextLoaded]);

  const loadSettings = async () => {
    try {
      setLoading(true);
      const subforumsApi = getApi(SubforumsApi);
      
      try {
        // Call the API method and get the response
        const response = await subforumsApi.getSubforumSettings(communityType, subforumName);
        
        // Use the actual settings from the API response
        setSettings(response.settings);
        setSubforumId(response.subforumId);
      } catch (error: unknown) {
        if (error && typeof error === 'object' && 'status' in error && error.status === 403) {
          toast.error('You do not have permission to view subforum settings');
          router.push(`/${fullSubforumPath}/moderation`);
          return;
        }
        throw new Error('Failed to load settings');
      }
    } catch (error) {
      console.error('Error loading settings:', error);
      toast.error('Failed to load subforum settings');
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    if (!settings) return;

    try {
      setSaving(true);
      const subforumsApi = getApi(SubforumsApi);
      
      await subforumsApi.updateSubforumSettings(communityType, subforumName, settings);
      
      toast.success('Settings saved successfully');
    } catch (error) {
      console.error('Error saving settings:', error);
      toast.error('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  // Check if user has moderator permissions and redirect if not
  useEffect(() => {
    if (!isLoading && isAuthenticated && user && subforumContextLoaded) {
      const hasManageSettings = user.capabilities?.includes('manage_subforum_settings');
      
      if (!hasManageSettings) {
        toast.error('You do not have permission to access subforum settings');
        router.push(`/${fullSubforumPath}`);
      }
    } else if (!isLoading && !isAuthenticated) {
      // Redirect unauthenticated users
      toast.error('You must be logged in to access subforum settings');
      router.push(`/${fullSubforumPath}`);
    }
  }, [user, isAuthenticated, isLoading, subforumContextLoaded, fullSubforumPath, router]);

  // Show loading state while checking permissions or loading subforum context
  if (isLoading || (isAuthenticated && !subforumContextLoaded) || loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading settings...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions
  const hasManageSettings = user?.capabilities?.includes('manage_subforum_settings');

  // Don't render the page if user doesn't have permissions (redirect will happen)
  if (!isAuthenticated || !hasManageSettings) {
    return null;
  }

  if (!settings) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Failed to load settings</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Link href={`/${fullSubforumPath}/moderation`}>
            <Button variant="ghost" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Moderation
            </Button>
          </Link>
        </div>
        <h1 className="text-3xl font-bold mb-2">Subforum Settings</h1>
        <p className="text-muted-foreground">
          Configure settings for{' '}
          <Badge variant="secondary" className={communityConfig.color}>
            {fullSubforumPath}
          </Badge>
          {' '}• Moderation Dashboard
        </p>
      </div>

      {/* Debug Info */}
      <DebugUserInfo />

      <div className="space-y-6">
        {/* Privacy Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Privacy Settings</CardTitle>
            <CardDescription>
              Configure privacy and access settings for this subforum
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center justify-between">
                <Label htmlFor="is-private">Private Subforum</Label>
                <Switch
                  id="is-private"
                  checked={settings.isPrivate}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, isPrivate: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="is-restricted">Restricted Subforum</Label>
                <Switch
                  id="is-restricted"
                  checked={settings.isRestricted}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, isRestricted: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="is-nsfw">NSFW Content</Label>
                <Switch
                  id="is-nsfw"
                  checked={settings.isNsfw}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, isNsfw: checked })
                  }
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Community Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Community Settings</CardTitle>
            <CardDescription>
              Configure description and sidebar text for this subforum
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Enter a description for this subforum..."
                value={settings.description}
                onChange={(e) =>
                  setSettings({ ...settings, description: e.target.value })
                }
                rows={3}
              />
            </div>
            <div>
              <Label htmlFor="sidebar-text">Sidebar Text</Label>
              <Textarea
                id="sidebar-text"
                placeholder="Enter sidebar text for this subforum..."
                value={settings.sidebarText}
                onChange={(e) =>
                  setSettings({ ...settings, sidebarText: e.target.value })
                }
                rows={4}
              />
            </div>
          </CardContent>
        </Card>

        {/* Rules Management */}
        {subforumId && (
          <SubforumRulesManager 
            communityType={communityType}
            subforumName={subforumName}
            subforumId={subforumId}
          />
        )}

        {/* Save Button */}
        <div className="flex justify-end">
          <Button onClick={saveSettings} disabled={saving}>
            <Save className="w-4 h-4 mr-2" />
            {saving ? 'Saving...' : 'Save Settings'}
          </Button>
        </div>
      </div>
    </div>
  );
} 