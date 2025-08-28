'use client';

import { ForumList } from '@/components/ForumList';
import { PlatformRulesDisplay } from '@/components/PlatformRulesDisplay';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';

export default function Home() {
  return (
    <div className="max-w-7xl mx-auto p-2 sm:p-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Welcome to HashPost</h1>
        <p className="text-muted-foreground">
          Discover and join communities, or create your own forum to start building something amazing.
        </p>
      </div>
      
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ForumList />
        </div>
        
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Platform Rules</CardTitle>
              <CardDescription>
                These rules apply to all content across HashPost. Please familiarize yourself with them.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <PlatformRulesDisplay compact={true} maxHeight="h-80" />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
