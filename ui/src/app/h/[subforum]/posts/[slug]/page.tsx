'use client';

import { useParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { PostDetailsResponseBody } from '@/generated/api/src/models';
import { toast } from 'sonner';
import { Badge } from '@/components/shadcn/badge';
import { 
  MessageSquare,
  Calendar,
  User
} from 'lucide-react';
import CommentForm from '@/components/CommentForm';
import Comment from '@/components/Comment';
import VoteButtons from '@/components/VoteButtons';
import { Comment as CommentType } from '@/generated/api/src/models';

export default function PostPage() {
  const params = useParams();
  const subforum = params.subforum as string;
  const slug = params.slug as string;
  const [postDetails, setPostDetails] = useState<PostDetailsResponseBody | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Add post voting handler
  const [isVoting, setIsVoting] = useState(false);

  useEffect(() => {
    if (subforum && slug) {
      loadPostDetails();
    }
  }, [subforum, slug]);

  const loadPostDetails = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.getPostBySlug(subforum, slug, 'best');
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

  // Add post voting handler
  const handlePostVote = async (voteValue: number) => {
    if (!postDetails) return;
    setIsVoting(true);
    
    // Store the previous state for rollback on error
    const previousPostDetails = { ...postDetails };
    
    // Optimistically update the local state
    setPostDetails(prev => {
      if (!prev) return prev;
      
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
      await contentApi.voteOnPost(postDetails.postId, { voteValue });
      // Don't reload the entire post - the optimistic update is sufficient
    } catch (err: unknown) {
      console.error('Error voting on post:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on post';
      toast.error('Failed to vote', { description: errorMessage });
      
      // Rollback to the previous state on error
      setPostDetails(previousPostDetails);
    } finally {
      setIsVoting(false);
    }
  };

  // Add comment voting handler
  const handleCommentVote = async (commentId: number, voteValue: number) => {
    if (!postDetails) return;
    
    // Store the previous state for rollback on error
    const previousPostDetails = { ...postDetails };
    
    // Optimistically update the comment state
    setPostDetails(prev => {
      if (!prev) return prev;
      
      const updateCommentVote = (comments: CommentType[]): CommentType[] => {
        return comments.map(comment => {
          if (comment.commentId === commentId) {
            // Calculate the new score based on the vote change
            let newScore = comment.score;
            
            // Remove the previous vote effect
            if (comment.userVote === 1) {
              newScore -= 1;
            } else if (comment.userVote === -1) {
              newScore += 1;
            }
            
            // Add the new vote effect
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
          
          // Recursively update replies
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
      // Don't reload the entire post - the optimistic update is sufficient
    } catch (err: unknown) {
      console.error('Error voting on comment:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to vote on comment';
      toast.error('Failed to vote', { description: errorMessage });
      
      // Rollback to the previous state on error
      setPostDetails(previousPostDetails);
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto p-6">
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
    <div className="max-w-4xl mx-auto p-6">
      {/* Navigation */}
      <div className="flex items-center gap-4 mb-8">
        <Link href={`/forums/${postDetails.subforum.name}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to h/{postDetails.subforum.name}
          </Button>
        </Link>
      </div>

      {/* Post Header */}
      <div className="mb-6">
        <div className="flex items-center gap-2 mb-2">
          <h1 className="text-2xl font-bold">{postDetails.title}</h1>
          {postDetails.isSpoiler && (
            <Badge variant="secondary" className="text-xs">
              Spoiler
            </Badge>
          )}
          {postDetails.isNsfw && (
            <Badge variant="destructive" className="text-xs">
              NSFW
            </Badge>
          )}
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
            <p className="text-foreground whitespace-pre-wrap">{postDetails.content}</p>
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
      <div className="mb-6">
        <h2 className="text-xl font-semibold mb-4">Add a Comment</h2>
        <CommentForm
          postId={postDetails.postId}
          onCommentSubmitted={loadPostDetails}
        />
      </div>

      {/* Comments Section */}
      <div id="comments" className="space-y-4">
        <h2 className="text-xl font-semibold">Comments ({comments?.length || 0})</h2>
        
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
    </div>
  );
} 