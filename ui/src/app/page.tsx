'use client';

import { ForumList } from '@/components/ForumList';

export default function Home() {
  return (
    <div className="max-w-7xl mx-auto p-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Welcome to HashPost</h1>
        <p className="text-muted-foreground">
          Discover and join communities, or create your own forum to start building something amazing.
        </p>
      </div>
      
      <ForumList />
    </div>
  );
}
