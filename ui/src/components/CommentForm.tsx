'use client';

import { useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { Textarea } from '@/components/shadcn/textarea';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';
import { Send, X } from 'lucide-react';

interface CommentFormProps {
  postId: number;
  parentCommentId?: number;
  onCommentSubmitted: () => void;
  onCancel?: () => void;
  placeholder?: string;
  isReply?: boolean;
}

export default function CommentForm({ 
  postId, 
  parentCommentId, 
  onCommentSubmitted, 
  onCancel,
  placeholder = "Write a comment...",
  isReply = false
}: CommentFormProps) {
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { user, isAuthenticated } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAuthenticated) {
      toast.error('Please log in to comment');
      return;
    }

    if (!content.trim()) {
      toast.error('Comment cannot be empty');
      return;
    }

    setIsSubmitting(true);
    
    try {
      const contentApi = getApi(ContentApi);
      
      const commentData = {
        content: content.trim(),
        ...(parentCommentId && { parentCommentId })
      };

      await contentApi.createComment(postId, commentData);
      
      setContent('');
      toast.success(isReply ? 'Reply posted!' : 'Comment posted!');
      onCommentSubmitted();
    } catch (error: unknown) {
      console.error('Error posting comment:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to post comment';
      toast.error('Failed to post comment', {
        description: errorMessage,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="bg-card border border-border rounded-lg p-4 mb-4">
        <p className="text-muted-foreground text-center">
          Please <button className="text-primary hover:underline">log in</button> to comment
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="mb-4">
      <div className="bg-card border border-border rounded-lg p-4">
        {isReply && (
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm text-muted-foreground">Replying to comment</span>
            {onCancel && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={onCancel}
                className="h-6 px-2"
              >
                <X className="w-3 h-3" />
              </Button>
            )}
          </div>
        )}
        
        <div className="flex items-start gap-3">
          <div className="flex-1">
            <Textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={placeholder}
              className="min-h-[80px] resize-none"
              disabled={isSubmitting}
            />
          </div>
        </div>
        
        <div className="flex items-center justify-between mt-3 pt-3 border-t border-border">
          <div className="text-xs text-muted-foreground">
            Commenting as <span className="font-medium">{user?.displayName}</span>
          </div>
          <Button 
            type="submit" 
            disabled={isSubmitting || !content.trim()}
            size="sm"
          >
            {isSubmitting ? (
              'Posting...'
            ) : (
              <>
                <Send className="w-3 h-3 mr-1" />
                {isReply ? 'Reply' : 'Comment'}
              </>
            )}
          </Button>
        </div>
      </div>
    </form>
  );
} 