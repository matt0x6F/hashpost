'use client';

import { ForumList } from '@/components/ForumList';

export default function ForumsPage() {
  return (
    <div className="max-w-7xl mx-auto p-2 sm:p-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Forums</h1>
        <p className="text-muted-foreground">
          Browse all available forums and communities on HashPost.
        </p>
      </div>
      
      <ForumList />
    </div>
  );
} 