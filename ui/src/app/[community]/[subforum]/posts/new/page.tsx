"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Label } from "@/components/shadcn/label";
import { Checkbox } from "@/components/shadcn/checkbox";
import { getApi } from "@/lib/api-client";
import { ContentApi } from "@/generated/api/src/apis/ContentApi";
import { toast } from "sonner";
import { Eye, EyeOff, Lock, Pin } from "lucide-react";
import MarkdownHelp from "@/components/MarkdownHelp";
import { MarkdownPreview } from "@/components/MarkdownPreview";
import { MarkdownTextarea } from "@/components/MarkdownTextarea";
import { useAuth } from "@/lib/auth-context";
import { authenticateUserForSubforum } from "@/lib/auth-utils";

import { COMMUNITY_CONFIG, type CommunityType } from '@/lib/community-config';


export default function NewPostPage() {
  const router = useRouter();
  const params = useParams();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;
  const { isAuthenticated } = useAuth();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const [isSticky, setIsSticky] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [isModerator, setIsModerator] = useState(false);

  const communityConfig = COMMUNITY_CONFIG[communityType];

  // Load subforum-specific user context to check moderator status
  useEffect(() => {
    if (fullSubforumPath && isAuthenticated) {
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, isAuthenticated]);

  const loadSubforumUserContext = async () => {
    try {
      const userData = await authenticateUserForSubforum(fullSubforumPath);
      if (userData) {
        const hasModerateContentCapability = userData.capabilities?.includes('moderate_content');
        setIsModerator(hasModerateContentCapability);
      }
    } catch (error) {
      console.error('Error loading subforum user context:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !content.trim()) {
      toast.error("Title and content are required");
      return;
    }
    setIsSubmitting(true);
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.createPost(fullSubforumPath, {
        title: title.trim(),
        content: content.trim(),
        postType: "text",
        isNsfw: false,
        isSpoiler: false,
        isLocked: isLocked,
        isSticky: isSticky,
      });
      toast.success("Post created successfully!");
      router.push(`/${communityType}/${subforumName}/posts/${response.slug}`);
    } catch {
      toast.error("Failed to create post");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto p-4">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Create a New Post</h1>
        <p className="text-muted-foreground">
          in{' '}
          <span className={`text-xs px-2 py-1 rounded ${communityConfig.color}`}>
            {fullSubforumPath}
          </span>
        </p>
      </div>
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Platform Rules Reminder */}
        <div className="bg-muted/50 border border-border rounded-lg p-4">
          <div className="flex items-start gap-3">
            <div className="w-5 h-5 text-muted-foreground mt-0.5">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                <path d="m9 12 2 2 4-4"/>
              </svg>
            </div>
            <div className="flex-1">
              <p className="text-sm text-muted-foreground">
                <strong>Before posting:</strong> Please ensure your content follows our{' '}
                <span className="text-foreground underline hover:text-foreground/80 transition-colors">
                  platform rules
                </span>
                {' '}and community guidelines. Violations may result in content removal or account restrictions.
              </p>
            </div>
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="title">Title</Label>
          <Input
            id="title"
            value={title}
            onChange={e => setTitle(e.target.value)}
            placeholder="Enter your post title..."
            required
            maxLength={300}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="content">Content</Label>
          <div className="flex items-center justify-between mb-2">
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
            <MarkdownPreview content={content} className="mb-3" />
          )}
          
          <MarkdownTextarea
            id="content"
            value={content}
            onChange={e => setContent(e.target.value)}
            placeholder="Write your post content... (supports markdown)"
            minHeight={200}
            maxHeight={800}
            required
          />
          <MarkdownHelp />
        </div>

        {/* Moderator Controls */}
        {isModerator && (
          <div className="space-y-3 p-4 bg-muted/20 border border-border rounded-lg">
            <Label className="text-sm font-medium">Moderator Options</Label>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="isSticky"
                  checked={isSticky}
                  onCheckedChange={(checked) => setIsSticky(checked as boolean)}
                  disabled={isSubmitting}
                />
                <Label htmlFor="isSticky" className="text-sm flex items-center gap-2">
                  <Pin className="w-4 h-4" />
                  Make this post sticky
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="isLocked"
                  checked={isLocked}
                  onCheckedChange={(checked) => setIsLocked(checked as boolean)}
                  disabled={isSubmitting}
                />
                <Label htmlFor="isLocked" className="text-sm flex items-center gap-2">
                  <Lock className="w-4 h-4" />
                  Lock this post (prevent new comments)
                </Label>
              </div>
            </div>
          </div>
        )}

        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting || !title.trim() || !content.trim()}>
            {isSubmitting ? "Posting..." : "Create Post"}
          </Button>
        </div>
      </form>
    </div>
  );
} 