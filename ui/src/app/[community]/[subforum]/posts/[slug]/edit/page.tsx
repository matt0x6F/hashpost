"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Label } from "@/components/shadcn/label";
import { getApi } from "@/lib/api-client";
import { ContentApi } from "@/generated/api/src/apis/ContentApi";
import { PostDetailsResponseBody } from "@/generated/api/src/models";
import { toast } from "sonner";
import { Eye, EyeOff, ArrowLeft } from "lucide-react";
import MarkdownHelp from "@/components/MarkdownHelp";
import { MarkdownPreview } from "@/components/MarkdownPreview";
import { MarkdownTextarea } from "@/components/MarkdownTextarea";
import { useAuth } from "@/lib/auth-context";
import { authenticateUserForSubforum } from "@/lib/auth-utils";
import Link from "next/link";

export default function EditPostPage() {
  const router = useRouter();
  const params = useParams();
  const communityType = params.community as string;
  const subforum = params.subforum as string;
  const slug = params.slug as string;
  const fullSubforumPath = `${communityType}/${subforum}`;
  const { user } = useAuth();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [postDetails, setPostDetails] = useState<PostDetailsResponseBody | null>(null);

  useEffect(() => {
    if (fullSubforumPath && slug) {
      loadPostDetails();
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, slug]);

  const loadSubforumUserContext = async () => {
    try {
      const userData = await authenticateUserForSubforum(fullSubforumPath);
      if (userData) {
        // Update user context if needed
      }
    } catch (error) {
      console.error('Error loading subforum user context:', error);
    }
  };

  const loadPostDetails = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.getPostBySlug(fullSubforumPath, slug, 'best');
      setPostDetails(response);
      
      // Check if user is the author
      if (user?.activePseudonymId !== response.author?.pseudonymId) {
        toast.error("You can only edit your own posts");
        router.push(`/${communityType}/${subforum}/posts/${slug}`);
        return;
      }
      
      setTitle(response.title);
      setContent(response.content || "");
    } catch (err: unknown) {
      console.error('Error loading post details:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load post';
      setError(errorMessage);
      
      toast.error('Failed to load post', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
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
      if (!postDetails) {
        toast.error("Post details not loaded");
        return;
      }
      
      const contentApi = getApi(ContentApi);
      await contentApi.editPost(postDetails.postId, {
        title: title.trim(),
        content: content.trim(),
      });
      toast.success("Post updated successfully!");
      router.push(`/${communityType}/${subforum}/posts/${slug}`);
    } catch {
      toast.error("Failed to update post");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <div className="h-8 w-8 bg-muted animate-pulse rounded" />
          <div className="h-8 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="space-y-4">
          <div className="h-8 bg-muted animate-pulse rounded" />
          <div className="h-4 bg-muted animate-pulse rounded" />
          <div className="h-4 bg-muted animate-pulse rounded w-2/3" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto p-6">
        <div className="flex items-center gap-4 mb-8">
          <Link href={`/${communityType}/${subforum}/posts/${slug}`}>
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Post
            </Button>
          </Link>
        </div>
        <div className="text-center py-12">
          <p className="text-muted-foreground">
            {error}
          </p>
          <Button 
            variant="outline" 
            onClick={loadPostDetails}
            className="mt-4"
          >
            Try Again
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="flex items-center gap-4 mb-8">
        <Link href={`/${communityType}/${subforum}/posts/${slug}`}>
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Post
          </Button>
        </Link>
      </div>
      
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Edit Post</h1>
        <p className="text-muted-foreground">
          in{' '}
          <span className="text-xs px-2 py-1 rounded bg-muted">
            {fullSubforumPath}
          </span>
        </p>
      </div>
      
      <form onSubmit={handleSubmit} className="space-y-6">
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

        <div className="flex justify-end gap-2">
          <Link href={`/${communityType}/${subforum}/posts/${slug}`}>
            <Button type="button" variant="outline">
              Cancel
            </Button>
          </Link>
          <Button type="submit" disabled={isSubmitting || !title.trim() || !content.trim()}>
            {isSubmitting ? "Updating..." : "Update Post"}
          </Button>
        </div>
      </form>
    </div>
  );
} 