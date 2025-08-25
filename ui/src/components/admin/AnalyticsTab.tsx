"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Badge } from "@/components/shadcn/badge";
import { 
  Users, 
  MessageSquare, 
  FileText, 
  TrendingUp, 
  TrendingDown,
  Activity,
  Shield,
  Globe,
  BarChart3,
  PieChart
} from "lucide-react";

export function AnalyticsTab() {
  // Mock data - in a real implementation, this would come from the API
  const stats = {
    totalUsers: 15420,
    totalPseudonyms: 23450,
    totalPosts: 89234,
    totalComments: 456789,
    activeSubforums: 156,
    totalReports: 1234,
    resolvedReports: 1189,
    platformGrowth: 12.5,
    contentGrowth: 8.3,
    userGrowth: 15.2
  };

  const recentActivity = [
    { type: "user_registration", count: 45, change: "+12%", trend: "up" },
    { type: "content_creation", count: 234, change: "+8%", trend: "up" },
    { type: "reports", count: 23, change: "-5%", trend: "down" },
    { type: "moderation_actions", count: 67, change: "+15%", trend: "up" }
  ];

  const topSubforums = [
    { name: "Technology", posts: 12345, users: 3456, growth: "+18%" },
    { name: "Gaming", posts: 9876, users: 2345, growth: "+12%" },
    { name: "Politics", posts: 8765, users: 1987, growth: "+8%" },
    { name: "Science", posts: 6543, users: 1654, growth: "+22%" },
    { name: "Sports", posts: 5432, users: 1432, growth: "+5%" }
  ];

  const getTrendIcon = (trend: string) => {
    if (trend === "up") {
      return <TrendingUp className="h-4 w-4 text-green-600" />;
    }
    return <TrendingDown className="h-4 w-4 text-red-600" />;
  };

  const getTrendColor = (trend: string) => {
    return trend === "up" ? "text-green-600" : "text-red-600";
  };

  return (
    <div className="space-y-6">
      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalUsers.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">
              +{stats.userGrowth}% from last month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Content</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats.totalPosts + stats.totalComments).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              +{stats.contentGrowth}% from last month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Subforums</CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeSubforums}</div>
            <p className="text-xs text-muted-foreground">
              Communities with recent activity
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Platform Growth</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">+{stats.platformGrowth}%</div>
            <p className="text-xs text-muted-foreground">
              Overall platform growth
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Recent Activity
          </CardTitle>
          <CardDescription>
            Platform activity over the last 24 hours
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {recentActivity.map((activity) => (
              <div key={activity.type} className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium capitalize">
                    {activity.type.replace('_', ' ')}
                  </span>
                  {getTrendIcon(activity.trend)}
                </div>
                <div className="text-2xl font-bold">{activity.count}</div>
                <p className={`text-xs ${getTrendColor(activity.trend)}`}>
                  {activity.change} from yesterday
                </p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Content Breakdown */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Content Distribution
            </CardTitle>
            <CardDescription>
              Breakdown of content types across the platform
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span>Posts</span>
                <div className="flex items-center gap-2">
                  <div className="w-32 bg-secondary rounded-full h-2">
                    <div 
                      className="bg-primary h-2 rounded-full" 
                      style={{ width: `${(stats.totalPosts / (stats.totalPosts + stats.totalComments)) * 100}%` }}
                    />
                  </div>
                  <span className="text-sm font-medium">{stats.totalPosts.toLocaleString()}</span>
                </div>
              </div>
              
              <div className="flex items-center justify-between">
                <span>Comments</span>
                <div className="flex items-center gap-2">
                  <div className="w-32 bg-secondary rounded-full h-2">
                    <div 
                      className="bg-primary h-2 rounded-full" 
                      style={{ width: `${(stats.totalComments / (stats.totalPosts + stats.totalComments)) * 100}%` }}
                    />
                  </div>
                  <span className="text-sm font-medium">{stats.totalComments.toLocaleString()}</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PieChart className="h-5 w-5" />
              User Engagement
            </CardTitle>
            <CardDescription>
              User activity and engagement metrics
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span>Total Pseudonyms</span>
                <Badge variant="secondary">{stats.totalPseudonyms.toLocaleString()}</Badge>
              </div>
              
              <div className="flex items-center justify-between">
                <span>Pseudonyms per User</span>
                <Badge variant="outline">
                  {(stats.totalPseudonyms / stats.totalUsers).toFixed(1)}
                </Badge>
              </div>
              
              <div className="flex items-center justify-between">
                <span>Content per User</span>
                <Badge variant="outline">
                  {((stats.totalPosts + stats.totalComments) / stats.totalUsers).toFixed(1)}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Top Subforums */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            Top Subforums by Activity
          </CardTitle>
          <CardDescription>
            Most active communities on the platform
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {topSubforums.map((subforum, index) => (
              <div key={subforum.name} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  <Badge variant="outline">{index + 1}</Badge>
                  <div>
                    <div className="font-medium">{subforum.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {subforum.posts.toLocaleString()} posts • {subforum.users.toLocaleString()} users
                    </div>
                  </div>
                </div>
                <Badge variant="secondary">{subforum.growth}</Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Moderation Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Moderation Overview
          </CardTitle>
          <CardDescription>
            Platform-wide moderation statistics
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="text-center">
              <div className="text-2xl font-bold">{stats.totalReports}</div>
              <p className="text-sm text-muted-foreground">Total Reports</p>
            </div>
            
            <div className="text-center">
              <div className="text-2xl font-bold">{stats.resolvedReports}</div>
              <p className="text-sm text-muted-foreground">Resolved Reports</p>
            </div>
            
            <div className="text-center">
              <div className="text-2xl font-bold">
                {((stats.resolvedReports / stats.totalReports) * 100).toFixed(1)}%
              </div>
              <p className="text-sm text-muted-foreground">Resolution Rate</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
