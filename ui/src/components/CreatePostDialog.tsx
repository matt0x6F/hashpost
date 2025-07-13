'use client';

import React, { useState } from 'react';
import { Button } from './shadcn/button';
import { Input } from './shadcn/input';
import { Label } from './shadcn/label';
import { Textarea } from './shadcn/textarea';
import { Checkbox } from './shadcn/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './shadcn/dialog';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { ContentApi } from '@/generated/api/src/apis/ContentApi';
import { toast } from 'sonner';
import { Plus, Pin, Lock } from 'lucide-react';

interface CreatePostDialogProps {
  subforumName: string;
  onPostCreated?: (postId: number) => void;
  children?: React.ReactNode;
}

export function CreatePostDialog({ subforumName, onPostCreated, children }: CreatePostDialogProps) {
  const { isAuthenticated, user } = useAuth();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    isSticky: false,
    isLocked: false,
  });

  // Debug logging
  console.log('[CreatePostDialog] User capabilities:', user?.capabilities);
  console.log('[CreatePostDialog] User roles:', user?.roles);
  console.log('[CreatePostDialog] Can create post:', user?.capabilities?.includes('create_content'));
  console.log('[CreatePostDialog] Is moderator:', user?.capabilities?.includes('moderate_content'));

  const isModerator = user?.capabilities?.includes('moderate_content') || 
                     user?.roles?.includes('platform_admin');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAuthenticated) {
      setError('You must be logged in to create a post');
      return;
    }

    // Validate required fields
    if (!formData.title.trim()) {
      setError('Title is required');
      return;
    }
    if (!formData.content.trim()) {
      setError('Content is required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.createPost(
        subforumName,
        {
          title: formData.title,
          content: formData.content,
          postType: 'text',
          isNsfw: false,
          isSpoiler: false,
          isSticky: formData.isSticky,
          isLocked: formData.isLocked,
        }
      );

      toast.success('Post created successfully!');
      setOpen(false);
      onPostCreated?.(response.postId);
    } catch (error) {
      console.error('Error creating post:', error);
      setError('Failed to create post. Please try again.');
      toast.error('Failed to create post');
    } finally {
      setIsLoading(false);
    }
  };

  const canCreatePost = user?.capabilities?.includes('create_content');

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Post
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Post in h/{subforumName}</DialogTitle>
          <DialogDescription>
            Share your thoughts, questions, or content with the community.
          </DialogDescription>
        </DialogHeader>
        
        {!isAuthenticated ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You must be logged in to create a post.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : !canCreatePost ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You don&apos;t have permission to create posts.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}
            
            <div className="space-y-2">
              <Label htmlFor="title">Title</Label>
              <Input
                id="title"
                placeholder="Enter your post title..."
                value={formData.title}
                onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                required
                maxLength={300}
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="content">Content</Label>
              <Textarea
                id="content"
                placeholder="Write your post content here..."
                value={formData.content}
                onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                required
                rows={8}
                maxLength={10000}
              />
              <p className="text-xs text-muted-foreground">
                {formData.content.length}/10,000 characters
              </p>
            </div>
            
            {/* Moderator Controls */}
            {isModerator && (
              <div className="space-y-3 p-4 bg-muted/20 rounded-lg border">
                <h4 className="text-sm font-medium text-muted-foreground">Moderator Options</h4>
                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="isSticky"
                      checked={formData.isSticky}
                      onCheckedChange={(checked: boolean | 'indeterminate') => 
                        setFormData(prev => ({ ...prev, isSticky: checked === true }))
                      }
                    />
                    <Label htmlFor="isSticky" className="text-sm font-normal flex items-center gap-2">
                      <Pin className="w-4 h-4" />
                      Sticky Post
                    </Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="isLocked"
                      checked={formData.isLocked}
                      onCheckedChange={(checked: boolean | 'indeterminate') => 
                        setFormData(prev => ({ ...prev, isLocked: checked === true }))
                      }
                    />
                    <Label htmlFor="isLocked" className="text-sm font-normal flex items-center gap-2">
                      <Lock className="w-4 h-4" />
                      Lock Post
                    </Label>
                  </div>
                </div>
              </div>
            )}
            
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Creating...' : 'Create Post'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
} 