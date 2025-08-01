'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft, Users, MessageSquare } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import { SubscribeButton } from '@/components/SubscribeButton';
import { toast } from 'sonner';
import { useAuth } from '@/lib/auth-context';
import type { Subforum } from '@/generated/api/src/models';
import type { CommunityType } from '@/lib/community-config';

export default function SubscriptionsPage() {
  const [subscriptions, setSubscriptions] = useState<Subforum[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { user, isAuthenticated } = useAuth();

  const loadSubscriptions = async () => {
    if (!isAuthenticated || !user?.activePseudonymId) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const subforumsApi = getApi(SubforumsApi);
      const response = await subforumsApi.getPseudonymSubscriptions(user.activePseudonymId);
      setSubscriptions(response.subforums || []);
    } catch (err: unknown) {
      console.error('Error loading subscriptions:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load subscriptions';
      setError(errorMessage);
      toast.error('Failed to load subscriptions', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadSubscriptions();
  }, [isAuthenticated, user?.activePseudonymId]);

  const handleSubscriptionChange = (subforumName: string, isSubscribed: boolean) => {
    if (!isSubscribed) {
      // Remove from local state when unsubscribed
      setSubscriptions(prev => prev.filter(sub => sub.name !== subforumName));
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <Link href="/forums">
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Forums
            </Button>
          </Link>
        </div>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Please log in to view your subscriptions</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <Link href="/forums">
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Forums
            </Button>
          </Link>
        </div>
        <div className="space-y-4">
          <div className="h-20 bg-muted animate-pulse rounded" />
          <div className="h-20 bg-muted animate-pulse rounded" />
          <div className="h-20 bg-muted animate-pulse rounded" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <Link href="/forums">
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Forums
            </Button>
          </Link>
        </div>
        <div className="text-center py-12">
          <p className="text-muted-foreground">{error}</p>
          <Button 
            variant="outline" 
            onClick={loadSubscriptions}
            className="mt-4"
          >
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex items-center gap-4 mb-8">
        <Link href="/forums">
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Forums
          </Button>
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">My Subscriptions</h1>
        <p className="text-muted-foreground">
          {subscriptions.length === 0 
            ? "You haven't subscribed to any subforums yet."
            : `You're subscribed to ${subscriptions.length} subforum${subscriptions.length === 1 ? '' : 's'}.`
          }
        </p>
      </div>

      {subscriptions.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">
            Start exploring subforums to build your feed
          </p>
          <Link href="/forums">
            <Button>Browse Forums</Button>
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {subscriptions.map((subforum) => (
            <Card key={subforum.name} className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <CardTitle className="text-lg">
                      <Link 
                        href={`/${subforum.communityType}/${subforum.name}`}
                        className="hover:text-primary transition-colors"
                      >
                        {subforum.communityType}/{subforum.name}
                      </Link>
                    </CardTitle>
                    <CardDescription className="mt-1">
                      {subforum.description}
                    </CardDescription>
                  </div>
                  <div className="ml-4">
                                         <SubscribeButton
                       communityType={subforum.communityType as CommunityType}
                       subforumName={subforum.name}
                       isSubscribed={true}
                       subscriberCount={subforum.subscriberCount || 0}
                       onSubscriptionChange={(isSubscribed) => 
                         handleSubscriptionChange(subforum.name, isSubscribed)
                       }
                     />
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Users className="w-4 h-4" />
                    <span>{subforum.subscriberCount?.toLocaleString() || 0} members</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <MessageSquare className="w-4 h-4" />
                    <span>{subforum.postCount?.toLocaleString() || 0} posts</span>
                  </div>
                  <span>Created {new Date(subforum.createdAt).toLocaleDateString()}</span>
                </div>
                {(subforum.isPrivate || subforum.isNsfw) && (
                  <div className="flex gap-2 mt-2">
                    {subforum.isPrivate && (
                      <span className="text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded">
                        Private
                      </span>
                    )}
                    {subforum.isNsfw && (
                      <span className="text-xs bg-destructive/10 text-destructive px-2 py-1 rounded">
                        NSFW
                      </span>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
} 