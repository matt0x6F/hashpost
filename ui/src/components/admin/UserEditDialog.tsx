"use client";

import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/shadcn/dialog';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Switch } from '@/components/shadcn/switch';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { AdminApi } from '@/generated/api/src/apis/AdminApi';
import { AdminUserInfo } from '@/generated/api/src/models/AdminUserInfo';
import { UpdateUserInputBody } from '@/generated/api/src/models/UpdateUserInputBody';

interface UserEditDialogProps {
  user: AdminUserInfo | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUserUpdated: (updatedUser: AdminUserInfo) => void;
}

export function UserEditDialog({ user, open, onOpenChange, onUserUpdated }: UserEditDialogProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState<UpdateUserInputBody>({
    email: undefined,
    isActive: undefined,
    isSuspended: undefined,
    emailVerified: undefined,
  });

  // Reset form when user changes
  useEffect(() => {
    if (user) {
      setFormData({
        email: user.email,
        isActive: user.isActive,
        isSuspended: user.isSuspended,
        emailVerified: user.emailVerified,
      });
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;

    setIsLoading(true);
    try {
      const api = getApi(AdminApi);
      
      // Only include fields that have been changed
      const updates: UpdateUserInputBody = {};
      if (formData.email !== user.email) updates.email = formData.email;
      if (formData.isActive !== user.isActive) updates.isActive = formData.isActive;
      if (formData.isSuspended !== user.isSuspended) updates.isSuspended = formData.isSuspended;
      if (formData.emailVerified !== user.emailVerified) updates.emailVerified = formData.emailVerified;

      // If no changes, just close the dialog
      if (Object.keys(updates).length === 0) {
        onOpenChange(false);
        return;
      }

      const response = await api.adminUpdateUser(parseInt(user.userId.toString()), updates);
      console.log('API response:', response);
      
      // Handle both response formats: direct AdminUserInfo or wrapped in body
      let userData = response;
      if (response && response.body && response.body.userId) {
        userData = response.body;
      } else if (response && response.userId) {
        userData = response;
      } else {
        console.error('Invalid response format:', response);
        toast.error("Failed to update user: Invalid response format");
        return;
      }
      
      onUserUpdated(userData);
      toast.success("User updated successfully");
      onOpenChange(false);
    } catch (error: unknown) {
      console.error('Failed to update user:', error);
      toast.error("Failed to update user. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit User</DialogTitle>
          <DialogDescription>
            Update user account details. Password changes are not allowed through this interface.
          </DialogDescription>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={formData.email || ''}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              placeholder="user@example.com"
            />
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="isActive"
              checked={formData.isActive || false}
              onCheckedChange={(checked) => setFormData({ ...formData, isActive: checked })}
            />
            <Label htmlFor="isActive">Account Active</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="isSuspended"
              checked={formData.isSuspended || false}
              onCheckedChange={(checked) => setFormData({ ...formData, isSuspended: checked })}
            />
            <Label htmlFor="isSuspended">Account Suspended</Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="emailVerified"
              checked={formData.emailVerified || false}
              onCheckedChange={(checked) => setFormData({ ...formData, emailVerified: checked })}
            />
            <Label htmlFor="emailVerified">Email Verified</Label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleCancel} disabled={isLoading}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Updating..." : "Update User"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
