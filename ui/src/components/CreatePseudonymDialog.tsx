"use client";

import { useState } from "react";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Label } from "@/components/shadcn/label";
import { Textarea } from "@/components/shadcn/textarea";
import { createPseudonym } from "@/lib/pseudonym-utils";
import { useAuth } from "@/lib/auth-context";

interface CreatePseudonymDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function CreatePseudonymDialog({ isOpen, onClose, onSuccess }: CreatePseudonymDialogProps) {
  const [displayName, setDisplayName] = useState("");
  const [slug, setSlug] = useState("");
  const [bio, setBio] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const { refreshUser } = useAuth();

  // Auto-generate slug from display name
  const handleDisplayNameChange = (value: string) => {
    setDisplayName(value);
    // Auto-generate slug from display name
    const generatedSlug = value
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
    setSlug(generatedSlug);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);

    try {
      const result = await createPseudonym(displayName, bio, undefined, slug);
      if (result.success) {
        await refreshUser(); // Refresh user data to get the new pseudonym
        onSuccess();
        onClose();
        setDisplayName("");
        setSlug("");
        setBio("");
      } else {
        setError(result.error || "Failed to create pseudonym");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create pseudonym");
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div 
        className="fixed inset-0 bg-black/50 z-[9999]"
        onClick={onClose}
      />
      <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-background border border-border p-6 rounded-lg z-[10000] min-w-[400px] max-w-[500px] shadow-lg">
        <h2 className="text-lg font-bold mb-4 text-foreground">Create New Pseudonym</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="displayName" className="text-foreground">Display Name</Label>
            <Input
              id="displayName"
              type="text"
              value={displayName}
              onChange={(e) => handleDisplayNameChange(e.target.value)}
              placeholder="Enter display name"
              required
              minLength={3}
              maxLength={50}
              className="mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Must be 3-50 characters, letters, numbers, and underscores only
            </p>
          </div>

          <div>
            <Label htmlFor="slug" className="text-foreground">URL Slug (Optional)</Label>
            <Input
              id="slug"
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="url-slug"
              maxLength={30}
              pattern="[a-z0-9-]+"
              title="Only lowercase letters, numbers, and hyphens allowed"
              className="mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Auto-generated from display name. Used in your profile URL.
            </p>
          </div>

          <div>
            <Label htmlFor="bio" className="text-foreground">Bio (Optional)</Label>
            <Textarea
              id="bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Tell us about yourself..."
              maxLength={500}
              className="mt-1"
              rows={3}
            />
            <p className="text-xs text-muted-foreground mt-1">
              {bio.length}/500 characters
            </p>
          </div>

          {error && (
            <div className="text-sm text-red-600 bg-red-50 dark:bg-red-900/20 p-3 rounded-lg">
              {error}
            </div>
          )}

          <div className="flex justify-end gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !displayName.trim()}
            >
              {isLoading ? "Creating..." : "Create Pseudonym"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
} 