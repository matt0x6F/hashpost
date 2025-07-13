import React from 'react';
import { MarkdownRenderer } from './MarkdownRenderer';

interface MarkdownPreviewProps {
  content: string;
  className?: string;
}

export function MarkdownPreview({ content, className = "" }: MarkdownPreviewProps) {
  if (!content.trim()) {
    return (
      <div className={`p-4 border border-dashed border-muted-foreground/20 rounded-md bg-muted/10 ${className}`}>
        <p className="text-sm text-muted-foreground text-center">
          Preview will appear here as you type...
        </p>
      </div>
    );
  }

  return (
    <div className={`p-4 border border-border rounded-md bg-card ${className}`}>
      <div className="prose prose-sm max-w-none">
        <MarkdownRenderer content={content} />
      </div>
    </div>
  );
} 