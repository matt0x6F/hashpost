'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Textarea } from '@/components/shadcn/textarea';
import { Switch } from '@/components/shadcn/switch';
import { Alert, AlertDescription } from '@/components/shadcn/alert';
import { ArrowLeft, Save, User } from 'lucide-react';
import { pseudonymsApi, extractApiErrorMessage } from '@/lib/api-client';
import { useAuth } from '@/lib/auth-context';

interface ProfileFormData {
  display_name: string;
  bio: string;
  website_url: string;
  slug: string;
  show_karma: boolean;
  allow_direct_messages: boolean;
}

export default function ProfileEditPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const slug = params.slug as string;
  
  const [formData, setFormData] = useState<ProfileFormData>({
    display_name: '',
    bio: '',
    website_url: '',
    slug: '',
    show_karma: true,
    allow_direct_messages: true,
  });
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [isOwnProfile, setIsOwnProfile] = useState(false);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        setLoading(true);
        const profile = await pseudonymsApi.getPseudonymProfileBySlug(slug);
        
        setFormData({
          display_name: profile.displayName,
          bio: profile.bio || '',
          website_url: profile.websiteUrl || '',
          slug: profile.slug || '',
          show_karma: profile.showKarma,
          allow_direct_messages: profile.allowDirectMessages,
        });

        // Check if this is the user's own profile
        if (user && user.pseudonyms.some(p => p.pseudonymId === profile.pseudonymId)) {
          setIsOwnProfile(true);
        }
      } catch (err) {
        setError('Failed to load profile');
        console.error('Error fetching profile:', err);
      } finally {
        setLoading(false);
      }
    };

    if (slug) {
      fetchProfile();
    }
  }, [slug, user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isOwnProfile) {
      setError('You can only edit your own profile');
      return;
    }

    try {
      setSaving(true);
      setError(null);
      setSuccess(null);

      // Find the pseudonym ID for this profile
      const profileResponse = await pseudonymsApi.getPseudonymProfileBySlug(slug);
      const pseudonymId = profileResponse.pseudonymId;

      await pseudonymsApi.updatePseudonymProfile(pseudonymId, {
        displayName: formData.display_name,
        bio: formData.bio,
        websiteUrl: formData.website_url,
        slug: formData.slug,
        showKarma: formData.show_karma,
        allowDirectMessages: formData.allow_direct_messages,
      });

      setSuccess('Profile updated successfully!');
      
      // Redirect to the new slug if it changed
      if (formData.slug !== slug) {
        setTimeout(() => {
          router.push(`/profiles/${formData.slug}`);
        }, 1000);
      }
    } catch (err) {
      const errorMessage = await extractApiErrorMessage(err);
      setError(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field: keyof ProfileFormData, value: string | boolean) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-6"></div>
          <div className="space-y-4">
            <div className="h-4 bg-muted rounded w-1/4"></div>
            <div className="h-4 bg-muted rounded w-1/2"></div>
            <div className="h-4 bg-muted rounded w-3/4"></div>
          </div>
        </div>
      </div>
    );
  }

  if (!isOwnProfile) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <h2 className="text-2xl font-bold text-muted-foreground mb-2">Access Denied</h2>
              <p className="text-muted-foreground mb-4">You can only edit your own profile.</p>
              <Button onClick={() => router.back()}>
                <ArrowLeft className="w-4 h-4 mr-2" />
                Go Back
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl">
      <div className="mb-6">
        <Button variant="ghost" onClick={() => router.back()}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Profile
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <User className="w-5 h-5 mr-2" />
            Edit Profile
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}
            
            {success && (
              <Alert>
                <AlertDescription>{success}</AlertDescription>
              </Alert>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <Label htmlFor="display_name">Display Name *</Label>
                <Input
                  id="display_name"
                  value={formData.display_name}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('display_name', e.target.value)}
                  placeholder="Enter your display name"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="slug">Profile URL Slug</Label>
                <Input
                  id="slug"
                  value={formData.slug}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('slug', e.target.value)}
                  placeholder="your-profile-slug"
                />
                <p className="text-sm text-muted-foreground">
                  Leave empty to auto-generate from display name
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="bio">Bio</Label>
              <Textarea
                id="bio"
                value={formData.bio}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('bio', e.target.value)}
                placeholder="Tell us about yourself..."
                rows={4}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="website_url">Website URL</Label>
              <Input
                id="website_url"
                type="url"
                value={formData.website_url}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('website_url', e.target.value)}
                placeholder="https://example.com"
              />
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="show_karma">Show Karma</Label>
                  <p className="text-sm text-muted-foreground">
                    Display your karma score on your profile
                  </p>
                </div>
                <Switch
                  id="show_karma"
                  checked={formData.show_karma}
                  onCheckedChange={(checked: boolean) => handleInputChange('show_karma', checked)}
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="allow_direct_messages">Allow Direct Messages</Label>
                  <p className="text-sm text-muted-foreground">
                    Allow other users to send you direct messages
                  </p>
                </div>
                <Switch
                  id="allow_direct_messages"
                  checked={formData.allow_direct_messages}
                  onCheckedChange={(checked: boolean) => handleInputChange('allow_direct_messages', checked)}
                />
              </div>
            </div>

            <div className="flex justify-end space-x-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => router.back()}
                disabled={saving}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={saving}>
                <Save className="w-4 h-4 mr-2" />
                {saving ? 'Saving...' : 'Save Changes'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
} 