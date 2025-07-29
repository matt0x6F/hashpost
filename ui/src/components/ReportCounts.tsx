'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Shield } from 'lucide-react';
import { getApi } from '@/lib/api-client';
import { ModerationApi } from '@/generated/api/src';
import Link from 'next/link';

interface ReportCountsProps {
  subforumPath: string;
  initialStatus?: string;
}

interface ReportCounts {
  pending: number;
  investigating: number;
  resolved: number;
  dismissed: number;
}

export function ReportCounts({ subforumPath, initialStatus = 'pending' }: ReportCountsProps) {
  const [counts, setCounts] = useState<ReportCounts>({
    pending: 0,
    investigating: 0,
    resolved: 0,
    dismissed: 0
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCounts = async () => {
      setLoading(true);
      setError(null);
      
      try {
        const moderationApi = getApi(ModerationApi);
        
        // Get all reports in one call with a high limit to get all counts
        const allReports = await moderationApi.getSubforumReports(subforumPath, '', '', '', 1, 1000);
        
        // Count reports by status on the frontend
        const counts = {
          pending: 0,
          investigating: 0,
          resolved: 0,
          dismissed: 0
        };
        
        allReports.reports?.forEach(report => {
          switch (report.status) {
            case 'pending':
              counts.pending++;
              break;
            case 'investigating':
              counts.investigating++;
              break;
            case 'resolved':
              counts.resolved++;
              break;
            case 'dismissed':
              counts.dismissed++;
              break;
          }
        });

        setCounts(counts);
      } catch (err) {
        console.error('Failed to fetch report counts:', err);
        setError('Failed to load report counts. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchCounts();
  }, [subforumPath]);

  if (loading) {
    return (
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            Reports Overview
          </CardTitle>
          <CardDescription>
            Manage content and user reports for this subforum
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="text-center p-4 bg-muted/50 rounded-lg">
                <div className="animate-pulse bg-muted h-8 w-16 rounded mx-auto mb-2"></div>
                <div className="animate-pulse bg-muted h-4 w-20 rounded mx-auto"></div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="mb-6">
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
    );
  }

  return (
    <Card className="mb-6">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="w-5 h-5" />
          Reports Overview
        </CardTitle>
        <CardDescription>
          Manage content and user reports for this subforum
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Link href={`/${subforumPath}/moderation/reports?status=pending`}>
            <div className={`text-center p-4 bg-yellow-50 dark:bg-yellow-950/20 rounded-lg hover:bg-yellow-100 dark:hover:bg-yellow-950/40 transition-colors cursor-pointer ${initialStatus === 'pending' ? 'ring-2 ring-yellow-500' : ''}`}>
              <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{counts.pending}</div>
              <div className="text-sm text-yellow-600 dark:text-yellow-400">Pending</div>
            </div>
          </Link>
          <Link href={`/${subforumPath}/moderation/reports?status=investigating`}>
            <div className={`text-center p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-950/40 transition-colors cursor-pointer ${initialStatus === 'investigating' ? 'ring-2 ring-blue-500' : ''}`}>
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{counts.investigating}</div>
              <div className="text-sm text-blue-600 dark:text-blue-400">Investigating</div>
            </div>
          </Link>
          <Link href={`/${subforumPath}/moderation/reports?status=resolved`}>
            <div className={`text-center p-4 bg-green-50 dark:bg-green-950/20 rounded-lg hover:bg-green-100 dark:hover:bg-green-950/40 transition-colors cursor-pointer ${initialStatus === 'resolved' ? 'ring-2 ring-green-500' : ''}`}>
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">{counts.resolved}</div>
              <div className="text-sm text-green-600 dark:text-green-400">Resolved</div>
            </div>
          </Link>
          <Link href={`/${subforumPath}/moderation/reports?status=dismissed`}>
            <div className={`text-center p-4 bg-gray-50 dark:bg-gray-950/20 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-950/40 transition-colors cursor-pointer ${initialStatus === 'dismissed' ? 'ring-2 ring-gray-500' : ''}`}>
              <div className="text-2xl font-bold text-gray-600 dark:text-gray-400">{counts.dismissed}</div>
              <div className="text-sm text-gray-600 dark:text-gray-400">Dismissed</div>
            </div>
          </Link>
        </div>
      </CardContent>
    </Card>
  );
} 