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
import { UserDisplay } from '@/components/UserDisplay';
import { VotingApi, PostsApi } from '@/generated/api/src/apis';
import { VoteOnPostRequestVoteTypeEnum } from '@/generated/api/src/models';
import { toast } from 'sonner';
import { MarkdownRenderer } from './MarkdownRenderer';
import { PostBadges } from './PostBadges';
import { ReportDialog } from './ReportDialog';
import { usePostData } from '@/hooks/usePostData';
import { useCapabilities } from '@/lib/capabilities';

interface PostCardProps {
  postId: string;
  subforumName?: string;
}

export function PostCard({ postId, subforumName }: PostCardProps) {
  const { user, isAuthenticated } = useAuth();
  const [showModeratorControls, setShowModeratorControls] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const capabilities = useCapabilities();
  
  // Use the new data fetching hook
  const { post, metrics, userVote, moderation, isLoading: dataLoading, error, refetch } = usePostData(postId);

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

  const [isModerator, setIsModerator] = useState(false);
  
  // Check moderator status using capabilities
  useEffect(() => {
    if (isAuthenticated && user) {
      capabilities.canModerateContent().then(setIsModerator);
    }
  }, [isAuthenticated, user, capabilities]);

  const isAuthor = user?.did === post?.author;

  const handleVote = async (voteType: 'up' | 'down' | null) => {
    if (!isAuthenticated) {
      toast.error('Please log in to vote');
      return;
    }

    try {
      const votingApi = getApi(VotingApi);
      if (voteType === null) {
        // Remove vote
        await votingApi.removeVoteFromPost(postId);
      } else {
        // Add vote
        await votingApi.voteOnPost(postId, { 
          voteType: voteType === 'up' ? VoteOnPostRequestVoteTypeEnum.UP : VoteOnPostRequestVoteTypeEnum.DOWN
        });
      }
      
      // Refresh the data after successful vote
      await refetch();
      
    } catch (err: unknown) {
      console.error('Error voting on post:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on post';
      toast.error('Failed to vote', { description: errorMessage });
    }
  };

  const handleModeratorAction = async (action: string, value: boolean) => {
    if (!isAuthenticated || !isModerator) {
      toast.error('You do not have permission to perform this action');
      return;
    }

    try {
      // TODO: Implement moderator actions with new API
      // This would need to be implemented in the backend
      console.log('Moderator action:', action, value);
      toast.success(`Post ${action} ${value ? 'enabled' : 'disabled'}`);
      
      // Refresh moderation state
      await refetch();
    } catch (err: unknown) {
      console.error('Error performing moderator action:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to perform action';
      toast.error('Failed to perform action', { description: errorMessage });
    } finally {
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
  if (moderation?.isRemoved && !isModerator) {
    return null;
  }

  // Show loading state
  if (dataLoading) {
    return (
      <div className="border border-border rounded-lg p-4 mb-4 bg-card">
        <div className="animate-pulse">
          <div className="h-6 bg-muted rounded mb-2" />
          <div className="h-4 bg-muted rounded mb-4" />
          <div className="h-4 bg-muted rounded w-2/3" />
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className="border border-border rounded-lg p-4 mb-4 bg-card">
        <p className="text-red-500">{error}</p>
      </div>
    );
  }

  // Show not found state
  if (!post) {
    return (
      <div className="border border-border rounded-lg p-4 mb-4 bg-card">
        <p className="text-muted-foreground">Post not found</p>
      </div>
    );
  }

  // Remove edit button and dialog from PostCard
  // Only show post content once, as a summary/preview
  return (
    <div className={`border border-border rounded-lg p-4 mb-4 ${
      moderation?.isLocked ? 'opacity-60 bg-muted/20' : 'bg-card'
    }`}>
      {/* Post Header */}
      <div className="flex items-center gap-2 mb-1 justify-between">
        <div className="flex items-center gap-2">
          <Link href={`/${subforumName || `h/${post.subforum}`}/posts/${post.id}`} className="hover:underline">
            <h3 className="text-lg font-semibold">{post.title}</h3>
          </Link>
          <PostBadges
            isSticky={moderation?.isPinned || false}
            isLocked={moderation?.isLocked || false}
            isRemoved={moderation?.isRemoved || false}
          />
        </div>
        <div className="relative">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowModeratorControls(!showModeratorControls)}
            disabled={dataLoading}
          >
            <MoreHorizontal className="w-4 h-4" />
          </Button>
          {showModeratorControls && (
            <div className="absolute right-0 top-full mt-1 bg-popover border border-border rounded-md shadow-lg z-50 min-w-48 dropdown-menu">
              <div className="p-2 space-y-1">
                {isAuthor && (
                  <>
                    <Link href={`/${subforumName || `h/${post.subforum}`}/posts/${post.id}/edit`}>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                      >
                        Edit
                      </Button>
                    </Link>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-destructive hover:text-destructive"
                      onClick={async () => {
                        if (confirm('Are you sure you want to delete this post? This action cannot be undone.')) {
                          try {
                            const postsApi = getApi(PostsApi);
                            await postsApi.deletePost(post.id);
                            toast.success('Post deleted successfully');
                            // Refresh the page or redirect
                            window.location.reload();
                          } catch (error: unknown) {
                            console.error('Error deleting post:', error);
                            const errorMessage = error instanceof Error ? error.message : 'Failed to delete post';
                            toast.error('Failed to delete post', { description: errorMessage });
                          } finally {
                            setShowModeratorControls(false);
                          }
                        }
                      }}
                      disabled={dataLoading}
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      Delete
                    </Button>
                  </>
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
                    disabled={dataLoading}
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
                    disabled={dataLoading}
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
                      onClick={() => handleModeratorAction('lock', !moderation?.isLocked)}
                      disabled={dataLoading}
                    >
                      {moderation?.isLocked ? <Unlock className="w-4 h-4 mr-2" /> : <Lock className="w-4 h-4 mr-2" />}
                      {moderation?.isLocked ? 'Unlock Post' : 'Lock Post'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start"
                      onClick={() => handleModeratorAction('sticky', !moderation?.isPinned)}
                      disabled={dataLoading}
                    >
                      {moderation?.isPinned ? <PinOff className="w-4 h-4 mr-2" /> : <Pin className="w-4 h-4 mr-2" />}
                      {moderation?.isPinned ? 'Unsticky Post' : 'Sticky Post'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-destructive hover:text-destructive"
                      onClick={() => handleModeratorAction('remove', !moderation?.isRemoved)}
                      disabled={dataLoading}
                    >
                      {moderation?.isRemoved ? <RotateCcw className="w-4 h-4 mr-2" /> : <Trash2 className="w-4 h-4 mr-2" />}
                      {moderation?.isRemoved ? 'Restore Post' : 'Remove Post'}
                    </Button>
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
        <span>by <UserDisplay author={post.author} /></span>
        <span>•</span>
        <span>{formatDate(post.createdAt.toString())}</span>
        <span>•</span>
        <span>h/{post.subforum}</span>
      </div>
      {/* Post Content Preview */}
      {post.content && (
        <div className="mb-4">
          <div className="text-sm text-muted-foreground line-clamp-3">
            <MarkdownRenderer content={post.content} />
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
              onClick={() => handleVote(userVote?.voteType === 'up' ? null : 'up')}
              disabled={dataLoading}
            >
              <ArrowUp className={`w-4 h-4 ${userVote?.voteType === 'up' ? 'text-emerald-500' : 'text-muted-foreground'}`} />
            </Button>
            <span className="text-sm font-medium min-w-[2rem] text-center">
              {formatScore(metrics?.score || 0)}
            </span>
            <Button 
              variant="ghost" 
              size="sm"
              className="h-8 px-2"
              onClick={() => handleVote(userVote?.voteType === 'down' ? null : 'down')}
              disabled={dataLoading}
            >
              <ArrowDown className={`w-4 h-4 ${userVote?.voteType === 'down' ? 'text-emerald-500' : 'text-muted-foreground'}`} />
            </Button>
          </div>
          <Link href={`/${subforumName || `h/${post.subforum}`}/posts/${post.id}#comments`} className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors">
            <MessageSquare className="w-4 h-4" />
            <span>{metrics?.commentCount || 0} comments</span>
          </Link>
        </div>
        

      </div>
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="post"
        contentId={parseInt(post.id)}
        reportedPseudonymId={post.author}
        contentTitle={post.title}
        contentPreview={post.content}
        reportedUserDisplayName={post.author}
      />
    </div>
  );
} 