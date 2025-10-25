'use client';

import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './shadcn/dialog';
import { Button } from './shadcn/button';
import { Input } from './shadcn/input';
import { Label } from './shadcn/label';
import { Textarea } from './shadcn/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './shadcn/select';
import { Plus, Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import { subforumsApiWithRefresh, extractApiErrorMessage } from '@/lib/api-client';
import { CreateSubforumRequest } from '@/generated/api/src/models/CreateSubforumRequest';
import { useForumRefresh } from '@/lib/forum-refresh-context';

interface CreateForumDialogProps {
  onForumCreated?: (forumName: string) => void;
  children?: React.ReactNode;
}

export function CreateForumDialog({ onForumCreated, children }: CreateForumDialogProps) {
  const { user, isAuthenticated } = useAuth();
  const { triggerRefresh } = useForumRefresh();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    slug: '',
    description: '',
    prefix_type: 't',
  });

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
    
    // Auto-generate slug from name
    if (field === 'name') {
      const slug = value
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
      setFormData(prev => ({
        ...prev,
        name: value,
        slug,
      }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAuthenticated) {
      toast.error('Please log in to create a forum');
      return;
    }

    if (!formData.name.trim() || !formData.slug.trim()) {
      toast.error('Please fill in all required fields');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const request: CreateSubforumRequest = {
        name: formData.name.trim(),
        slug: formData.slug.trim(),
        description: formData.description.trim() || undefined,
        prefix_type: formData.prefix_type,
      };

      const createdSubforum = await subforumsApiWithRefresh.createSubforum(request);
      
      toast.success(`Forum "${createdSubforum.name}" created successfully!`);
      
      // Trigger forum list refresh
      triggerRefresh();
      
      // Call the callback if provided
      if (onForumCreated) {
        onForumCreated(createdSubforum.name);
      }
      
      // Close the dialog and reset form
      setOpen(false);
      setFormData({
        name: '',
        slug: '',
        description: '',
        prefix_type: 't',
      });
      setError(null);
      
    } catch (error) {
      console.error('Error creating forum:', error);
      const errorMessage = await extractApiErrorMessage(error);
      setError(errorMessage);
      toast.error(`Failed to create forum: ${errorMessage}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      setFormData({
        name: '',
        slug: '',
        description: '',
        prefix_type: 't',
      });
      setError(null);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="text-center text-muted-foreground">
        Please log in to create a forum
      </div>
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        {children || (
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create Forum
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Forum</DialogTitle>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Forum Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              placeholder="Enter forum name"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="prefix_type">Forum Type *</Label>
            <Select value={formData.prefix_type} onValueChange={(value) => handleInputChange('prefix_type', value)}>
              <SelectTrigger>
                <SelectValue placeholder="Select forum type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="t">Topical (t-) - General discussion topics</SelectItem>
                <SelectItem value="r">Regional (r-) - Location-based discussions</SelectItem>
                <SelectItem value="h">HashPost (h-) - Platform-specific content</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Choose the type of forum you want to create
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="slug">Forum Slug *</Label>
            <Input
              id="slug"
              value={formData.slug}
              onChange={(e) => handleInputChange('slug', e.target.value)}
              placeholder="forum-slug"
              required
            />
            <p className="text-xs text-muted-foreground">
              This will be the URL for your forum (e.g., /forums/{formData.prefix_type}-forum-slug)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              placeholder="Describe your forum..."
              rows={3}
            />
          </div>

          {error && (
            <div className="text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !formData.name.trim() || !formData.slug.trim()}
            >
              {isLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4 mr-2" />
                  Create Forum
                </>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}