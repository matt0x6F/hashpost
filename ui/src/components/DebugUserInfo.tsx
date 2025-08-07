"use client";

import React, { useState } from 'react';
import { useAuth } from '@/lib/auth-context';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { ChevronDown, ChevronRight, Bug } from 'lucide-react';

export function DebugUserInfo() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const [isExpanded, setIsExpanded] = useState(false);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated || !user) {
    return <div>Not authenticated</div>;
  }

  const isModerator = user?.capabilities?.includes('moderate_content');

  return (
    <Card className="mb-4">
      <CardHeader className={isExpanded ? "pb-2" : "py-2"}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bug className="w-4 h-4 text-muted-foreground" />
            <CardTitle className="text-sm">Debug Info</CardTitle>
            {!isExpanded && (
              <span className="text-xs text-muted-foreground">
                (Click to expand)
              </span>
            )}
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
            className="h-6 w-6 p-0"
          >
            {isExpanded ? (
              <ChevronDown className="w-4 h-4" />
            ) : (
              <ChevronRight className="w-4 h-4" />
            )}
          </Button>
        </div>
      </CardHeader>
      {isExpanded && (
        <CardContent>
          <div className="space-y-2 text-sm">
            <div><strong>User ID:</strong> {user.userId}</div>
            <div><strong>Email:</strong> {user.email}</div>
            <div><strong>Active Pseudonym:</strong> {user.activePseudonymId}</div>
            <div><strong>Display Name:</strong> {user.displayName}</div>
            <div><strong>Capabilities:</strong> {user.capabilities?.join(', ') || 'None'}</div>
            <div><strong>Is Moderator:</strong> {isModerator ? 'Yes' : 'No'}</div>
            <div><strong>Has moderate_content:</strong> {user.capabilities?.includes('moderate_content') ? 'Yes' : 'No'}</div>
            <div><strong>Has platform_admin:</strong> {user.capabilities?.includes('platform_admin') ? 'Yes' : 'No'}</div>
          </div>
        </CardContent>
      )}
    </Card>
  );
} 