'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Badge } from '@/components/shadcn/badge';
import { RadioGroup, RadioGroupItem } from '@/components/shadcn/radio-group';
import { Label } from '@/components/shadcn/label';
import { 
  ArrowLeft,
  Settings,
  Eye,
  Users,
  Lock,
  Save
} from 'lucide-react';
import { getApiWithRefresh } from '@/lib/api-client';
import { ProfilesApi } from '@/generated/api/src/apis/ProfilesApi';
import { UserProfile, ProfileVisibility } from '@/generated/api/src/models';
import { useRequireOwnership } from '@/lib/route-guards';
import { UnauthorizedPage } from '@/components/UnauthorizedPage';
import { toast } from 'sonner';

interface ProfileEditPageProps {}

export default function ProfileEditPage({}: ProfileEditPageProps) {
  const params = useParams();
  const router = useRouter();
  const handle = params.handle as string;
  
  // Use the ownership guard hook
  const { authorized, isLoading: authLoading, error: authError } = useRequireOwnership(handle);
  
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [visibility, setVisibility] = useState<ProfileVisibility>('public');
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (handle && authorized) {
      loadProfile();
    }
  }, [handle, authorized]);

  const loadProfile = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const api = getApiWithRefresh(ProfilesApi);
      const profileData = await api.getUserProfile(handle);
      setProfile(profileData);
      setVisibility(profileData.profileVisibility);
    } catch (err: unknown) {
      console.error('Failed to load profile:', err);
      
      if (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'status' in err.response) {
        const status = (err.response as { status: number }).status;
        if (status === 404) {
          setError("User not found");
        } else {
          setError("Failed to load profile");
        }
      } else {
        setError("Failed to load profile");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    if (!authorized) return;
    
    setIsSaving(true);
    try {
      const api = getApiWithRefresh(ProfilesApi);
      await api.updateProfileVisibility({ visibility });
      
      toast.success('Profile visibility updated successfully');
      router.push(`/u/${handle}`);
    } catch (err: unknown) {
      console.error('Failed to update profile visibility:', err);
      toast.error('Failed to update profile visibility');
    } finally {
      setIsSaving(false);
    }
  };

  const getVisibilityIcon = (vis: ProfileVisibility) => {
    switch (vis) {
      case 'public':
        return <Eye className="h-4 w-4" />;
      case 'authenticated':
        return <Users className="h-4 w-4" />;
      case 'private':
        return <Lock className="h-4 w-4" />;
      default:
        return <Eye className="h-4 w-4" />;
    }
  };

  const getVisibilityDescription = (vis: ProfileVisibility) => {
    switch (vis) {
      case 'public':
        return 'Anyone can view your profile and posts';
      case 'authenticated':
        return 'Only logged-in users can view your profile';
      case 'private':
        return 'Only you can view your profile';
      default:
        return 'Anyone can view your profile and posts';
    }
  };

  // Show loading state during auth check
  if (authLoading) {
    return (
      <div className="container mx-auto py-6 max-w-2xl">
        <div className="flex items-center justify-center min-h-64">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  // Show unauthorized page if not authorized
  if (!authorized) {
    return (
      <UnauthorizedPage
        title="Access Denied"
        message={authError || "You can only edit your own profile."}
        icon="user-x"
      />
    );
  }

  // Show loading state while loading profile data
  if (isLoading) {
    return (
      <div className="container mx-auto py-6 max-w-2xl">
        <div className="flex items-center justify-center min-h-64">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  // Show error if profile failed to load
  if (error) {
    return (
      <div className="container mx-auto py-6 max-w-2xl">
        <Card>
          <CardContent className="p-8 text-center">
            <div className="text-muted-foreground mb-4">{error}</div>
            <Button onClick={() => router.back()} variant="outline">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Go Back
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="container mx-auto py-6 max-w-2xl">
        <Card>
          <CardContent className="p-8 text-center">
            <div className="text-muted-foreground">Profile not found</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 max-w-2xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Button onClick={() => router.back()} variant="outline" size="sm">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Profile
        </Button>
        <div>
          <h1 className="text-3xl font-bold">Edit Profile</h1>
          <p className="text-muted-foreground">Manage your profile settings</p>
        </div>
      </div>

      {/* Profile Visibility Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Profile Visibility
          </CardTitle>
          <CardDescription>
            Control who can view your profile and posts
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <RadioGroup value={visibility} onValueChange={(value) => setVisibility(value as ProfileVisibility)}>
            <div className="flex items-start space-x-3 p-4 border rounded-lg">
              <RadioGroupItem value="public" id="public" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="public" className="flex items-center gap-2 cursor-pointer">
                  {getVisibilityIcon('public')}
                  <span className="font-medium">Public</span>
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  {getVisibilityDescription('public')}
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3 p-4 border rounded-lg">
              <RadioGroupItem value="authenticated" id="authenticated" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="authenticated" className="flex items-center gap-2 cursor-pointer">
                  {getVisibilityIcon('authenticated')}
                  <span className="font-medium">Authenticated Users Only</span>
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  {getVisibilityDescription('authenticated')}
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3 p-4 border rounded-lg">
              <RadioGroupItem value="private" id="private" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="private" className="flex items-center gap-2 cursor-pointer">
                  {getVisibilityIcon('private')}
                  <span className="font-medium">Private</span>
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  {getVisibilityDescription('private')}
                </p>
              </div>
            </div>
          </RadioGroup>

          <div className="flex items-center justify-between pt-4 border-t">
            <div className="text-sm text-muted-foreground">
              Current setting: <Badge variant="outline" className="ml-1">
                {getVisibilityIcon(visibility)}
                {visibility === 'public' ? 'Public' : visibility === 'authenticated' ? 'Authenticated Users Only' : 'Private'}
              </Badge>
            </div>
            <Button onClick={handleSave} disabled={isSaving}>
              {isSaving ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Changes
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
