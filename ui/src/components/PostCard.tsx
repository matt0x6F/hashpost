'use client';

import React, { useState } from 'react';
import { Button } from './shadcn/button';
import { Badge } from './shadcn/badge';
import Link from 'next/link';
import { 
  Lock, 
  Unlock, 
  Pin, 
  PinOff, 
  Trash2, 
  RotateCcw, 
  MessageSquare, 
  ArrowUp, 
  ArrowDown,
  MoreHorizontal,
  EyeOff
} from 'lucide-react';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';

interface PostCardProps {
  post: {
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
  };
  onPostUpdated?: (postId: number) => void;
}

export function PostCard({ post, onPostUpdated }: PostCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showModeratorControls, setShowModeratorControls] = useState(false);
  const [localPost, setLocalPost] = useState(post);
  const { user, isAuthenticated } = useAuth();

  // Update local post when prop changes
  React.useEffect(() => {
    setLocalPost(post);
  }, [post]);

  const isModerator = user?.roles?.includes('moderator') || 
                     user?.roles?.includes('subforum_owner') || 
                     user?.roles?.includes('platform_admin') || 
                     user?.roles?.includes('trust_safety') || 
                     user?.roles?.includes('legal_team') ||
                     user?.capabilities?.includes('moderate_content');

  // Debug logging
  console.log('[PostCard] User roles:', user?.roles);
  console.log('[PostCard] User capabilities:', user?.capabilities);
  console.log('[PostCard] Is moderator:', isModerator);

  const handleModeratorAction = async (action: string, value: boolean) => {
    if (!isAuthenticated || !isModerator) {
      toast.error('You do not have permission to perform this action');
      return;
    }

    setIsLoading(true);
    try {
      const contentApi = getApi(ContentApi);
      
      switch (action) {
        case 'lock':
          await contentApi.lockPost(localPost.postId, { locked: value });
          break;
        case 'sticky':
          await contentApi.stickyPost(localPost.postId, { sticky: value });
          break;
        case 'remove':
          await contentApi.removePost(localPost.postId, { removed: value });
          break;
      }
      
      // Update local state optimistically
      setLocalPost(prev => ({
        ...prev,
        isLocked: action === 'lock' ? value : prev.isLocked,
        isSticky: action === 'sticky' ? value : prev.isSticky,
        isRemoved: action === 'remove' ? value : prev.isRemoved,
      }));
      
      onPostUpdated?.(localPost.postId);
      toast.success(`Post ${action === 'lock' ? (value ? 'locked' : 'unlocked') : action === 'sticky' ? (value ? 'stickied' : 'unstickied') : (value ? 'removed' : 'restored')}`);
    } catch (error: unknown) {
      console.error(`Error ${action}ing post:`, error);
      const errorMessage = error instanceof Error ? error.message : `Failed to ${action} post`;
      toast.error(`Failed to ${action} post`, {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
      setShowModeratorControls(false);
    }
  };

  const handleVote = async (voteValue: number) => {
    if (!isAuthenticated) {
      toast.error('Please log in to vote');
      return;
    }

    setIsLoading(true);
    
    // Store the previous state for rollback on error
    const previousPost = { ...localPost };
    
    // Optimistically update the local state
    setLocalPost(prev => {
      // Calculate the new score based on the vote change
      let newScore = prev.score;
      let newUpvotes = prev.upvotes;
      let newDownvotes = prev.downvotes;
      
      // Remove the previous vote effect
      if (prev.userVote === 1) {
        newScore -= 1;
        newUpvotes -= 1;
      } else if (prev.userVote === -1) {
        newScore += 1;
        newDownvotes -= 1;
      }
      
      // Add the new vote effect
      if (voteValue === 1) {
        newScore += 1;
        newUpvotes += 1;
      } else if (voteValue === -1) {
        newScore -= 1;
        newDownvotes += 1;
      }
      
      return {
        ...prev,
        score: newScore,
        upvotes: newUpvotes,
        downvotes: newDownvotes,
        userVote: voteValue
      };
    });
    
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.voteOnPost(localPost.postId, { voteValue });
      // Don't reload the entire post list - the optimistic update is sufficient
    } catch (err: unknown) {
      console.error('Error voting on post:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on post';
      toast.error('Failed to vote', { description: errorMessage });
      
      // Rollback to the previous state on error
      setLocalPost(previousPost);
    } finally {
      setIsLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatScore = (score: number) => {
    if (score >= 1000) {
      return `${(score / 1000).toFixed(1)}k`;
    }
    return score.toString();
  };

  // If post is removed and user is not a moderator, don't show it
  if (localPost.isRemoved && !isModerator) {
    return null;
  }

  return (
    <div className={`border border-border rounded-lg p-4 mb-4 ${
      localPost.isRemoved ? 'opacity-60 bg-muted/20' : 'bg-card'
    }`}>
      {/* Post Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <Link href={`/h/${localPost.subforum.name}/posts/${localPost.slug}`} className="hover:underline">
              <h3 className="text-lg font-semibold">{localPost.title}</h3>
            </Link>
            {localPost.isSticky && (
              <Badge variant="secondary" className="text-xs">
                <Pin className="w-3 h-3 mr-1" />
                Sticky
              </Badge>
            )}
            {localPost.isLocked && (
              <Badge variant="destructive" className="text-xs">
                <Lock className="w-3 h-3 mr-1" />
                Locked
              </Badge>
            )}
            {localPost.isRemoved && (
              <Badge variant="destructive" className="text-xs">
                <EyeOff className="w-3 h-3 mr-1" />
                Removed
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>by {localPost.author.displayName}</span>
            <span>•</span>
            <span>{formatDate(localPost.createdAt)}</span>
            <span>•</span>
            <span>h/{localPost.subforum.name}</span>
          </div>
        </div>
        
        {/* Moderator Controls */}
        {isModerator && (
          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowModeratorControls(!showModeratorControls)}
              disabled={isLoading}
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>
            
            {showModeratorControls && (
              <div className="absolute right-0 top-full mt-1 bg-popover border border-border rounded-md shadow-lg z-10 min-w-48">
                <div className="p-2 space-y-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    onClick={() => handleModeratorAction('lock', !localPost.isLocked)}
                    disabled={isLoading}
                  >
                    {localPost.isLocked ? <Unlock className="w-4 h-4 mr-2" /> : <Lock className="w-4 h-4 mr-2" />}
                    {localPost.isLocked ? 'Unlock Post' : 'Lock Post'}
                  </Button>
                  
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    onClick={() => handleModeratorAction('sticky', !localPost.isSticky)}
                    disabled={isLoading}
                  >
                    {localPost.isSticky ? <PinOff className="w-4 h-4 mr-2" /> : <Pin className="w-4 h-4 mr-2" />}
                    {localPost.isSticky ? 'Unsticky Post' : 'Sticky Post'}
                  </Button>
                  
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start text-destructive hover:text-destructive"
                    onClick={() => handleModeratorAction('remove', !localPost.isRemoved)}
                    disabled={isLoading}
                  >
                    {localPost.isRemoved ? <RotateCcw className="w-4 h-4 mr-2" /> : <Trash2 className="w-4 h-4 mr-2" />}
                    {localPost.isRemoved ? 'Restore Post' : 'Remove Post'}
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Post Content */}
      <div className="mb-4">
        <p className="text-sm text-muted-foreground line-clamp-3">
          {localPost.content}
        </p>
      </div>

      {/* Post Footer */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1">
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-8 px-2"
              onClick={() => handleVote(localPost.userVote === 1 ? 0 : 1)}
              disabled={isLoading}
            >
              <ArrowUp className={`w-4 h-4 ${localPost.userVote === 1 ? 'text-emerald-500' : 'text-muted-foreground'}`} />
            </Button>
            <span className="text-sm font-medium min-w-[2rem] text-center">
              {formatScore(localPost.score)}
            </span>
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-8 px-2"
              onClick={() => handleVote(localPost.userVote === -1 ? 0 : -1)}
              disabled={isLoading}
            >
              <ArrowDown className={`w-4 h-4 ${localPost.userVote === -1 ? 'text-emerald-500' : 'text-muted-foreground'}`} />
            </Button>
          </div>
          
          <Link href={`/h/${localPost.subforum.name}/posts/${localPost.slug}#comments`} className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors">
            <MessageSquare className="w-4 h-4" />
            <span>{localPost.commentCount} comments</span>
          </Link>
        </div>
      </div>
    </div>
  );
} 