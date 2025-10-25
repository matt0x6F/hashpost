'use client';

import React, { useEffect, useRef, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/shadcn/select';
import { Activity } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, AreaChart, Area, BarChart, Bar } from 'recharts';
import { getApi } from '@/lib/api-client';

interface EngagementAnalyticsProps {
  subforumPath: string;
  timeRange?: '7d' | '14d' | '30d';
}

// Analytics not available in atproto system
export function EngagementAnalytics({ subforumPath, timeRange = '30d' }: EngagementAnalyticsProps) {
  const [data, setData] = useState<any[]>([]);
  const [selectedTimeRange, setSelectedTimeRange] = useState<'7d' | '14d' | '30d'>(timeRange || '30d');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const lastFetchKey = useRef<string>('');

  useEffect(() => {
    const fetchKey = `${subforumPath}-${selectedTimeRange}`;
    if (lastFetchKey.current === fetchKey) return;
    
    const fetchData = async () => {
      setIsLoading(true);
      setError(null);
      
      try {
        // Analytics not available in atproto system
        setData([]);
        setError('Analytics are not available in the atproto system');
        lastFetchKey.current = fetchKey;
      } catch (err) {
        console.error('Failed to fetch engagement analytics:', err);
        setError('Failed to load engagement analytics. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [selectedTimeRange, subforumPath]);

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

  // Dark mode chart styling - use a more reliable theme detection method
  const [isDarkMode, setIsDarkMode] = useState(false);
  
  useEffect(() => {
    const checkTheme = () => {
      // Check if the document has the 'dark' class
      const isDark = document.documentElement.classList.contains('dark');
      setIsDarkMode(isDark);
    };

    // Check on mount
    checkTheme();

    // Listen for theme changes
    const observer = new MutationObserver(checkTheme);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });

    return () => observer.disconnect();
  }, []);
  
  const chartColors = {
    posts: isDarkMode ? '#8884d8' : '#8884d8',
    comments: isDarkMode ? '#82ca9d' : '#82ca9d',
    votes: isDarkMode ? '#ffc658' : '#ffc658',
    postVotes: isDarkMode ? '#8884d8' : '#8884d8',
    commentVotes: isDarkMode ? '#82ca9d' : '#82ca9d',
    postUpvotes: isDarkMode ? '#22c55e' : '#22c55e',
    postDownvotes: isDarkMode ? '#ef4444' : '#ef4444',
    commentUpvotes: isDarkMode ? '#8b5cf6' : '#8b5cf6',
    commentDownvotes: isDarkMode ? '#f59e0b' : '#f59e0b',
  };

  const chartStyle = {
    backgroundColor: isDarkMode ? 'transparent' : 'white',
    color: isDarkMode ? '#f5f7fa' : '#0a0a0a',
  };

  const tooltipStyle = {
    backgroundColor: isDarkMode ? 'oklch(0.21 0.006 285.885)' : 'white',
    border: isDarkMode ? '1px solid oklch(1 0 0 / 10%)' : '1px solid #e5e5e5',
    borderRadius: '6px',
    color: isDarkMode ? '#f5f7fa' : '#0a0a0a',
  };

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
      {/* Chart Section */}
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
              <AreaChart 
                data={chartData} 
                style={chartStyle}
                margin={{ left: -10, right: 10, top: 5, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke={isDarkMode ? '#323846' : '#e5e5e5'} />
                <XAxis 
                  dataKey="date" 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <YAxis 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <Tooltip 
                  contentStyle={tooltipStyle}
                  labelStyle={{ color: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                />
                <Legend />
                <Area 
                  type="monotone" 
                  dataKey="posts" 
                  stackId="1" 
                  stroke={chartColors.posts} 
                  fill={chartColors.posts} 
                  name="Posts" 
                />
                <Area 
                  type="monotone" 
                  dataKey="comments" 
                  stackId="1" 
                  stroke={chartColors.comments} 
                  fill={chartColors.comments} 
                  name="Comments" 
                />
                <Area 
                  type="monotone" 
                  dataKey="totalVotes" 
                  stackId="1" 
                  stroke={chartColors.votes} 
                  fill={chartColors.votes} 
                  name="Votes" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Voting Breakdown */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Voting Activity</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart 
                data={chartData} 
                style={chartStyle}
                margin={{ left: -10, right: 10, top: 5, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke={isDarkMode ? '#323846' : '#e5e5e5'} />
                <XAxis 
                  dataKey="date" 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <YAxis 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <Tooltip 
                  contentStyle={tooltipStyle}
                  labelStyle={{ color: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                />
                <Legend />
                <Bar 
                  dataKey="postVotes" 
                  fill={chartColors.postVotes} 
                  name="Post Votes"
                  radius={[2, 2, 0, 0]}
                />
                <Bar 
                  dataKey="commentVotes" 
                  fill={chartColors.commentVotes} 
                  name="Comment Votes"
                  radius={[2, 2, 0, 0]}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Content Creation Trend */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Content Creation Trend</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart 
                data={chartData} 
                style={chartStyle}
                margin={{ left: -10, right: 10, top: 5, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke={isDarkMode ? '#323846' : '#e5e5e5'} />
                <XAxis 
                  dataKey="date" 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <YAxis 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <Tooltip 
                  contentStyle={tooltipStyle}
                  labelStyle={{ color: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                />
                <Legend />
                <Line 
                  type="monotone" 
                  dataKey="posts" 
                  stroke={chartColors.posts} 
                  strokeWidth={2} 
                  name="Posts" 
                />
                <Line 
                  type="monotone" 
                  dataKey="comments" 
                  stroke={chartColors.comments} 
                  strokeWidth={2} 
                  name="Comments" 
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Voting Sentiment */}
          <div>
            <h3 className="text-lg font-semibold mb-4">Voting Sentiment</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData} style={chartStyle}>
                <CartesianGrid strokeDasharray="3 3" stroke={isDarkMode ? '#323846' : '#e5e5e5'} />
                <XAxis 
                  dataKey="date" 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <YAxis 
                  tick={{ fill: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                  axisLine={{ stroke: isDarkMode ? '#323846' : '#e5e5e5' }}
                />
                <Tooltip 
                  contentStyle={tooltipStyle}
                  labelStyle={{ color: isDarkMode ? '#f5f7fa' : '#0a0a0a' }}
                />
                <Legend />
                <Bar 
                  dataKey="postUpvotes" 
                  stackId="posts" 
                  fill={chartColors.postUpvotes} 
                  name="Post Upvotes"
                  radius={[2, 2, 0, 0]}
                />
                <Bar 
                  dataKey="postDownvotes" 
                  stackId="posts" 
                  fill={chartColors.postDownvotes} 
                  name="Post Downvotes"
                  radius={[2, 2, 0, 0]}
                />
                <Bar 
                  dataKey="commentUpvotes" 
                  stackId="comments" 
                  fill={chartColors.commentUpvotes} 
                  name="Comment Upvotes"
                  radius={[2, 2, 0, 0]}
                />
                <Bar 
                  dataKey="commentDownvotes" 
                  stackId="comments" 
                  fill={chartColors.commentDownvotes} 
                  name="Comment Downvotes"
                  radius={[2, 2, 0, 0]}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
} 