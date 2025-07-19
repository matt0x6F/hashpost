'use client';

import React, { useState, useEffect } from 'react';
import { Button } from './shadcn/button';
import { PostCard } from './PostCard';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';
import { Plus, Loader2 } from 'lucide-react';
import Link from 'next/link';

interface Post {
  postId: number;
  slug: string;
  title: string;
  content: string;
  author: {
    displayName: string;
    pseudonymId: string;
  };
  createdAt: string;
  score: number;
  upvotes: number;
  downvotes: number;
  commentCount: number;
  isLocked?: boolean;
  isSticky?: boolean;
  isRemoved?: boolean;
  subforum: {
    name: string;
  };
  userVote?: number; // Add userVote field
}

interface PostListProps {
  subforumName: string;
}

export function PostList({ subforumName }: PostListProps) {
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    loadPosts();
  }, [subforumName, page]);

  const loadPosts = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.getSubforumPosts(subforumName, undefined, undefined, page, 20, 'new', 'all');
      
      if (response.posts && Array.isArray(response.posts)) {
        const newPosts = response.posts;
        if (page === 1) {
          setPosts(newPosts);
        } else {
          setPosts(prev => [...prev, ...newPosts]);
        }
        
        // Check if we have more posts
        setHasMore(newPosts.length === 20);
      }
    } catch (err: unknown) {
      console.error('Error loading posts:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load posts';
      setError(errorMessage);
      
      toast.error('Failed to load posts', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const loadMore = () => {
    setPage(prev => prev + 1);
  };

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={loadPosts} variant="outline">
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Create Post Button */}
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Posts</h2>
        <Link href={`/h/${subforumName}/posts/new`}>
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Post
          </Button>
        </Link>
      </div>

      {/* Posts List */}
      <div className="space-y-4">
        {posts.length === 0 && !isLoading ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">
              No posts yet. Be the first to create a post in this forum!
            </p>
          </div>
        ) : (
          posts.map((post) => (
            <PostCard
              key={post.postId}
              post={post}
            />
          ))
        )}
      </div>

      {/* Load More Button */}
      {hasMore && posts.length > 0 && (
        <div className="text-center">
          <Button
            onClick={loadMore}
            variant="outline"
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Loading...
              </>
            ) : (
              'Load More Posts'
            )}
          </Button>
        </div>
      )}

      {/* Loading State */}
      {isLoading && posts.length === 0 && (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="border border-border rounded-lg p-4">
              <div className="h-6 bg-muted animate-pulse rounded mb-2" />
              <div className="h-4 bg-muted animate-pulse rounded mb-4" />
              <div className="h-4 bg-muted animate-pulse rounded w-2/3" />
            </div>
          ))}
        </div>
      )}
    </div>
  );
} 