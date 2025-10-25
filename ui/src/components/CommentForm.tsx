'use client';

import { useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { CommentsApi } from '@/generated/api/src/apis/CommentsApi';
import { toast } from 'sonner';
import { Send, X, Eye, EyeOff } from 'lucide-react';
import MarkdownHelp from './MarkdownHelp';
import { MarkdownPreview } from './MarkdownPreview';
import { MarkdownTextarea } from './MarkdownTextarea';


interface CommentFormProps {
  postId: string;
  parentCommentId?: string;
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
  const [showPreview, setShowPreview] = useState(false);
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
      // Comment creation not available in atproto system
      toast.error('Comment creation is not available in the atproto system');
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
      {/* Platform Rules Reminder */}
      <div className="bg-muted/30 border border-border rounded-lg p-3 mb-3">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 text-muted-foreground">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              <path d="m9 12 2 2 4-4"/>
            </svg>
          </div>
          <p className="text-xs text-muted-foreground">
            Remember to follow our{' '}
            <span className="text-foreground underline hover:text-foreground/80 transition-colors">
              platform rules
            </span>
            {' '}when commenting.
          </p>
        </div>
      </div>
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
        
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setShowPreview(!showPreview)}
                className="h-8 px-3"
              >
                {showPreview ? <EyeOff className="w-3 h-3 mr-1" /> : <Eye className="w-3 h-3 mr-1" />}
                {showPreview ? 'Hide Preview' : 'Show Preview'}
              </Button>
            </div>
          </div>
          
          {showPreview && (
            <MarkdownPreview content={content} />
          )}
          
          <MarkdownTextarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder={placeholder + " (supports markdown)"}
            minHeight={80}
            maxHeight={400}
            disabled={isSubmitting}
          />
          <div className="mt-2">
            <MarkdownHelp />
          </div>
        </div>
        
        <div className="flex items-center justify-between mt-3 pt-3 border-t border-border">
          <div className="text-xs text-muted-foreground">
            Commenting as <span className="font-medium">{user?.handle}</span>
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