'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { ContentApi, ModerationApi } from '@/generated/api/src';
import { PostDetailsResponseBody } from '@/generated/api/src/models';
import { toast } from 'sonner';
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
import { authenticateUserForSubforum } from '@/lib/auth-utils';

export default function PostPage() {
  const params = useParams();
  const communityType = params.community as string;
  const subforum = params.subforum as string;
  const slug = params.slug as string;
  const fullSubforumPath = `${communityType}/${subforum}`;
  const [postDetails, setPostDetails] = useState<PostDetailsResponseBody | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isVoting, setIsVoting] = useState(false);
  const { user, login } = useAuth();
  const [showDropdown, setShowDropdown] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);

  useEffect(() => {
    if (fullSubforumPath && slug) {
      loadPostDetails();
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, slug]);

  const loadSubforumUserContext = async () => {
    try {
      const userData = await authenticateUserForSubforum(fullSubforumPath);
      if (userData) {
        login(userData);
      }
    } catch (error) {
      console.error('Error loading subforum user context:', error);
    }
  };

  const loadPostDetails = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.getPostBySlug(fullSubforumPath, slug, 'best');
      setPostDetails(response);
    } catch (err: unknown) {
      console.error('Error loading post details:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load post';
      setError(errorMessage);
      
      toast.error('Failed to load post', {
        description: errorMessage,
      });
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

  const handlePostVote = async (voteValue: number) => {
    if (!postDetails) return;
    setIsVoting(true);
    
    const previousPostDetails = { ...postDetails };
    
    setPostDetails(prev => {
      if (!prev) return prev;
      
      let newScore = prev.score;
      let newUpvotes = prev.upvotes;
      let newDownvotes = prev.downvotes;
      
      if (prev.userVote === 1) {
        newScore -= 1;
        newUpvotes -= 1;
      } else if (prev.userVote === -1) {
        newScore += 1;
        newDownvotes -= 1;
      }
      
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
      await contentApi.voteOnPost(postDetails.postId, { voteValue });
    } catch (err: unknown) {
      console.error('Error voting on post:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on post';
      toast.error('Failed to vote', { description: errorMessage });
      setPostDetails(previousPostDetails);
    } finally {
      setIsVoting(false);
    }
  };

  const handleCommentVote = async (commentId: number, voteValue: number) => {
    if (!postDetails) return;
    
    const previousPostDetails = { ...postDetails };
    
    setPostDetails(prev => {
      if (!prev) return prev;
      
      const updateCommentVote = (comments: CommentType[]): CommentType[] => {
        return comments.map(comment => {
          if (comment.commentId === commentId) {
            let newScore = comment.score;
            
            if (comment.userVote === 1) {
              newScore -= 1;
            } else if (comment.userVote === -1) {
              newScore += 1;
            }
            
            if (voteValue === 1) {
              newScore += 1;
            } else if (voteValue === -1) {
              newScore -= 1;
            }
            
            return {
              ...comment,
              score: newScore,
              userVote: voteValue
            };
          }
          
          if (comment.replies && comment.replies.length > 0) {
            return {
              ...comment,
              replies: updateCommentVote(comment.replies)
            };
          }
          
          return comment;
        });
      };
      
      return {
        ...prev,
        comments: updateCommentVote(prev.comments || [])
      };
    });
    
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.voteOnComment(commentId, { voteValue });
    } catch (err: unknown) {
      console.error('Error voting on comment:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on comment';
      toast.error('Failed to vote', { description: errorMessage });
      setPostDetails(previousPostDetails);
    }
  };

  const isPostAuthor = user?.activePseudonymId && postDetails?.author?.pseudonymId && (user.activePseudonymId === postDetails.author.pseudonymId);
  const isModerator = user?.capabilities?.includes('moderate_content');
  
  const handleModeratorAction = async (action: string, value: boolean) => {
    if (!isModerator || !postDetails) {
      toast.error('You do not have permission to perform this action');
      return;
    }
    setIsLoading(true);
    try {
      const moderationApi = getApi(ModerationApi);
      switch (action) {
        case 'lock':
          await moderationApi.lockPost(postDetails.postId, { locked: value });
          break;
        case 'sticky':
          await moderationApi.stickyPost(postDetails.postId, { sticky: value });
          break;
        case 'remove':
          await moderationApi.removePost(postDetails.postId, { removed: value });
          break;
      }
      setPostDetails(prev => prev ? {
        ...prev,
        isLocked: action === 'lock' ? value : (prev.isLocked ?? false),
        isSticky: action === 'sticky' ? value : (prev.isSticky ?? false),
        isRemoved: action === 'remove' ? value : (prev.isRemoved ?? false),
      } : prev);
      toast.success(`Post ${action === 'lock' ? (value ? 'locked' : 'unlocked') : action === 'sticky' ? (value ? 'stickied' : 'unstickied') : (value ? 'removed' : 'restored')}`);
    } catch {
      toast.error('Failed to update post');
    } finally {
      setIsLoading(false);
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

  if (error || !postDetails) {
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
            onClick={loadPostDetails}
            className="mt-4"
          >
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  const { comments } = postDetails;

  return (
    <div className="max-w-7xl mx-auto p-2 sm:p-4">
      {/* Navigation */}
      <div className="flex items-center gap-4 mb-8">
        <Link href={`/${communityType}/${postDetails.subforum.name}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to {communityType}/{postDetails.subforum.name}
          </Button>
        </Link>
      </div>

      {/* Post Header */}
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-2 justify-between">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{postDetails.title}</h1>
            <PostBadges
              isSpoiler={postDetails.isSpoiler}
              isNsfw={postDetails.isNsfw}
              isSticky={postDetails.isSticky}
              isLocked={postDetails.isLocked}
              isRemoved={postDetails.isRemoved}
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
                      <Link href={`/${communityType}/${postDetails.subforum.name}/posts/${postDetails.slug}/edit`}>
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
                            setIsLoading(true);
                            try {
                              const contentApi = getApi(ContentApi);
                              await contentApi.deletePost(postDetails.postId, { reason: 'User requested deletion' });
                              toast.success('Post deleted');
                              window.location.href = `/${communityType}/${postDetails.subforum.name}`;
                            } catch (error: unknown) {
                              console.error('Error deleting post:', error);
                              const errorMessage = error instanceof Error ? error.message : 'Failed to delete post';
                              toast.error('Failed to delete post', { description: errorMessage });
                            } finally {
                              setIsLoading(false);
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
                        onClick={() => handleModeratorAction('lock', !(postDetails.isLocked ?? false))}
                        disabled={isLoading}
                      >
                        {postDetails.isLocked ? <Unlock className="w-4 h-4 mr-2" /> : <Lock className="w-4 h-4 mr-2" />}
                        {postDetails.isLocked ? 'Unlock Post' : 'Lock Post'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={() => handleModeratorAction('sticky', !(postDetails.isSticky ?? false))}
                        disabled={isLoading}
                      >
                        {postDetails.isSticky ? <PinOff className="w-4 h-4 mr-2" /> : <Pin className="w-4 h-4 mr-2" />}
                        {postDetails.isSticky ? 'Unsticky Post' : 'Sticky Post'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start text-destructive hover:text-destructive"
                        onClick={() => handleModeratorAction('remove', !(postDetails.isRemoved ?? false))}
                        disabled={isLoading}
                      >
                        {postDetails.isRemoved ? <RotateCcw className="w-4 h-4 mr-2" /> : <Trash2 className="w-4 h-4 mr-2" />}
                        {postDetails.isRemoved ? 'Restore Post' : 'Remove Post'}
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
            <span>by {postDetails.author.displayName}</span>
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="w-4 h-4" />
            <span>{formatDate(postDetails.createdAt)}</span>
          </div>
          <div className="flex items-center gap-1">
            <MessageSquare className="w-4 h-4" />
            <span>{postDetails.commentCount} comments</span>
          </div>
        </div>
      </div>

      {/* Post Content */}
      <div className="bg-card border border-border rounded-lg p-6 mb-6">
        {postDetails.content && (
          <div className="prose prose-sm max-w-none mb-4">
            <MarkdownRenderer content={postDetails.content} />
          </div>
        )}
        
        {postDetails.url && !postDetails.isSelfPost && (
          <div className="mt-4">
            <a 
              href={postDetails.url} 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-primary hover:underline"
            >
              {postDetails.url}
            </a>
          </div>
        )}

        {/* Post Actions */}
        <div className="flex items-center justify-between mt-6 pt-4 border-t border-border">
          <VoteButtons
            score={postDetails.score}
            userVote={postDetails.userVote}
            onVote={handlePostVote}
            disabled={isVoting}
            size="md"
          />
        </div>
      </div>

      {/* Comment Form */}
      {!postDetails.isLocked && (
        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-4">Add a Comment</h2>
          <CommentForm
            postId={postDetails.postId}
            onCommentSubmitted={loadPostDetails}
          />
        </div>
      )}
      
      {postDetails.isLocked && (
        <div className="mb-6 p-4 bg-muted/20 border border-border rounded-lg">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Lock className="w-4 h-4" />
            <span className="text-sm">This post is locked. New comments are not allowed.</span>
          </div>
        </div>
      )}

      {/* Comments Section */}
      <div id="comments" className="space-y-4">
        <h2 className="text-xl font-semibold">Comments ({postDetails.commentCount || 0})</h2>
        
        {!comments || comments.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground">No comments yet. Be the first to comment!</p>
          </div>
        ) : (
          <div className="space-y-4">
            {comments.map((comment) => (
              <Comment
                key={comment.commentId}
                comment={comment}
                postId={postDetails.postId}
                onCommentUpdated={loadPostDetails}
                onCommentVoted={handleCommentVote}
              />
            ))}
          </div>
        )}
      </div>
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="post"
        contentId={postDetails.postId}
        reportedPseudonymId={postDetails.author.pseudonymId}
        contentTitle={postDetails.title}
        contentPreview={postDetails.content}
        reportedUserDisplayName={postDetails.author.displayName}
      />
    </div>
  );
} 