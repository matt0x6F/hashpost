"use client";

import { useState } from "react";
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
import { MarkdownTextarea } from "@/components/MarkdownTextarea";

export default function NewPostPage() {
  const router = useRouter();
  const params = useParams();
  const subforumName = params.subforum as string;
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !content.trim()) {
      toast.error("Title and content are required");
      return;
    }
    setIsSubmitting(true);
    try {
      const contentApi = getApi(ContentApi);
      const response = await contentApi.createPost(subforumName, {
        title: title.trim(),
        content: content.trim(),
        postType: "text",
        isNsfw: false,
        isSpoiler: false,
        isLocked: false,
        isSticky: false,
      });
      toast.success("Post created successfully!");
      router.push(`/h/${subforumName}/posts/${response.slug}`);
    } catch {
      toast.error("Failed to create post");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Create a New Post in h/{subforumName}</h1>
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
        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting || !title.trim() || !content.trim()}>
            {isSubmitting ? "Posting..." : "Create Post"}
          </Button>
        </div>
      </form>
    </div>
  );
} 