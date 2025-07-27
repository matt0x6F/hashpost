'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import type { SubforumDetailsResponseBody } from '@/generated/api/src/models/SubforumDetailsResponseBody';
import type { SubforumModerator } from '@/generated/api/src/models/SubforumModerator';
import { PostList } from '@/components/PostList';
import { toast } from 'sonner';
import { useAuth } from '@/lib/auth-context';
import { authenticateUserForSubforum } from '@/lib/auth-utils';

export default function CreatorCommunityPage() {
  const params = useParams();
  const subforumName = params.subforum as string;
  const [forum, setForum] = useState<SubforumDetailsResponseBody | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();

  useEffect(() => {
    if (subforumName) {
      loadForum();
      // Load subforum-specific user context
      loadSubforumUserContext();
    }
  }, [subforumName]);

  const loadSubforumUserContext = async () => {
    try {
      const userData = await authenticateUserForSubforum(`c/${subforumName}`);
      if (userData) {
        login(userData);
      }
    } catch (error) {
      console.error('Error loading subforum user context:', error);
      // Don't show error toast for this - it's not critical
    }
  };

  const loadForum = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const subforumsApi = getApi(SubforumsApi);
      const response = await subforumsApi.getSubforumDetails('c', subforumName);
      setForum(response);
    } catch (err: unknown) {
      console.error('Error loading forum:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load forum';
      setError(errorMessage);
      
      toast.error('Failed to load forum', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <div className="h-8 w-8 bg-muted animate-pulse rounded" />
          <div className="h-8 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="space-y-4">
          <div className="h-20 bg-muted animate-pulse rounded" />
          <div className="h-20 bg-muted animate-pulse rounded" />
          <div className="h-20 bg-muted animate-pulse rounded" />
        </div>
      </div>
    );
  }

  if (error || !forum) {
    return (
      <div className="max-w-6xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <Link href="/forums">
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Forums
            </Button>
          </Link>
        </div>
        <div className="text-center py-12">
          <p className="text-muted-foreground">
            {error || 'Forum not found'}
          </p>
          <Button 
            variant="outline" 
            onClick={loadForum}
            className="mt-4"
          >
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative max-w-7xl mx-auto p-6">
      <div className="flex items-center gap-4 mb-8">
        <Link href="/forums">
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Forums
          </Button>
        </Link>
      </div>

      {/* Main content and sidebar layout */}
      <div className="flex flex-col lg:flex-row">
        {/* Main content */}
        <div className="flex-1 lg:pr-80">
          <div className="mb-8">
            <div className="flex items-center gap-2 mb-2">
              <h1 className="text-3xl font-bold">c/{forum.subforum.name}</h1>
              <span className="text-xs bg-orange-100 text-orange-800 px-2 py-1 rounded">
                Creator
              </span>
              {forum.subforum.isPrivate && (
                <span className="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded">
                  Private
                </span>
              )}
              {forum.subforum.isNsfw && (
                <span className="text-xs bg-destructive/10 text-destructive px-2 py-1 rounded">
                  NSFW
                </span>
              )}
            </div>
            <p className="text-muted-foreground mb-4">
              {forum.subforum.description}
            </p>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span>{forum.subforum.subscriberCount?.toLocaleString() || 0} members</span>
              <span>{forum.subforum.postCount?.toLocaleString() || 0} posts</span>
              <span>Created {new Date(forum.subforum.createdAt).toLocaleDateString()}</span>
            </div>
          </div>

          {/* Posts List */}
          <PostList 
            subforumName={forum.subforum.name} 
          />
        </div>

        {/* Sidebar - fixed on large screens, hugs right edge, dark background */}
        <aside
          className="hidden lg:block fixed top-16 right-0 h-[calc(100vh-4rem)] w-80 z-30 bg-background border-l border-border px-6 py-8"
        >
          <div className="sticky top-8">
            <h2 className="text-lg font-semibold mb-4 text-foreground">c/{forum.subforum.name}</h2>
            <div className="space-y-3 text-sm mb-8">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Subscribers</span>
                <span className="font-medium text-foreground">{forum.subforum.subscriberCount?.toLocaleString() || 0}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Posts</span>
                <span className="font-medium text-foreground">{forum.subforum.postCount?.toLocaleString() || 0}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span className="font-medium text-foreground">{new Date(forum.subforum.createdAt).toLocaleDateString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Governance</span>
                <span className="font-medium text-foreground capitalize">{forum.subforum.governanceStyle}</span>
              </div>
            </div>
            {/* Moderators List */}
            {forum.moderators && forum.moderators.length > 0 && (
              <div className="mb-6">
                <h3 className="text-md font-semibold mb-2 text-foreground">Moderators</h3>
                <ul className="space-y-1">
                  {forum.moderators.map((mod: SubforumModerator) => (
                    <li key={mod.pseudonymId} className="text-sm text-foreground">
                      {mod.displayName || mod.pseudonymId}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </aside>
      </div>
    </div>
  );
} 