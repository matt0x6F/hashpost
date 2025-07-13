'use client';

import { MarkdownRenderer } from '@/components/MarkdownRenderer';

const testMarkdown = `# Markdown Test Page

This is a **bold** text and this is *italic* text. You can also use ~~strikethrough~~.

## Links
Here's a [link to Google](https://google.com) and another to [GitHub](https://github.com).

## Lists

### Unordered List
- Item 1
- Item 2
  - Nested item
  - Another nested item
- Item 3

### Ordered List
1. First item
2. Second item
3. Third item

## Code

Inline code: \`console.log('Hello World')\`

Code block:
\`\`\`javascript
function hello() {
  console.log('Hello, World!');
  return 'Hello';
}
\`\`\`

## Emojis
Here are some emojis: :smile: :heart: :rocket: :thumbsup:

## Blockquotes
> This is a blockquote
> 
> It can span multiple lines

## Tables

| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |
| Bob  | 35  | SF   |

## Horizontal Rule

---

This is content after the horizontal rule.
`;

export default function MarkdownTestPage() {
  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <h1 className="text-3xl font-bold mb-6">Markdown Test Page</h1>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div>
          <h2 className="text-xl font-semibold mb-4">Raw Markdown</h2>
          <pre className="bg-muted p-4 rounded-lg overflow-x-auto text-sm">
            {testMarkdown}
          </pre>
        </div>
        
        <div>
          <h2 className="text-xl font-semibold mb-4">Rendered Output</h2>
          <div className="bg-card border rounded-lg p-6">
            <MarkdownRenderer content={testMarkdown} />
          </div>
        </div>
      </div>
    </div>
  );
} 