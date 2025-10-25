'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/shadcn/button';
import { Textarea } from '@/components/shadcn/textarea';
import { UserDisplay } from '@/components/UserDisplay';
import { 
  MessageSquare, 
  Calendar, 
  User, 
  MoreHorizontal,
  Edit,
  Trash2,
  RotateCcw,
  Flag,
  Reply,
  ChevronDown
} from 'lucide-react';
import VoteButtons from '@/components/VoteButtons';
import { Comment as CommentType } from '@/generated/api/src/models';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { VotingApi, CommentsApi } from '@/generated/api/src/apis';
import { MarkdownRenderer } from '@/components/MarkdownRenderer';
import CommentForm from './CommentForm';
import { ReportDialog } from './ReportDialog';

interface CommentProps {
  comment: CommentType;
  postId: string;
  onCommentUpdated: () => void;
  onCommentVoted: (commentId: string, voteValue: number) => void;
}

export default function Comment({ comment, postId, onCommentUpdated, onCommentVoted }: CommentProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);
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

  const formatDate = (date: Date | string) => {
    const dateObj = typeof date === 'string' ? new Date(date) : date;
    return dateObj.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Check if current user is the comment author
  const isCommentAuthor = user?.did && comment.author && 
                         (user.did === comment.author);
  
  // Check if user is a moderator (not available in atproto system)
  const isModerator = false;

  // In atproto system, comments are immutable once created
  const isDeletedByUser = false;

  // Disable actions for deleted comments
  const isActionDisabled = isSubmitting || isDeleting || isDeletedByUser;

  const handleEdit = async () => {
    if (!editContent.trim()) {
      toast.error('Comment cannot be empty');
      return;
    }

    setIsSubmitting(true);
    try {
      // Comment editing not available in atproto system
      toast.error('Comment editing is not available in the atproto system');
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
      // Comment deletion not available in atproto system
      toast.error('Comment deletion is not available in the atproto system');
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
      // Moderation actions not available in atproto system
      toast.error('Moderation actions are not available in the atproto system');
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

    if (isCommentAuthor) {
      toast.error('You cannot report your own comment');
      return;
    }

    setShowReportDialog(true);
    setShowDropdown(false);
  };

  // If comment is deleted by user, show minimal information
  if (isDeletedByUser) {
    return (
      <div className="relative">
        {/* Thread line */}
        <div className="absolute left-4 top-8 bottom-0 w-px bg-border" />
        
        <div className="flex items-start space-x-3 mb-4">
          {/* Collapse/expand button */}
          <div className="flex-shrink-0 mt-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0 rounded-full hover:bg-muted"
              disabled={!comment.replyCount || comment.replyCount === 0}
              onClick={() => setIsCollapsed(!isCollapsed)}
            >
              <ChevronDown className={`w-4 h-4 transition-transform ${isCollapsed ? 'rotate-180' : ''}`} />
            </Button>
          </div>
          
          {/* Comment content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
              <User className="w-4 h-4" />
              <span>Deleted by user</span>
              {/* Deletion timestamp not available in atproto system */}
            </div>
            {!isCollapsed && (
              <div className="text-sm text-muted-foreground italic">
                This comment has been deleted by the user.
              </div>
            )}
          </div>
        </div>
        
        {/* Show replies even for deleted comments */}
        {comment.replyCount && comment.replyCount > 0 && !isCollapsed && (
          <div className="ml-8 space-y-4">
            {/* Replies not available in atproto system */}
            <div className="text-sm text-muted-foreground italic">
              Replies are not available in the atproto system
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="relative">
      {/* Thread line */}
      <div className="absolute left-4 top-8 bottom-0 w-px bg-border" />
      
      <div className="flex items-start space-x-3 mb-4">
        {/* Collapse/expand button */}
        <div className="flex-shrink-0 mt-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0 rounded-full hover:bg-muted"
            disabled={!comment.replyCount || comment.replyCount === 0}
            onClick={() => setIsCollapsed(!isCollapsed)}
          >
            <ChevronDown className={`w-4 h-4 transition-transform ${isCollapsed ? 'rotate-180' : ''}`} />
          </Button>
        </div>
        
        {/* Comment content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between mb-2">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <User className="w-4 h-4" />
              <UserDisplay author={comment.author} />
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

          {!isCollapsed && (
            <>
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

              <div className="flex items-center gap-4">
                <VoteButtons
                  score={(comment.upvotes || 0) - (comment.downvotes || 0)}
                  userVote={0}
                  onVote={(voteValue) => onCommentVoted(comment.id, voteValue)}
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
                      <span>{comment.replyCount || 0} replies</span>
                    </div>
                  </div>
                )}
              </div>

              {/* Reply form */}
              {showReplyForm && (
                <div className="mt-4">
                  <CommentForm
                    postId={postId}
                    parentCommentId={comment.id}
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
            </>
          )}
        </div>
      </div>

      {/* Nested replies */}
      {comment.replyCount && comment.replyCount > 0 && !isCollapsed && (
        <div className="ml-8 space-y-4">
          {/* Replies not available in atproto system */}
          <div className="text-sm text-muted-foreground italic">
            Replies are not available in the atproto system
          </div>
        </div>
      )}
      
      {/* Report Dialog */}
      <ReportDialog
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        contentType="comment"
        contentId={parseInt(comment.id)}
        reportedPseudonymId={comment.author}
        contentPreview={comment.content}
        reportedUserDisplayName={comment.author}
      />
    </div>
  );
} 