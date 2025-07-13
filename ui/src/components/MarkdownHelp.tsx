'use client';

import { useState } from 'react';
import { Button } from '@/components/shadcn/button';
import { HelpCircle, ChevronDown, ChevronUp } from 'lucide-react';

export default function MarkdownHelp({ className = '' }: { className?: string }) {
  const [expanded, setExpanded] = useState(false);

  const markdownExamples = [
    {
      title: 'Text Formatting',
      examples: [
        { syntax: '**bold text**', description: 'Bold text' },
        { syntax: '*italic text*', description: 'Italic text' },
        { syntax: '~~strikethrough~~', description: 'Strikethrough text' },
        { syntax: '`code`', description: 'Inline code' },
      ]
    },
    {
      title: 'Links & References',
      examples: [
        { syntax: '[link text](url)', description: 'Create a link' },
        { syntax: '![alt text](image-url)', description: 'Add an image' },
      ]
    },
    {
      title: 'Lists',
      examples: [
        { syntax: '- item 1\n- item 2', description: 'Unordered list' },
        { syntax: '1. item 1\n2. item 2', description: 'Ordered list' },
      ]
    },
    {
      title: 'Code Blocks',
      examples: [
        { syntax: '```\ncode block\n```', description: 'Code block' },
        { syntax: '```javascript\nconst x = 1;\n```', description: 'Syntax highlighted code' },
      ]
    },
    {
      title: 'Quotes & Headers',
      examples: [
        { syntax: '> quoted text', description: 'Blockquote' },
        { syntax: '# Header 1\n## Header 2', description: 'Headers' },
      ]
    },
    {
      title: 'Emojis',
      examples: [
        { syntax: ':smile:', description: '😊 Smile emoji' },
        { syntax: ':heart:', description: '❤️ Heart emoji' },
        { syntax: ':thumbsup:', description: '👍 Thumbs up' },
      ]
    }
  ];

  return (
    <div className={`mt-4 ${className}`}>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setExpanded((v) => !v)}
        className="flex items-center gap-2 mb-2"
        aria-expanded={expanded}
        aria-controls="markdown-help-content"
      >
        <HelpCircle className="w-4 h-4" />
        {expanded ? 'Hide Markdown Help' : 'Show Markdown Help'}
        {expanded ? <ChevronUp className="w-4 h-4 ml-1" /> : <ChevronDown className="w-4 h-4 ml-1" />}
      </Button>
      {expanded && (
        <div
          id="markdown-help-content"
          className="bg-muted/50 border border-border rounded-lg p-4 space-y-4 text-sm max-h-72 overflow-y-auto"
        >
          {markdownExamples.map((section) => (
            <div key={section.title}>
              <h4 className="font-medium mb-2">{section.title}</h4>
              <div className="space-y-2">
                {section.examples.map((example, index) => (
                  <div key={index} className="flex items-start gap-2">
                    <code className="bg-background border border-border rounded px-2 py-1 text-xs flex-1 font-mono">
                      {example.syntax}
                    </code>
                    <span className="text-muted-foreground text-xs">
                      {example.description}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
} 