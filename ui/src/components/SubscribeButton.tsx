'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/shadcn/button';
import { Bell, BellOff } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { SubscriptionsApi } from '@/generated/api/src/apis/SubscriptionsApi';
import { toast } from 'sonner';
import { useAuth } from '@/lib/auth-context';

interface SubscribeButtonProps {
  subforumSlug: string;
  isSubscribed: boolean;
  subscriberCount: number;
  onSubscriptionChange?: (isSubscribed: boolean, newCount: number) => void;
}

export function SubscribeButton({
  subforumSlug,
  isSubscribed,
  subscriberCount: initialSubscriberCount,
  onSubscriptionChange,
}: SubscribeButtonProps) {
  const [subscriberCount, setSubscriberCount] = useState(initialSubscriberCount);
  const [isLoading, setIsLoading] = useState(false);
  const { user, isAuthenticated } = useAuth();

  // Sync subscriber count with prop
  useEffect(() => {
    setSubscriberCount(initialSubscriberCount);
  }, [initialSubscriberCount]);

  const handleSubscribe = async () => {
    if (!isAuthenticated) {
      toast.error('Please log in to subscribe to subforums');
      return;
    }



    setIsLoading(true);
    try {
      const subscriptionsApi = getApi(SubscriptionsApi);
      
      if (isSubscribed) {
        // Unsubscribe
        await subscriptionsApi.unsubscribeFromSubforum(subforumSlug);
        setSubscriberCount(prev => Math.max(0, prev - 1));
        toast.success('Unsubscribed from subforum');
      } else {
        // Subscribe
        await subscriptionsApi.subscribeToSubforum(subforumSlug);
        setSubscriberCount(prev => prev + 1);
        toast.success('Subscribed to subforum');
      }

      onSubscriptionChange?.(!isSubscribed, isSubscribed ? subscriberCount - 1 : subscriberCount + 1);
    } catch (error) {
      console.error('Subscription error:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to update subscription';
      toast.error(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <Button
        variant="outline"
        size="sm"
        onClick={() => toast.error('Please log in to subscribe to subforums')}
        className="gap-2"
      >
        <Bell className="w-4 h-4" />
        Subscribe
      </Button>
    );
  }

  return (
    <Button
      variant={isSubscribed ? "default" : "outline"}
      size="sm"
      onClick={handleSubscribe}
      disabled={isLoading}
      className="gap-2"
    >
      {isLoading ? (
        <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
      ) : isSubscribed ? (
        <BellOff className="w-4 h-4" />
      ) : (
        <Bell className="w-4 h-4" />
      )}
      {isSubscribed ? 'Unsubscribe' : 'Subscribe'}
    </Button>
  );
} 