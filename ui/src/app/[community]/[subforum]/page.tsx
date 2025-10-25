'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState, useCallback, useRef } from 'react';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft, Hash, MapPin, Tag } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
// Removed SubforumDetailsResponseBody and SubforumModerator - not available in atproto system
import { PostList } from '@/components/PostList';
import { SubforumRulesDisplay } from '@/components/SubforumRulesDisplay';
import { SubscribeButton } from '@/components/SubscribeButton';
import { toast } from 'sonner';
import { useAuth } from '@/lib/auth-context';

import type { CommunityType } from '@/lib/community-config';

export default function SubforumPage() {
  const params = useParams();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;
  
  const [forum, setForum] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();
  const hasLoaded = useRef(false);

  const handleSubscriptionChange = (isSubscribed: boolean, newCount: number) => {
    if (forum) {
      setForum({
        ...forum,
        isSubscribed,
        subforum: {
          ...forum.subforum,
          subscriberCount: newCount,
        },
      });
    }
  };

  const loadSubforumUserContext = useCallback(async () => {
    try {
      // In atproto system, capabilities are handled globally via RBAC
      // No need for subforum-specific authentication
      console.log('Subforum context loading not needed in atproto system');
    } catch (error) {
      console.error('Error loading subforum user context:', error);
    }
  }, [fullSubforumPath, login]);

  const loadForum = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const subforumsApi = getApi(SubforumsApi);
      const subforum = await subforumsApi.getSubforumBySlug(subforumName);
      
      // Transform the API response to match the expected format
      setForum({
        subforum: {
          id: subforum.id,
          name: subforum.name,
          slug: subforum.slug,
          description: subforum.description,
          createdBy: subforum.createdBy,
          createdByHandle: subforum.createdByHandle,
          createdAt: subforum.createdAt,
          updatedAt: subforum.updatedAt,
          subscriberCount: subforum.subscriberCount,
          postCount: subforum.postCount,
          commentCount: subforum.commentCount,
          isPrivate: false, // Default to public
          isNsfw: false, // Default to SFW
        },
        isSubscribed: false, // Default to not subscribed
        moderators: [], // Default to no moderators
      });
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
  }, [communityType, subforumName]);

  useEffect(() => {
    if (subforumName && communityType && !hasLoaded.current) {
      hasLoaded.current = true;
      loadForum();
      loadSubforumUserContext();
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subforumName, communityType]);

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto p-4">
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
      <div className="max-w-7xl mx-auto p-4">
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
    <div className="relative max-w-7xl mx-auto p-2 sm:p-4">
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
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <h1 className="text-3xl font-bold">{communityType}/{forum.subforum.name}</h1>
                {forum.subforum.prefixType === 'h' && (
                  <div className="flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                    <Hash className="w-3 h-3" />
                    <span>HashPost</span>
                  </div>
                )}
                {forum.subforum.prefixType === 'r' && (
                  <div className="flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                    <MapPin className="w-3 h-3" />
                    <span>Regional</span>
                  </div>
                )}
                {forum.subforum.prefixType === 't' && (
                  <div className="flex items-center gap-1 px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
                    <Tag className="w-3 h-3" />
                    <span>Topical</span>
                  </div>
                )}
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
              <SubscribeButton
                subforumSlug={subforumName}
                isSubscribed={forum.isSubscribed}
                subscriberCount={forum.subforum.subscriberCount || 0}
                onSubscriptionChange={handleSubscriptionChange}
              />
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
            subforumName={fullSubforumPath} 
          />
        </div>

        {/* Sidebar - fixed on large screens, hugs right edge, dark background */}
        <aside
          className="hidden lg:block fixed top-16 right-0 h-[calc(100vh-4rem)] w-80 z-30 bg-background border-l border-border px-6 py-8"
        >
          <div className="sticky top-8">
            <h2 className="text-lg font-semibold mb-4 text-foreground">{communityType}/{forum.subforum.name}</h2>
            
            {/* Sidebar Text */}
            {forum.subforum.sidebarText && (
              <div className="mb-6">
                <div className="text-sm text-muted-foreground whitespace-pre-wrap">
                  {forum.subforum.sidebarText}
                </div>
              </div>
            )}
            
            <div className="space-y-3 text-sm mb-6">
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
            </div>
            
            {/* Subscription Button in Sidebar */}
            <div className="mb-6">
              <SubscribeButton
                subforumSlug={subforumName}
                isSubscribed={forum.isSubscribed}
                subscriberCount={forum.subforum.subscriberCount || 0}
                onSubscriptionChange={handleSubscriptionChange}
              />
            </div>
            {/* Moderators List */}
            {forum.moderators && forum.moderators.length > 0 && (
              <div className="mb-6">
                <h3 className="text-md font-semibold mb-2 text-foreground">Moderators</h3>
                <ul className="space-y-1">
                  {forum.moderators.map((mod: any) => (
                    <li key={mod.pseudonymId} className="text-sm text-foreground">
                      {mod.displayName || mod.pseudonymId}
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {/* Subforum Rules */}
            <div className="mt-6">
              <SubforumRulesDisplay 
                communityType={communityType} 
                subforumName={subforumName} 
              />
            </div>
            

          </div>
        </aside>
      </div>
    </div>
  );
} 