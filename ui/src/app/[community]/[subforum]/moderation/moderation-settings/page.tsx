'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Label } from '@/components/shadcn/label';
import { Switch } from '@/components/shadcn/switch';
import { Save, ArrowLeft } from 'lucide-react';
import { type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { authenticateUserForSubforum } from '@/lib/auth-utils';
import { toast } from 'sonner';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import { SubforumSettings } from '@/generated/api/src/models';

export default function ModerationSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;

  const { login } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const [settings, setSettings] = useState<SubforumSettings | null>(null);
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
        const response = await subforumsApi.getSubforumSettings(communityType, subforumName);
        setSettings(response.settings);
      } catch (error: unknown) {
        if (error && typeof error === 'object' && 'status' in error && error.status === 403) {
          toast.error('You do not have permission to view moderation settings');
          router.push(`/${fullSubforumPath}/moderation`);
          return;
        }
        throw new Error('Failed to load settings');
      }
    } catch (error) {
      console.error('Error loading settings:', error);
      toast.error('Failed to load moderation settings');
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
      toast.success('Moderation settings saved successfully');
    } catch (error) {
      console.error('Error saving settings:', error);
      toast.error('Failed to save moderation settings');
    } finally {
      setSaving(false);
    }
  };

  // Check permissions
  if (isLoading || !user || (isAuthenticated && !subforumContextLoaded)) {
    return (
      <div className="max-w-7xl mx-auto p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions
  const hasModerateContent = user?.capabilities?.includes('moderate_content') || false;

  if (!hasModerateContent) {
    return (
      <div className="max-w-7xl mx-auto p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">You do not have permission to access this page.</p>
          <p className="text-sm text-muted-foreground mt-2">
            Debug: Auth={isAuthenticated}, Mod={hasModerateContent}, Caps={user?.capabilities?.join(', ')}
          </p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading moderation settings...</p>
        </div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="max-w-7xl mx-auto p-4">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Failed to load moderation settings</p>
          <Button onClick={loadSettings} className="mt-4">
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-4">
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
          <h1 className="text-3xl font-bold mb-2">Moderation Settings</h1>
          <p className="text-muted-foreground">
            Configure automated moderation features for {communityType}/{subforumName}
          </p>
        </div>

        {/* Moderation Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Automated Moderation</CardTitle>
            <CardDescription>
              Configure automated moderation features to help manage content
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label htmlFor="auto-mod">Auto-moderation</Label>
                <Switch
                  id="auto-mod"
                  checked={settings.autoModerationEnabled}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, autoModerationEnabled: checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="require-approval">Require Approval</Label>
                <Switch
                  id="require-approval"
                  checked={settings.requireApproval}
                  onCheckedChange={(checked) =>
                    setSettings({ ...settings, requireApproval: checked })
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
            {saving ? 'Saving...' : 'Save Moderation Settings'}
          </Button>
        </div>
      </div>
    </div>
  );
} 