'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Flag, Users, History, Calendar, MessageSquare, TrendingUp } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { ModerationApi, GetModerationStatsTimeRangeEnum } from '@/generated/api/src';
import type { ModerationStats } from '@/generated/api/src';

interface ModerationStatsProps {
  subforumPath: string;
  timeRange?: '7d' | '14d' | '30d';
}

export function ModerationStats({ subforumPath, timeRange = '30d' }: ModerationStatsProps) {
  const [stats, setStats] = useState<ModerationStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const lastFetchKey = useRef<string>('');

  useEffect(() => {
    const fetchKey = `${subforumPath}-${timeRange}`;
    if (lastFetchKey.current === fetchKey) return;
    
    const fetchStats = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        const moderationApi = getApi(ModerationApi);
        const timeRangeEnum = timeRange as GetModerationStatsTimeRangeEnum;
        const response = await moderationApi.getModerationStats(subforumPath, timeRangeEnum);
        setStats(response);
        lastFetchKey.current = fetchKey;
      } catch (err) {
        console.error('Failed to fetch moderation stats:', err);
        setError('Failed to load moderation stats. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchStats();
  }, [subforumPath, timeRange]);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Loading...</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                <div className="animate-pulse bg-muted h-8 w-16 rounded"></div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-muted-foreground mb-2">{error}</p>
              <button 
                onClick={() => window.location.reload()} 
                className="text-sm text-primary hover:underline"
              >
                Try again
              </button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  return (
    <div className="space-y-4">
      {/* Content Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Posts</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalPosts}</div>
            <p className="text-xs text-muted-foreground">Last {timeRange.replace('d', ' days')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Comments</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalComments}</div>
            <p className="text-xs text-muted-foreground">Last {timeRange.replace('d', ' days')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Votes</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalVotes}</div>
            <p className="text-xs text-muted-foreground">Last {timeRange.replace('d', ' days')}</p>
          </CardContent>
        </Card>
      </div>

      {/* Moderation Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pending Reports</CardTitle>
            <Flag className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pendingReports}</div>
            <p className="text-xs text-muted-foreground">Reports awaiting review</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Banned Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.bannedUsers}</div>
            <p className="text-xs text-muted-foreground">Users currently banned</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Mod Actions</CardTitle>
            <History className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.modActions}</div>
            <p className="text-xs text-muted-foreground">Actions this week</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
} 