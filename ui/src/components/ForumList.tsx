'use client';

import React, { useState, useEffect } from 'react';
import { Button } from './shadcn/button';
import { Input } from './shadcn/input';
import { Label } from './shadcn/label';
import { Search, Users, MessageSquare, Calendar, Shield } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import type { Subforum } from '@/generated/api/src/models/Subforum';
import { toast } from 'sonner';
import Link from 'next/link';

interface ForumListProps {
  className?: string;
}

export function ForumList({ className }: ForumListProps) {
  const [forums, setForums] = useState<Subforum[]>([]);
  const [filteredForums, setFilteredForums] = useState<Subforum[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'name' | 'subscribers' | 'posts' | 'created_at'>('subscribers');

  useEffect(() => {
    loadForums();
  }, []);

  useEffect(() => {
    filterAndSortForums();
  }, [forums, searchTerm, sortBy]);

  const loadForums = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const subforumsApi = getApi(SubforumsApi);
      const response = await subforumsApi.getSubforums(1, 100, sortBy);
      setForums(response.subforums || []);
    } catch (err: unknown) {
      console.error('Error loading forums:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load forums';
      setError(errorMessage);
      
      // Show error toast using Sonner
      toast.error('Failed to load forums', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const filterAndSortForums = () => {
    let filtered = forums;

    // Filter by search term
    if (searchTerm.trim()) {
      filtered = forums.filter(forum => 
        forum.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        forum.displayName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        forum.description.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Sort forums
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.displayName.localeCompare(b.displayName);
        case 'subscribers':
          return (b.subscriberCount || 0) - (a.subscriberCount || 0);
        case 'posts':
          return (b.postCount || 0) - (a.postCount || 0);
        case 'created_at':
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        default:
          return 0;
      }
    });

    setFilteredForums(filtered);
  };

  const formatDate = (date: Date) => {
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    }).format(date);
  };

  if (isLoading) {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">Forums</h2>
          <div className="h-10 w-32 bg-muted animate-pulse rounded" />
        </div>
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-20 bg-muted animate-pulse rounded" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Forums</h2>
      </div>

      {error && (
        <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-md">
          <p className="text-sm text-destructive">{error}</p>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={loadForums}
            className="mt-2"
          >
            Try Again
          </Button>
        </div>
      )}

      <div className="space-y-4">
        <div className="flex gap-4">
          <div className="flex-1">
            <Label htmlFor="search" className="sr-only">Search forums</Label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
              <Input
                id="search"
                placeholder="Search forums..."
                value={searchTerm}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
          <div>
            <Label htmlFor="sort" className="sr-only">Sort by</Label>
            <select
              id="sort"
              value={sortBy}
              onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setSortBy(e.target.value as 'name' | 'subscribers' | 'posts' | 'created_at')}
              className="h-10 px-3 py-2 border border-input bg-background rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="subscribers">Most Popular</option>
              <option value="posts">Most Active</option>
              <option value="name">Alphabetical</option>
              <option value="created_at">Newest</option>
            </select>
          </div>
        </div>

        {filteredForums.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">
              {searchTerm ? 'No forums found matching your search.' : 'No forums available.'}
            </p>
          </div>
        ) : (
          <div className="grid gap-4">
            {filteredForums.map((forum) => (
              <Link key={forum.name} href={`/h/${forum.name}`} className="block">
                <div
                  className="p-4 border border-border rounded-lg cursor-pointer"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-lg font-semibold truncate">
                          h/{forum.name}
                        </h3>
                        <div className="flex items-center gap-1">
                          {forum.isPrivate && (
                            <Shield className="w-4 h-4 text-muted-foreground" />
                          )}
                          {forum.isNsfw && (
                            <span className="text-xs bg-destructive/10 text-destructive px-2 py-1 rounded">
                              NSFW
                            </span>
                          )}
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground mb-3 line-clamp-2">
                        {forum.description}
                      </p>
                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Users className="w-3 h-3" />
                          <span>{forum.subscriberCount?.toLocaleString() || 0} members</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <MessageSquare className="w-3 h-3" />
                          <span>{forum.postCount?.toLocaleString() || 0} posts</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Calendar className="w-3 h-3" />
                          <span>Created {formatDate(forum.createdAt)}</span>
                        </div>
                      </div>
                    </div>
                    <Button variant="outline" size="sm">
                      View
                    </Button>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
} 