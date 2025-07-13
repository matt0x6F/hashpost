'use client';

import { useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';
import { 
  Reply,
  MoreHorizontal
} from 'lucide-react';
import CommentForm from './CommentForm';
import type { Comment } from '@/generated/api/src/models';
import VoteButtons from './VoteButtons';

interface CommentProps {
  comment: Comment;
  postId: number;
  onCommentUpdated: () => void;
  onCommentVoted?: (commentId: number, voteValue: number) => void;
  depth?: number;
}

export default function Comment({ 
  comment, 
  postId, 
  onCommentUpdated, 
  onCommentVoted,
  depth = 0 
}: CommentProps) {
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [isVoting, setIsVoting] = useState(false);
  const { user, isAuthenticated } = useAuth();

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const handleVote = async (voteValue: number) => {
    if (!isAuthenticated) {
      toast.error('Please log in to vote');
      return;
    }

    setIsVoting(true);
    
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.voteOnComment(comment.commentId, { voteValue });
      
      // Use the new onCommentVoted handler if available, otherwise fall back to onCommentUpdated
      if (onCommentVoted) {
        onCommentVoted(comment.commentId, voteValue);
      } else {
        onCommentUpdated();
      }
    } catch (error: unknown) {
      console.error('Error voting on comment:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to vote';
      toast.error('Failed to vote', {
        description: errorMessage,
      });
    } finally {
      setIsVoting(false);
    }
  };

  const handleReplySubmitted = () => {
    setShowReplyForm(false);
    onCommentUpdated();
  };

  const isOwnComment = user?.activePseudonymId === comment.author.pseudonymId;
  const canEdit = isOwnComment;
  const canRemove = isOwnComment || user?.roles?.includes('moderator');

  return (
    <div className={`${depth > 0 ? 'ml-6 border-l-2 border-border pl-4' : ''}`}>
      <div className="bg-card border border-border rounded-lg p-4 mb-4">
        <div className="flex items-start gap-3">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-sm font-medium">{comment.author.displayName}</span>
              <span className="text-xs text-muted-foreground">
                {formatDate(comment.createdAt)}
              </span>
            </div>
            
            <p className="text-sm text-foreground whitespace-pre-wrap mb-3">
              {comment.content}
            </p>
            
            <div className="flex items-center gap-4">
              <VoteButtons
                score={comment.score}
                userVote={comment.userVote}
                onVote={handleVote}
                disabled={isVoting}
                size="sm"
              />
              <Button 
                variant="ghost" 
                size="sm" 
                className="h-6 px-2"
                onClick={() => setShowReplyForm(!showReplyForm)}
              >
                <Reply className="w-3 h-3 mr-1" />
                Reply
              </Button>
              
              {(canEdit || canRemove) && (
                <div className="flex items-center gap-1">
                  <Button variant="ghost" size="sm" className="h-6 px-2">
                    <MoreHorizontal className="w-3 h-3" />
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
      
      {showReplyForm && (
        <CommentForm
          postId={postId}
          parentCommentId={comment.commentId}
          onCommentSubmitted={handleReplySubmitted}
          onCancel={() => setShowReplyForm(false)}
          placeholder={`Replying to ${comment.author.displayName}...`}
          isReply={true}
        />
      )}
      
      {/* Render replies */}
      {comment.replies && comment.replies.length > 0 && (
        <div className="space-y-2">
          {comment.replies.map((reply) => (
            <Comment
              key={reply.commentId}
              comment={reply}
              postId={postId}
              onCommentUpdated={onCommentUpdated}
              onCommentVoted={onCommentVoted}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
} 