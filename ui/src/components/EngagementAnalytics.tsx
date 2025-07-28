'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/shadcn/select';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, LineChart, Line, Area, AreaChart } from 'recharts';
import { Calendar, MessageSquare, ThumbsUp, TrendingUp, Activity } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { ModerationApi, EngagementDataPoint, GetEngagementAnalyticsTimeRangeEnum } from '@/generated/api/src';

interface EngagementAnalyticsProps {
  subforumPath: string;
  timeRange?: '7d' | '14d' | '30d';
}

export function EngagementAnalytics({ subforumPath, timeRange = '14d' }: EngagementAnalyticsProps) {
  const [data, setData] = useState<EngagementDataPoint[]>([]);
  const [selectedTimeRange, setSelectedTimeRange] = useState<'7d' | '14d' | '30d'>(timeRange || '14d');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        const moderationApi = getApi(ModerationApi);
        const timeRangeEnum = selectedTimeRange as GetEngagementAnalyticsTimeRangeEnum;
        const response = await moderationApi.getEngagementAnalytics(subforumPath, timeRangeEnum);
        setData(response.dataPoints || []);
      } catch (err) {
        console.error('Failed to fetch engagement analytics:', err);
        setError('Failed to load engagement analytics. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [selectedTimeRange, subforumPath]);

  const totalPosts = data.reduce((sum, point) => sum + point.posts, 0);
  const totalComments = data.reduce((sum, point) => sum + point.comments, 0);
  const totalVotes = data.reduce((sum, point) => sum + point.totalVotes, 0);
  const avgEngagement = data.length > 0 ? Math.round((totalPosts + totalComments + totalVotes) / data.length) : 0;

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return selectedTimeRange === '7d' ? date.toLocaleDateString('en-US', { weekday: 'short' }) :
           selectedTimeRange === '14d' ? date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) :
           date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const chartData = data.map(point => ({
    ...point,
    date: formatDate(point.date),
  }));

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="w-5 h-5" />
            Engagement Analytics
          </CardTitle>
          <CardDescription>
            Activity metrics for {subforumPath}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="w-5 h-5" />
            Engagement Analytics
          </CardTitle>
          <CardDescription>
            Activity metrics for {subforumPath}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <p className="text-muted-foreground mb-2">{error}</p>
              <button 
                onClick={() => window.location.reload()} 
                className="text-sm text-primary hover:underline"
              >
                Try again
              </button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Posts</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalPosts}</div>
            <p className="text-xs text-muted-foreground">
              Last {selectedTimeRange}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Comments</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalComments}</div>
            <p className="text-xs text-muted-foreground">
              Last {selectedTimeRange}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Votes</CardTitle>
            <ThumbsUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalVotes}</div>
            <p className="text-xs text-muted-foreground">
              Last {selectedTimeRange}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Engagement</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{avgEngagement}</div>
            <p className="text-xs text-muted-foreground">
              Per day
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Time Range Selector */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Activity className="w-5 h-5" />
                Engagement Analytics
              </CardTitle>
              <CardDescription>
                Activity metrics for {subforumPath}
              </CardDescription>
            </div>
            <Select value={selectedTimeRange} onValueChange={(value: '7d' | '14d' | '30d') => setSelectedTimeRange(value)}>
              <SelectTrigger className="w-36">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7d">Last 7 days</SelectItem>
                <SelectItem value="14d">Last 14 days</SelectItem>
                <SelectItem value="30d">Last 30 days</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Combined Activity Chart */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Combined Activity</h3>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Area type="monotone" dataKey="posts" stackId="1" stroke="#8884d8" fill="#8884d8" name="Posts" />
                <Area type="monotone" dataKey="comments" stackId="1" stroke="#82ca9d" fill="#82ca9d" name="Comments" />
                <Area type="monotone" dataKey="totalVotes" stackId="1" stroke="#ffc658" fill="#ffc658" name="Votes" />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Voting Breakdown */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Voting Activity</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="postVotes" fill="#8884d8" name="Post Votes" />
                <Bar dataKey="commentVotes" fill="#82ca9d" name="Comment Votes" />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Content Creation Trend */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Content Creation Trend</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="posts" stroke="#8884d8" strokeWidth={2} name="Posts" />
                <Line type="monotone" dataKey="comments" stroke="#82ca9d" strokeWidth={2} name="Comments" />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Voting Sentiment */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Voting Sentiment</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="postUpvotes" stackId="posts" fill="#22c55e" name="Post Upvotes" />
                <Bar dataKey="postDownvotes" stackId="posts" fill="#ef4444" name="Post Downvotes" />
                <Bar dataKey="commentUpvotes" stackId="comments" fill="#10b981" name="Comment Upvotes" />
                <Bar dataKey="commentDownvotes" stackId="comments" fill="#f87171" name="Comment Downvotes" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
} 