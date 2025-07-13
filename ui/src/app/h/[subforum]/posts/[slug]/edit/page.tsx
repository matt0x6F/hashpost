"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Label } from "@/components/shadcn/label";
import { getApi } from "@/lib/api-client";
import { ContentApi } from "@/generated/api/src/apis/ContentApi";
import { toast } from "sonner";
import { Eye, EyeOff } from "lucide-react";
import MarkdownHelp from "@/components/MarkdownHelp";
import { MarkdownPreview } from "@/components/MarkdownPreview";
import { AutoResizeTextarea } from "@/components/AutoResizeTextarea";

export default function EditPostPage() {
  const router = useRouter();
  const params = useParams();
  const subforum = params.subforum as string;
  const slug = params.slug as string;
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  useEffect(() => {
    const fetchPost = async () => {
      setIsLoading(true);
      try {
        const contentApi = getApi(ContentApi);
        const post = await contentApi.getPostBySlug(subforum, slug);
        setTitle(post.title);
        setContent(post.content);
      } catch {
        toast.error("Failed to load post");
        router.push(`/h/${subforum}/posts/${slug}`);
      } finally {
        setIsLoading(false);
      }
    };
    fetchPost();
  }, [subforum, slug, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !content.trim() || !postId) {
      toast.error("Title, content, and postId are required");
      return;
    }
    setIsSubmitting(true);
    try {
      const contentApi = getApi(ContentApi);
      await contentApi.editPost(postId, { title: title.trim(), content: content.trim() });
      toast.success("Post updated successfully!");
      router.push(`/h/${subforum}/posts/${slug}`);
    } catch {
      toast.error("Failed to update post");
    } finally {
      setIsSubmitting(false);
    }
  };

  // We need postId for editPost, so fetch it from the loaded post
  const [postId, setPostId] = useState<number | null>(null);
  useEffect(() => {
    if (!isLoading && title && content) {
      // Fetch postId from the API again (since getPostBySlug returns it)
      const fetchPostId = async () => {
        try {
          const contentApi = getApi(ContentApi);
          const post = await contentApi.getPostBySlug(subforum, slug);
          setPostId(post.postId);
        } catch {}
      };
      fetchPostId();
    }
  }, [isLoading, title, content, subforum, slug]);

  if (isLoading) {
    return <div className="max-w-2xl mx-auto p-6 text-center">Loading...</div>;
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Edit Post</h1>
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
          
                      <AutoResizeTextarea
              id="content"
              value={content}
              onChange={e => setContent(e.target.value)}
              placeholder="Edit your post content... (supports markdown)"
              minHeight={200}
              maxHeight={800}
              required
            />
          <MarkdownHelp />
        </div>
        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting || !title.trim() || !content.trim() || !postId}>
            {isSubmitting ? "Saving..." : "Save Changes"}
          </Button>
        </div>
      </form>
    </div>
  );
} 