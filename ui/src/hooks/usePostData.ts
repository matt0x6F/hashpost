import { useState, useEffect, useCallback } from 'react';
import { getApi } from '@/lib/api-client';
import { PostsApi } from '@/generated/api/src/apis/PostsApi';
import type { Post, PostMetrics, UserVote, ModerationState } from '@/generated/api/src/models';

export interface PostData {
  post: Post | null;
  metrics: PostMetrics | null;
  userVote: UserVote | null;
  moderation: ModerationState | null;
  isLoading: boolean;
  error: string | null;
}

export function usePostData(postId: string): PostData & { refetch: () => Promise<void> } {
  const [post, setPost] = useState<Post | null>(null);
  const [metrics, setMetrics] = useState<PostMetrics | null>(null);
  const [userVote, setUserVote] = useState<UserVote | null>(null);
  const [moderation, setModeration] = useState<ModerationState | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchPostData = useCallback(async () => {
    if (!postId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const postsApi = getApi(PostsApi);
      
      // Fetch all data in parallel
      const [postRes, metricsRes, userVoteRes, moderationRes] = await Promise.allSettled([
        postsApi.getPostByID(postId),
        postsApi.getPostMetrics(postId),
        postsApi.getPostUserVote(postId),
        postsApi.getPostModerationState(postId)
      ]);
      
      // Handle post data
      if (postRes.status === 'fulfilled') {
        setPost(postRes.value);
      } else {
        console.error('Failed to fetch post:', postRes.reason);
        setError('Failed to load post');
      }
      
      // Handle metrics data
      if (metricsRes.status === 'fulfilled') {
        setMetrics(metricsRes.value);
      } else {
        console.error('Failed to fetch metrics:', metricsRes.reason);
        // Don't set error for metrics - it's not critical
      }
      
      // Handle user vote data
      if (userVoteRes.status === 'fulfilled') {
        setUserVote(userVoteRes.value);
      } else {
        console.error('Failed to fetch user vote:', userVoteRes.reason);
        // Don't set error for user vote - user might not be logged in
      }
      
      // Handle moderation data
      if (moderationRes.status === 'fulfilled') {
        setModeration(moderationRes.value);
      } else {
        console.error('Failed to fetch moderation state:', moderationRes.reason);
        // Don't set error for moderation - it's not critical
      }
      
    } catch (err) {
      console.error('Error fetching post data:', err);
      setError('Failed to load post data');
    } finally {
      setIsLoading(false);
    }
  }, [postId]);

  useEffect(() => {
    fetchPostData();
  }, [fetchPostData]);

  return {
    post,
    metrics,
    userVote,
    moderation,
    isLoading,
    error,
    refetch: fetchPostData
  };
}

// Hook for fetching just metrics (useful for vote updates)
export function usePostMetrics(postId: string) {
  const [metrics, setMetrics] = useState<PostMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = useCallback(async () => {
    if (!postId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const postsApi = getApi(PostsApi);
      const metricsData = await postsApi.getPostMetrics(postId);
      setMetrics(metricsData);
    } catch (err) {
      console.error('Error fetching post metrics:', err);
      setError('Failed to load metrics');
    } finally {
      setIsLoading(false);
    }
  }, [postId]);

  return {
    metrics,
    isLoading,
    error,
    refetch: fetchMetrics
  };
}

// Hook for fetching just user vote (useful for vote updates)
export function useUserVote(postId: string) {
  const [userVote, setUserVote] = useState<UserVote | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchUserVote = useCallback(async () => {
    if (!postId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const postsApi = getApi(PostsApi);
      const voteData = await postsApi.getPostUserVote(postId);
      setUserVote(voteData);
    } catch (err) {
      console.error('Error fetching user vote:', err);
      setError('Failed to load user vote');
    } finally {
      setIsLoading(false);
    }
  }, [postId]);

  return {
    userVote,
    isLoading,
    error,
    refetch: fetchUserVote
  };
}
