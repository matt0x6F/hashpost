'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Switch } from '@/components/shadcn/switch';
import { Save, ArrowLeft } from 'lucide-react';
import { type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
// Removed SubforumSettings - not available in atproto system

export default function ContentRulesPage() {
  const params = useParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;

  const { login } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const [settings, setSettings] = useState<any | null>(null);
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
      // In atproto system, capabilities are handled globally via RBAC
      // No need for subforum-specific authentication
      console.log('Subforum context loading not needed in atproto system');
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
      // Content rules not available in atproto system
      setSettings(null);
    } catch (error) {
      console.error('Error loading settings:', error);
      toast.error('Failed to load content rules');
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    if (!settings) return;

    try {
      setSaving(true);
      // Content rules not available in atproto system
      toast.error('Content rules are not available in the atproto system');
    } catch (error) {
      console.error('Error saving settings:', error);
      toast.error('Failed to save content rules');
    } finally {
      setSaving(false);
    }
  };

  // Check permissions
  if (isLoading || !user || (isAuthenticated && !subforumContextLoaded)) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions
  // In atproto system, permissions are handled via RBAC - for now, assume no permissions
  const hasModerateContent = false;

  if (!hasModerateContent) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">You do not have permission to access this page.</p>
          <p className="text-sm text-muted-foreground mt-2">
            Debug: Auth={isAuthenticated}, Mod={hasModerateContent}, Caps=N/A
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading content rules...</p>
        </div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Failed to load content rules</p>
          <Button onClick={loadSettings} className="mt-4">
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-2 sm:p-4">
      <div className="flex items-center gap-4 mb-8">
        <Link href={`/${fullSubforumPath}/moderation`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Moderation
          </Button>
        </Link>
      </div>

      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Content Rules</h1>
          <p className="text-muted-foreground">
            Configure what types of content are allowed in {communityType}/{subforumName}
          </p>
        </div>

        {/* Content Type Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Content Types</CardTitle>
            <CardDescription>
              Control what types of content users can post in this subforum
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label htmlFor="allow-images">Allow Images</Label>
                <Switch
                  id="allow-images"
                  checked={settings.allowImages}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, allowImages: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="allow-videos">Allow Videos</Label>
                <Switch
                  id="allow-videos"
                  checked={settings.allowVideos}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, allowVideos: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="allow-polls">Allow Polls</Label>
                <Switch
                  id="allow-polls"
                  checked={settings.allowPolls}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, allowPolls: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="require-flair">Require Flair</Label>
                <Switch
                  id="require-flair"
                  checked={settings.requireFlair}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, requireFlair: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="allow-crossposts">Allow Crossposts</Label>
                <Switch
                  id="allow-crossposts"
                  checked={settings.allowCrossposts}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, allowCrossposts: checked })
                  }
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Posting Requirements */}
        <Card>
          <CardHeader>
            <CardTitle>Posting Requirements</CardTitle>
            <CardDescription>
              Set minimum requirements for users to post in this subforum
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label htmlFor="min-age">Minimum Account Age (hours)</Label>
                <Input
                  id="min-age"
                  type="number"
                  value={settings.minimumAccountAgeHours}
                  onChange={(e) =>
                    setSettings({
                      ...settings,
                      minimumAccountAgeHours: parseInt(e.target.value) || 0,
                    })
                  }
                />
              </div>
              <div>
                <Label htmlFor="min-karma">Minimum Karma Required</Label>
                <Input
                  id="min-karma"
                  type="number"
                  value={settings.minimumKarmaRequired}
                  onChange={(e) =>
                    setSettings({
                      ...settings,
                      minimumKarmaRequired: parseInt(e.target.value) || 0,
                    })
                  }
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Save Button */}
        <div className="flex justify-end">
          <Button onClick={saveSettings} disabled={saving}>
            <Save className="w-4 h-4 mr-2" />
            {saving ? 'Saving...' : 'Save Content Rules'}
          </Button>
        </div>
      </div>
    </div>
  );
} 