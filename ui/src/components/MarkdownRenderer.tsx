'use client';

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkNodeEmoji from '@/lib/remark-node-emoji';

import rehypeHighlight from 'rehype-highlight';
import { cn } from '@/lib/utils';
import type { Components } from 'react-markdown';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
  const components: Components = {
    // Links with custom styling
    a: ({ href, children, ...props }) => (
      <a
        href={href}
        className="text-primary hover:underline focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 rounded"
        target="_blank"
        rel="noopener noreferrer"
        {...props}
      >
        {children}
      </a>
    ),

    // Code blocks with syntax highlighting
    code: ({ children, className, ...props }) => {
      // Check if this is an inline code or block code based on the node
      const isInline = !className?.includes('language-');
      const match = /language-(\w+)/.exec(className || '');
      
      return !isInline ? (
        <pre className="bg-muted border rounded-lg p-4 overflow-x-auto">
          <code
            className={cn(
              'text-sm',
              match && `language-${match[1]}`,
              className
            )}
            {...props}
          >
            {children}
          </code>
        </pre>
      ) : (
        <code
          className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono"
          {...props}
        >
          {children}
        </code>
      );
    },

    // Headings with proper hierarchy
    h1: ({ children, ...props }) => (
      <h1 className="text-2xl font-bold mt-6 mb-4" {...props}>
        {children}
      </h1>
    ),
    h2: ({ children, ...props }) => (
      <h2 className="text-xl font-semibold mt-5 mb-3" {...props}>
        {children}
      </h2>
    ),
    h3: ({ children, ...props }) => (
      <h3 className="text-lg font-semibold mt-4 mb-2" {...props}>
        {children}
      </h3>
    ),

    // Lists with proper spacing
    ul: ({ children, ...props }) => (
      <ul className="list-disc list-inside space-y-1 my-4" {...props}>
        {children}
      </ul>
    ),
    ol: ({ children, ...props }) => (
      <ol className="list-decimal list-inside space-y-1 my-4" {...props}>
        {children}
      </ol>
    ),
    li: ({ children, ...props }) => (
      <li className="ml-2" {...props}>
        {children}
      </li>
    ),

    // Blockquotes
    blockquote: ({ children, ...props }) => (
      <blockquote
        className="border-l-4 border-primary pl-4 italic text-muted-foreground my-4"
        {...props}
      >
        {children}
      </blockquote>
    ),

    // Tables (from GFM)
    table: ({ children, ...props }) => (
      <div className="overflow-x-auto my-4">
        <table className="w-full border-collapse border border-border" {...props}>
          {children}
        </table>
      </div>
    ),
    thead: ({ children, ...props }) => (
      <thead className="bg-muted" {...props}>
        {children}
      </thead>
    ),
    th: ({ children, ...props }) => (
      <th className="border border-border px-3 py-2 text-left font-semibold" {...props}>
        {children}
      </th>
    ),
    td: ({ children, ...props }) => (
      <td className="border border-border px-3 py-2" {...props}>
        {children}
      </td>
    ),

    // Paragraphs with proper spacing
    p: ({ children, ...props }) => (
      <p className="mb-4 last:mb-0" {...props}>
        {children}
      </p>
    ),

    // Horizontal rules
    hr: ({ ...props }) => (
      <hr className="border-border my-6" {...props} />
    ),
  };

  return (
    <div className={cn('prose prose-sm max-w-none', className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkNodeEmoji]}
        rehypePlugins={[rehypeHighlight]}
        components={components}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
} 