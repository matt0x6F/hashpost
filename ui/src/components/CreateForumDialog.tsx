'use client';

import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './shadcn/dialog';
import { Button } from './shadcn/button';
import { Input } from './shadcn/input';
import { Label } from './shadcn/label';
import { Textarea } from './shadcn/textarea';
import { Checkbox } from './shadcn/checkbox';
import { Plus, Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import type { SubforumCreateBody } from '@/generated/api/src/models/SubforumCreateBody';
import { toast } from 'sonner';

interface CreateForumDialogProps {
  onForumCreated?: (forumName: string) => void;
  children?: React.ReactNode;
}

export function CreateForumDialog({ onForumCreated, children }: CreateForumDialogProps) {
  const { user, isAuthenticated } = useAuth();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [formData, setFormData] = useState<SubforumCreateBody>({
    name: '',
    slug: '',
    description: '',
    isNsfw: false,
    isPrivate: false,
    isRestricted: false,
    rulesText: '',
    sidebarText: '',
  });

  const handleInputChange = (field: keyof SubforumCreateBody, value: string | boolean) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
    
    // Auto-generate slug from name
    if (field === 'name') {
      const slug = (value as string)
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
      setFormData(prev => ({
        ...prev,
        name: value as string,
        slug,
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAuthenticated) {
      setError('You must be logged in to create a forum');
      return;
    }

    // Validate required fields
    if (!formData.name.trim()) {
      setError('Forum name is required');
      return;
    }
    if (!formData.description.trim()) {
      setError('Description is required');
      return;
    }
    if (!formData.slug.trim()) {
      setError('Slug is required');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const subforumsApi = getApi(SubforumsApi);
      const response = await subforumsApi.createSubforum(formData);
      
      setOpen(false);
      setFormData({
        name: '',
        slug: '',
        description: '',
        isNsfw: false,
        isPrivate: false,
        isRestricted: false,
        rulesText: '',
        sidebarText: '',
      });
      
      // Show success toast using Sonner
      toast.success(`Forum "r/${response.name}" created successfully!`, {
        description: `Your new forum is now live and ready for discussions.`,
        action: {
          label: "View Forum",
          onClick: () => {
            // TODO: Navigate to the forum page
          },
        },
      });
      
      if (onForumCreated) {
        onForumCreated(response.name);
      }
    } catch (err: unknown) {
      console.error('Error creating forum:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to create forum. Please try again.';
      setError(errorMessage);
      
      // Show error toast using Sonner
      toast.error('Failed to create forum', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const canCreateForum = user?.capabilities?.includes('create_subforum');

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button disabled={!isAuthenticated || !canCreateForum}>
            <Plus className="w-4 h-4 mr-2" />
            Create Forum
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Forum</DialogTitle>
        </DialogHeader>
        
        {!isAuthenticated ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You must be logged in to create a forum.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : !canCreateForum ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You don&apos;t have permission to create forums.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}

            <div className="space-y-6">
              <div>
                <Label htmlFor="name" className="mb-1 block">Forum Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('name', e.target.value)}
                  placeholder="Enter forum name"
                  maxLength={50}
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This will be the display name of your forum
                </p>
              </div>

              <div>
                <Label htmlFor="slug" className="mb-1 block">URL Slug *</Label>
                <Input
                  id="slug"
                  value={formData.slug}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('slug', e.target.value)}
                  placeholder="forum-url-slug"
                  maxLength={30}
                  pattern="[a-z0-9-]+"
                  title="Only lowercase letters, numbers, and hyphens allowed"
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This will be used in the URL: /r/{formData.slug || 'forum-name'}
                </p>
              </div>

              <div>
                <Label htmlFor="description" className="mb-1 block">Description *</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('description', e.target.value)}
                  placeholder="Describe what this forum is about"
                  maxLength={500}
                  rows={3}
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  {formData.description.length}/500 characters
                </p>
              </div>

              <div className="space-y-3">
                <Label className="mb-1 block">Forum Settings</Label>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isNsfw"
                    checked={formData.isNsfw}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isNsfw', checked === true)}
                  />
                  <Label htmlFor="isNsfw" className="text-sm font-normal">
                    NSFW Content
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isPrivate"
                    checked={formData.isPrivate}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isPrivate', checked === true)}
                  />
                  <Label htmlFor="isPrivate" className="text-sm font-normal">
                    Private Forum (requires approval to join)
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isRestricted"
                    checked={formData.isRestricted}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isRestricted', checked === true)}
                  />
                  <Label htmlFor="isRestricted" className="text-sm font-normal">
                    Restricted (only approved users can post)
                  </Label>
                </div>
              </div>

              <div>
                <Label htmlFor="rulesText" className="mb-1 block">Rules (Optional)</Label>
                <Textarea
                  id="rulesText"
                  value={formData.rulesText || ''}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('rulesText', e.target.value)}
                  placeholder="Forum rules and guidelines"
                  maxLength={2000}
                  rows={4}
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  {(formData.rulesText || '').length}/2000 characters
                </p>
              </div>

              <div>
                <Label htmlFor="sidebarText" className="mb-1 block">Sidebar Content (Optional)</Label>
                <Textarea
                  id="sidebarText"
                  value={formData.sidebarText || ''}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('sidebarText', e.target.value)}
                  placeholder="Additional information to display in the sidebar"
                  maxLength={1000}
                  rows={3}
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  {(formData.sidebarText || '').length}/1000 characters
                </p>
              </div>
            </div>

            <div className="flex justify-end space-x-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Forum'
                )}
              </Button>
            </div>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
} 