'use client';

import React, { useState, useEffect } from 'react';
import { Button } from './shadcn/button';
import Link from 'next/link';
import { 
  Lock, 
  Pin, 
  ArrowUp, 
  ArrowDown,
  MessageSquare,
  MoreHorizontal,
  Unlock,
  PinOff,
  Trash2,
  RotateCcw,
  Flag
} from 'lucide-react';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';
import { MarkdownRenderer } from './MarkdownRenderer';
import { PostBadges } from './PostBadges';
import { ReportDialog } from './ReportDialog';

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
  subforumName?: string; // Optional subforum name for subforum-specific authentication
}

export function PostCard({ post, subforumName }: PostCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [localPost, setLocalPost] = useState(post);
  const { user, isAuthenticated } = useAuth();
  const [showModeratorControls, setShowModeratorControls] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      // Don't close if clicking on the dropdown itself
      const target = event.target as Element;
      if (target.closest('.dropdown-menu')) {
        return;
      }
      
      if (showModeratorControls) {
        setShowModeratorControls(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showModeratorControls]);

  // Update local post when prop changes
  React.useEffect(() => {
    setLocalPost(post);
  }, [post]);

  // Remove the useEffect that calls authenticateUserForSubforum
  // This should be handled at a higher level to avoid multiple API calls

  const isModerator = user?.capabilities?.includes('moderate_content');

  const isAuthor = user?.activePseudonymId === localPost.author.pseudonymId;

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
      setLocalPost(prev => ({
        ...prev,
        isLocked: action === 'lock' ? value : prev.isLocked,
        isSticky: action === 'sticky' ? value : prev.isSticky,
        isRemoved: action === 'remove' ? value : prev.isRemoved,
      }));
      toast.success(`Post ${action === 'lock' ? (value ? 'locked' : 'unlocked') : action === 'sticky' ? (value ? 'stickied' : 'unstickied') : (value ? 'removed' : 'restored')}`);
    } catch (error: unknown) {
      console.error(`Error ${action}ing post:`, error);
      const errorMessage = error instanceof Error ? error.message : `Failed to ${action} post`;
      toast.error(`Failed to ${action} post`, { description: errorMessage });
    } finally {
      setIsLoading(false);
      setShowModeratorControls(false);
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

  // Debug: Check author match for edit button
  // If post is removed and user is not a moderator, don't show it
  if (localPost.isRemoved && !isModerator) {
    return null;
  }

  // Remove edit button and dialog from PostCard
  // Only show post content once, as a summary/preview
  return (
    <div className={`border border-border rounded-lg p-4 mb-4 ${
      localPost.isRemoved ? 'opacity-60 bg-muted/20' : 'bg-card'
    }`}>
      {/* Post Header */}
      <div className="flex items-center gap-2 mb-1 justify-between">
        <div className="flex items-center gap-2">
          <Link href={`/${subforumName || `h/${localPost.subforum.name}`}/posts/${localPost.slug}`} className="hover:underline">
            <h3 className="text-lg font-semibold">{localPost.title}</h3>
          </Link>
          <PostBadges
            isSticky={localPost.isSticky}
            isLocked={localPost.isLocked}
            isRemoved={localPost.isRemoved}
          />
        </div>
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
            <div className="absolute right-0 top-full mt-1 bg-popover border border-border rounded-md shadow-lg z-50 min-w-48 dropdown-menu">
              <div className="p-2 space-y-1">
                {isAuthor && (
                  <Link href={`/${subforumName || `h/${localPost.subforum.name}`}/posts/${localPost.slug}/edit`}>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start"
                    >
                      Edit
                    </Button>
                  </Link>
                )}
                {isAuthor && <hr className="my-2 border-border" />}
                
                {/* Report Button */}
                {isAuthenticated && !isAuthor && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start text-orange-600 hover:text-orange-700"
                    onClick={() => {
                      setShowReportDialog(true);
                      setShowModeratorControls(false);
                    }}
                    disabled={isLoading}
                  >
                    <Flag className="w-4 h-4 mr-2" />
                    Report Post
                  </Button>
                )}
                
                {/* Report button for non-authenticated users */}
                {!isAuthenticated && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start text-muted-foreground"
                    onClick={() => {
                      toast.error('Please log in to report posts');
                      setShowModeratorControls(false);
                    }}
                    disabled={isLoading}
                  >
                    <Flag className="w-4 h-4 mr-2" />
                    Report Post (Login Required)
                  </Button>
                )}
                
                {/* Moderator Actions */}
                {isModerator && (
                  <>
                    <hr className="my-2 border-border" />
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
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
        <span>by {localPost.author.displayName}</span>
        <span>•</span>
        <span>{formatDate(localPost.createdAt)}</span>
        <span>•</span>
        <span>h/{localPost.subforum.name}</span>
      </div>
      {/* Post Content Preview */}
      {localPost.content && (
        <div className="mb-4">
          <div className="text-sm text-muted-foreground line-clamp-3">
            <MarkdownRenderer content={localPost.content} />
          </div>
        </div>
      )}
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
          <Link href={`/${subforumName || `h/${localPost.subforum.name}`}/posts/${localPost.slug}#comments`} className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors">
            <MessageSquare className="w-4 h-4" />
            <span>{localPost.commentCount} comments</span>
          </Link>
        </div>
        

      </div>
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="post"
        contentId={localPost.postId}
        reportedPseudonymId={localPost.author.pseudonymId}
        contentTitle={localPost.title}
        contentPreview={localPost.content}
        reportedUserDisplayName={localPost.author.displayName}
      />
    </div>
  );
} 