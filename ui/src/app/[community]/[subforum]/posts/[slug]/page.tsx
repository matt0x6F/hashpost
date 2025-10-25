'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { PostsApi } from '@/generated/api/src/apis/PostsApi';
import { VotingApi } from '@/generated/api/src/apis/VotingApi';
import { VoteOnPostRequestVoteTypeEnum } from '@/generated/api/src/models';
import { toast } from 'sonner';
import { usePostData } from '@/hooks/usePostData';
import { CommentsApi } from '@/generated/api/src/apis/CommentsApi';
import { UserDisplay } from '@/components/UserDisplay';
import { 
  MessageSquare,
  Calendar,
  User,
  MoreHorizontal,
  Lock,
  Unlock,
  Pin,
  PinOff,
  Trash2,
  RotateCcw,
  Flag
} from 'lucide-react';
import CommentForm from '@/components/CommentForm';
import Comment from '@/components/Comment';
import VoteButtons from '@/components/VoteButtons';
import { Comment as CommentType } from '@/generated/api/src/models';
import { useAuth } from '@/lib/auth-context';
import { MarkdownRenderer } from '@/components/MarkdownRenderer';
import { PostBadges } from '@/components/PostBadges';
import { ReportDialog } from '@/components/ReportDialog';

export default function PostPage() {
  const params = useParams();
  const communityType = params.community as string;
  const subforum = params.subforum as string;
  const slug = params.slug as string;
  const fullSubforumPath = `${communityType}/${subforum}`;
  const { user, login } = useAuth();
  const [showDropdown, setShowDropdown] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const [comments, setComments] = useState<any[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(false);
  
  // Use the proper data fetching hooks
  const { post, metrics, userVote, moderation, isLoading, error, refetch } = usePostData(slug);

  // Fetch comments for the post
  const fetchComments = async () => {
    if (!post?.id) return;
    
    setCommentsLoading(true);
    try {
      const commentsApi = getApi(CommentsApi);
      const response = await commentsApi.listComments(post.id);
      setComments(response.comments || []);
    } catch (err) {
      console.error('Error fetching comments:', err);
      toast.error('Failed to load comments');
    } finally {
      setCommentsLoading(false);
    }
  };

  useEffect(() => {
    if (fullSubforumPath && slug) {
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, slug]);

  // Fetch comments when post is loaded
  useEffect(() => {
    if (post?.id) {
      fetchComments();
    }
  }, [post?.id]);

  const loadSubforumUserContext = async () => {
    try {
      // In atproto system, capabilities are handled globally via RBAC
      // No need for subforum-specific authentication
      console.log('Subforum context loading not needed in atproto system');
    } catch (error) {
      console.error('Error loading subforum user context:', error);
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

  const handlePostVote = async (voteValue: number) => {
    if (!user) {
      toast.error('Please log in to vote');
      return;
    }

    try {
      const votingApi = getApi(VotingApi);
      if (voteValue === 0) {
        // Remove vote
        await votingApi.removeVoteFromPost(slug);
      } else {
        // Add vote
        await votingApi.voteOnPost(slug, { 
          voteType: voteValue === 1 ? VoteOnPostRequestVoteTypeEnum.UP : VoteOnPostRequestVoteTypeEnum.DOWN
        });
      }
      
      // Note: UI updates optimistically, no need to refetch
      
    } catch (err: unknown) {
      console.error('Error voting on post:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on post';
      toast.error('Failed to vote', { description: errorMessage });
    }
  };

  const handleCommentVote = async (commentId: string, voteValue: number) => {
    if (!user) {
      toast.error('Please log in to vote');
      return;
    }

    try {
      const votingApi = getApi(VotingApi);
      if (voteValue === 0) {
        // Remove vote
        await votingApi.removeVoteFromComment(commentId);
      } else {
        // Add vote
        await votingApi.voteOnComment(commentId, { 
          voteType: voteValue === 1 ? VoteOnPostRequestVoteTypeEnum.UP : VoteOnPostRequestVoteTypeEnum.DOWN
        });
      }
      
      // Refresh comments to get updated vote counts
      await fetchComments();
      
    } catch (err: unknown) {
      console.error('Error voting on comment:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on comment';
      toast.error('Failed to vote', { description: errorMessage });
      
      // Refresh comments on error to revert optimistic updates
      await fetchComments();
    }
  };

  // In atproto system, user permissions are handled via RBAC
  const isPostAuthor = user?.did === post?.author?.did;
  const isModerator = false; // For now, assume no permissions
  
  const handleModeratorAction = async (action: string, value: boolean) => {
    if (!isModerator || !post) {
      toast.error('You do not have permission to perform this action');
      return;
    }
    try {
      // Moderation actions not available in atproto system
      toast.error("Moderation actions are not available in the atproto system");
    } catch {
      toast.error('Failed to update post');
    } finally {
      setShowDropdown(false);
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
        <div className="flex items-center gap-4 mb-8">
          <div className="h-8 w-8 bg-muted animate-pulse rounded" />
          <div className="h-8 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="space-y-4">
          <div className="h-8 bg-muted animate-pulse rounded" />
          <div className="h-4 bg-muted animate-pulse rounded" />
          <div className="h-4 bg-muted animate-pulse rounded w-2/3" />
        </div>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="max-w-7xl mx-auto p-2 sm:p-4">
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
            {error || 'Post not found'}
          </p>
          <Button 
            variant="outline" 
            onClick={refetch}
            className="mt-4"
          >
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-2 sm:p-4">
      {/* Navigation */}
      <div className="flex items-center gap-4 mb-8">
        <Link href={`/${communityType}/${subforum}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to {communityType}/{subforum}
          </Button>
        </Link>
      </div>

      {/* Post Header */}
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-2 justify-between">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{post.title}</h1>
            <PostBadges
              isSpoiler={false}
              isNsfw={false}
              isSticky={false}
              isLocked={moderation?.isLocked || false}
              isRemoved={moderation?.isRemoved || false}
            />
          </div>
          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDropdown(!showDropdown)}
              disabled={isLoading}
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>
            {showDropdown && (
              <div className="absolute right-0 top-full mt-1 bg-popover border border-border rounded-md shadow-lg z-50 min-w-48">
                <div className="p-2 space-y-1">
                  {/* Report Button - for all users except author */}
                  {!isPostAuthor && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-orange-600 hover:text-orange-700"
                      onClick={() => {
                        setShowReportDialog(true);
                        setShowDropdown(false);
                      }}
                      disabled={isLoading}
                    >
                      <Flag className="w-4 h-4 mr-2" />
                      Report Post
                    </Button>
                  )}
                  
                  {/* Author Actions */}
                  {isPostAuthor && (
                    <>
                      <Link href={`/${communityType}/${subforum}/posts/${slug}/edit`}>
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
                              // Post deletion not available in atproto system
                              toast.error("Post deletion is not available in the atproto system");
                              window.location.href = `/${communityType}/${subforum}`;
                            } catch (error: unknown) {
                              console.error('Error deleting post:', error);
                              const errorMessage = error instanceof Error ? error.message : 'Failed to delete post';
                              toast.error('Failed to delete post', { description: errorMessage });
                            } finally {
                              setShowDropdown(false);
                            }
                          }
                        }}
                        disabled={isLoading}
                      >
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                      </Button>
                    </>
                  )}
                  
                  {/* Moderator Actions */}
                  {isModerator && (
                    <>
                      {(isPostAuthor || !isPostAuthor) && <hr className="my-2 border-border" />}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={() => handleModeratorAction('lock', !(moderation?.isLocked ?? false))}
                        disabled={isLoading}
                      >
                        {moderation?.isLocked ? <Unlock className="w-4 h-4 mr-2" /> : <Lock className="w-4 h-4 mr-2" />}
                        {moderation?.isLocked ? 'Unlock Post' : 'Lock Post'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={() => handleModeratorAction('sticky', !(moderation?.isPinned ?? false))}
                        disabled={isLoading}
                      >
                        {moderation?.isPinned ? <PinOff className="w-4 h-4 mr-2" /> : <Pin className="w-4 h-4 mr-2" />}
                        {moderation?.isPinned ? 'Unsticky Post' : 'Sticky Post'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start text-destructive hover:text-destructive"
                        onClick={() => handleModeratorAction('remove', !(moderation?.isRemoved ?? false))}
                        disabled={isLoading}
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
        
        <div className="flex items-center gap-4 text-sm text-muted-foreground mb-4">
          <div className="flex items-center gap-1">
            <User className="w-4 h-4" />
            <span>by <UserDisplay author={post.author} /></span>
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="w-4 h-4" />
            <span>{formatDate(post.createdAt.toString())}</span>
          </div>
          <div className="flex items-center gap-1">
            <MessageSquare className="w-4 h-4" />
            <span>{metrics?.commentCount || 0} comments</span>
          </div>
        </div>
      </div>

      {/* Post Content */}
      <div className="bg-card border border-border rounded-lg p-6 mb-6">
        {post.content && (
          <div className="prose prose-sm max-w-none mb-4">
            <MarkdownRenderer content={post.content} />
          </div>
        )}
        
        {/* URL section removed - not implemented yet */}

        {/* Post Actions */}
        <div className="flex items-center justify-between mt-6 pt-4 border-t border-border">
          <VoteButtons
            score={metrics?.score || 0}
            userVote={userVote?.voteType === 'up' ? 1 : userVote?.voteType === 'down' ? -1 : 0}
            onVote={handlePostVote}
            disabled={isLoading}
            size="md"
          />
        </div>
      </div>

      {/* Comment Form */}
      {!moderation?.isLocked && (
        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-4">Add a Comment</h2>
          <CommentForm
            postId={post.id}
            onCommentSubmitted={async () => {
              await refetch();
              // Add a small delay to allow for eventual consistency
              toast.info('Comment posted! Refreshing comments...');
              setTimeout(async () => {
                await fetchComments();
              }, 1500);
            }}
          />
        </div>
      )}
      
      {moderation?.isLocked && (
        <div className="mb-6 p-4 bg-muted/20 border border-border rounded-lg">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Lock className="w-4 h-4" />
            <span className="text-sm">This post is locked. New comments are not allowed.</span>
          </div>
        </div>
      )}

      {/* Comments Section */}
      <div id="comments" className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Comments ({metrics?.commentCount || 0})</h2>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchComments}
            disabled={commentsLoading}
          >
            {commentsLoading ? 'Loading...' : 'Refresh'}
          </Button>
        </div>
        
        {commentsLoading ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground">Loading comments...</p>
          </div>
        ) : comments && comments.length > 0 ? (
          <div className="space-y-4">
            {comments.map((comment) => (
              <Comment
                key={comment.id}
                comment={comment}
                postId={post.id}
                onCommentUpdated={refetch}
                onCommentVoted={handleCommentVote}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-muted-foreground">No comments yet. Be the first to comment!</p>
          </div>
        )}
      </div>
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="post"
        contentId={parseInt(post.id) || 0}
        reportedPseudonymId={post.author?.did || ''}
        contentTitle={post.title}
        contentPreview={post.content}
        reportedUserDisplayName={post.author?.handle || ''}
      />
    </div>
  );
} 