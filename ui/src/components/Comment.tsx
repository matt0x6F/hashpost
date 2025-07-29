'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/shadcn/button';
import { Textarea } from '@/components/shadcn/textarea';
import { 
  MessageSquare, 
  Calendar, 
  User, 
  MoreHorizontal,
  Edit,
  Trash2,
  RotateCcw,
  Flag,
  Reply
} from 'lucide-react';
import VoteButtons from '@/components/VoteButtons';
import { Comment as CommentType } from '@/generated/api/src/models';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { MarkdownRenderer } from '@/components/MarkdownRenderer';
import CommentForm from './CommentForm';
import { ReportDialog } from './ReportDialog';

interface CommentProps {
  comment: CommentType;
  postId: number;
  onCommentUpdated: () => void;
  onCommentVoted: (commentId: number, voteValue: number) => void;
}

export default function Comment({ comment, postId, onCommentUpdated, onCommentVoted }: CommentProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const { user, isAuthenticated } = useAuth();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      // Don't close if clicking on the dropdown itself
      const target = event.target as Element;
      if (target.closest('.dropdown-menu')) {
        return;
      }
      
      if (showDropdown) {
        setShowDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showDropdown]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Check if current user is the comment author
  const isCommentAuthor = user?.activePseudonymId && comment.author?.pseudonymId && 
                         (user.activePseudonymId === comment.author.pseudonymId);
  
  // Check if user is a moderator
  const isModerator = user?.roles?.includes('moderator') || user?.capabilities?.includes('moderate_content');

  // Check if comment is deleted by user
  const isDeletedByUser = comment.isDeleted;

  // Disable actions for deleted comments
  const isActionDisabled = isSubmitting || isDeleting || isDeletedByUser;

  const handleEdit = async () => {
    if (!editContent.trim()) {
      toast.error('Comment cannot be empty');
      return;
    }

    setIsSubmitting(true);
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.editComment(comment.commentId, { content: editContent });
      setIsEditing(false);
      onCommentUpdated();
      toast.success('Comment updated');
    } catch (error: unknown) {
      console.error('Error editing comment:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to edit comment';
      toast.error('Failed to edit comment', { description: errorMessage });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this comment? This action cannot be undone.')) {
      return;
    }

    setIsDeleting(true);
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.deleteComment(comment.commentId, { reason: 'User requested deletion' });
      onCommentUpdated();
      toast.success('Comment deleted');
    } catch (error: unknown) {
      console.error('Error deleting comment:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete comment';
      toast.error('Failed to delete comment', { description: errorMessage });
    } finally {
      setIsDeleting(false);
      setShowDropdown(false);
    }
  };

  const handleModeratorAction = async (action: string, value: boolean) => {
    if (!isModerator) {
      toast.error('You do not have permission to perform this action');
      return;
    }

    setIsSubmitting(true);
    try {
      const contentApi = getApi(ContentApi);
      switch (action) {
        case 'remove':
          await contentApi.removeComment(comment.commentId, { removed: value });
          break;
      }
      onCommentUpdated();
      toast.success(`Comment ${action === 'remove' ? (value ? 'removed' : 'restored') : 'updated'}`);
    } catch (error: unknown) {
      console.error('Error performing moderator action:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to perform action';
      toast.error('Failed to perform action', { description: errorMessage });
    } finally {
      setIsSubmitting(false);
      setShowDropdown(false);
    }
  };

  const handleReport = () => {
    console.log('handleReport called', { isCommentAuthor, isAuthenticated });
    if (isCommentAuthor) {
      toast.error('You cannot report your own comment');
      return;
    }
    console.log('Setting showReportDialog to true');
    setShowReportDialog(true);
    setShowDropdown(false);
  };

  // If comment is deleted by user, show minimal information
  if (isDeletedByUser) {
    return (
      <div className="bg-muted/50 border border-border rounded-lg p-4 mb-4">
        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
          <User className="w-4 h-4" />
          <span>Deleted by user</span>
          {comment.deletedAt && (
            <>
              <span>•</span>
              <Calendar className="w-4 h-4" />
              <span>{formatDate(comment.deletedAt)}</span>
            </>
          )}
        </div>
        <div className="text-sm text-muted-foreground italic">
          This comment has been deleted by the user.
        </div>
        
        {/* Show replies even for deleted comments */}
        {comment.replies && comment.replies.length > 0 && (
          <div className="mt-4 space-y-4">
            {comment.replies.map((reply) => (
              <Comment
                key={reply.commentId}
                comment={reply}
                postId={postId}
                onCommentUpdated={onCommentUpdated}
                onCommentVoted={onCommentVoted}
              />
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="bg-card border border-border rounded-lg p-4 mb-4">
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <User className="w-4 h-4" />
          <span>{comment.author?.displayName || 'Unknown'}</span>
          <span>•</span>
          <Calendar className="w-4 h-4" />
          <span>{formatDate(comment.createdAt)}</span>
        </div>
        
        {!isDeletedByUser && (
          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDropdown(!showDropdown)}
              disabled={isActionDisabled}
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>
            {showDropdown && (
              <div className="absolute right-0 top-full mt-1 bg-popover border border-border rounded-md shadow-lg z-10 min-w-48 dropdown-menu">
                <div className="p-2 space-y-1">
                  {isCommentAuthor && (
                    <>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={() => {
                          setIsEditing(true);
                          setShowDropdown(false);
                        }}
                        disabled={isSubmitting}
                      >
                        <Edit className="w-4 h-4 mr-2" />
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start text-destructive hover:text-destructive"
                        onClick={handleDelete}
                        disabled={isDeleting}
                      >
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                      </Button>
                    </>
                  )}
                  {isCommentAuthor && isModerator && <hr className="my-2 border-border" />}
                  {isModerator && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-destructive hover:text-destructive"
                      onClick={() => handleModeratorAction('remove', true)}
                      disabled={isSubmitting}
                    >
                      <RotateCcw className="w-4 h-4 mr-2" />
                      Remove Comment
                    </Button>
                  )}
                  {!isCommentAuthor && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-orange-600 hover:text-orange-700"
                      onClick={() => {
                        console.log('Report button clicked');
                        handleReport();
                      }}
                      disabled={isSubmitting}
                    >
                      <Flag className="w-4 h-4 mr-2" />
                      Report Comment
                    </Button>
                  )}
                  
                  {/* Report button for non-authenticated users */}
                  {!isCommentAuthor && !isAuthenticated && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start text-muted-foreground"
                      onClick={() => {
                        toast.error('Please log in to report comments');
                        setShowDropdown(false);
                      }}
                      disabled={isSubmitting}
                    >
                      <Flag className="w-4 h-4 mr-2" />
                      Report Comment (Login Required)
                    </Button>
                  )}

                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {isEditing ? (
        <div className="space-y-2">
          <Textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            placeholder="Edit your comment..."
            className="min-h-[100px]"
          />
          <div className="flex gap-2">
            <Button
              onClick={handleEdit}
              disabled={isSubmitting || !editContent.trim()}
              size="sm"
            >
              Save
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setIsEditing(false);
                setEditContent(comment.content);
              }}
              disabled={isSubmitting}
              size="sm"
            >
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <div className="prose prose-sm max-w-none mb-4">
          <MarkdownRenderer content={comment.content} />
        </div>
      )}

      <div className="flex items-center justify-between">
        <VoteButtons
          score={comment.score}
          userVote={comment.userVote}
          onVote={(voteValue) => onCommentVoted(comment.commentId, voteValue)}
          disabled={isActionDisabled}
          size="sm"
        />
        
        {!isDeletedByUser && (
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowReplyForm(!showReplyForm)}
              disabled={isActionDisabled}
              className="text-muted-foreground hover:text-foreground"
            >
              <Reply className="w-4 h-4 mr-1" />
              Reply
            </Button>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <MessageSquare className="w-4 h-4" />
              <span>{comment.replies?.length || 0} replies</span>
            </div>
          </div>
        )}
      </div>

      {/* Reply form */}
      {showReplyForm && (
        <div className="mt-4">
          <CommentForm
            postId={postId}
            parentCommentId={comment.commentId}
            onCommentSubmitted={() => {
              onCommentUpdated();
              setShowReplyForm(false);
            }}
            onCancel={() => setShowReplyForm(false)}
            placeholder="Write a reply..."
            isReply={true}
          />
        </div>
      )}

      {/* Nested replies */}
      {comment.replies && comment.replies.length > 0 && (
        <div className="mt-4 space-y-4">
          {comment.replies.map((reply) => (
            <Comment
              key={reply.commentId}
              comment={reply}
              postId={postId}
              onCommentUpdated={onCommentUpdated}
              onCommentVoted={onCommentVoted}
            />
          ))}
        </div>
      )}
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="comment"
        contentId={comment.commentId}
        reportedPseudonymId={comment.author?.pseudonymId}
        contentPreview={comment.content}
        reportedUserDisplayName={comment.author?.displayName}
      />
    </div>
  );
} 