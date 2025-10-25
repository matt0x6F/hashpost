'use client';

import { useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/shadcn/dialog';
import { Label } from '@/components/shadcn/label';
import { Input } from '@/components/shadcn/input';
import { getApi } from '@/lib/api-client';
import { toast } from 'sonner';
import type { Comment } from '@/generated/api/src/models';
import { Eye, EyeOff } from 'lucide-react';
import MarkdownHelp from './MarkdownHelp';
import { MarkdownPreview } from './MarkdownPreview';
import { MarkdownTextarea } from './MarkdownTextarea';

interface EditCommentDialogProps {
  comment: Comment;
  isOpen: boolean;
  onClose: () => void;
  onCommentEdited: () => void;
}

export default function EditCommentDialog({
  comment,
  isOpen,
  onClose,
  onCommentEdited
}: EditCommentDialogProps) {
  const [content, setContent] = useState(comment.content);
  const [editReason, setEditReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!content.trim()) {
      toast.error('Comment content cannot be empty');
      return;
    }

    setIsSubmitting(true);
    
    try {
      // Comment editing not available in atproto system
      toast.error('Comment editing is not available in the atproto system');
    } catch (error: unknown) {
      console.error('Error editing comment:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to edit comment';
      toast.error('Failed to edit comment', {
        description: errorMessage,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setContent(comment.content);
    setEditReason('');
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="w-full max-w-4xl min-w-3xl">
        <DialogHeader>
          <DialogTitle>Edit Comment</DialogTitle>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="content">Comment</Label>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setShowPreview(!showPreview)}
                  className="h-8 px-3"
                  disabled={isSubmitting}
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
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Enter your comment... (supports markdown)"
              minHeight={120}
              maxHeight={600}
              required
            />
            <MarkdownHelp />
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="editReason">Edit Reason (Optional)</Label>
            <Input
              id="editReason"
              value={editReason}
              onChange={(e) => setEditReason(e.target.value)}
              placeholder="e.g., Fixed typo, Clarified point..."
            />
          </div>
          
          <div className="flex justify-end gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || !content.trim()}
            >
              {isSubmitting ? 'Updating...' : 'Update Comment'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
} 