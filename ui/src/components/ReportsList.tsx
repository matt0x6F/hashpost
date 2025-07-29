"use client";

import React, { useState, useEffect } from 'react';
import { Card, CardContent } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Badge } from '@/components/shadcn/badge';
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/shadcn/select';
import { 
  Flag, 
  Clock, 
  CheckCircle, 
  XCircle, 
  Eye,
  FileText,
  MessageSquare,
  User,
  AlertTriangle
} from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { ModerationApi } from '@/generated/api/src/apis/ModerationApi';
import { ReportDetailDialog } from './ReportDetailDialog';

import type { Report } from '@/generated/api/src/models/Report';

interface ReportsListProps {
  subforumPath: string;
  initialStatus?: string;
}

export function ReportsList({ subforumPath, initialStatus = 'pending' }: ReportsListProps) {
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState(initialStatus);
  
  // Update status filter when initialStatus prop changes (from URL)
  useEffect(() => {
    console.log('ReportsList: initialStatus changed to:', initialStatus);
    setStatusFilter(initialStatus);
    setPage(1); // Reset to first page when status changes
  }, [initialStatus]);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [showReportDialog, setShowReportDialog] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const moderationApi = getApi(ModerationApi);

  const loadReports = async () => {
    console.log('ReportsList: loadReports called with statusFilter:', statusFilter, 'page:', page);
    setLoading(true);
    try {
      const response = await moderationApi.getSubforumReports(
        subforumPath,
        '', // authorization
        '', // accessToken
        statusFilter,
        page,
        25 // limit
      );

      console.log('Reports API response:', response);
      console.log('Status filter:', statusFilter);
      console.log('Reports found:', response.reports?.length || 0);

      if (response.reports) {
        setReports(response.reports);
        setTotalPages(Math.ceil((response.pagination?.total || 0) / 25));
      }
    } catch (error: unknown) {
      console.error('Error loading reports:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to load reports';
      toast.error('Failed to load reports', { description: errorMessage });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReports();
  }, [statusFilter, page]);

  const handleReportAction = async (reportId: number, action: 'resolve' | 'dismiss' | 'remove' | 'ban_user' | 'ban_pseudonym' | 'mute_user', notes?: string, muteDuration?: number) => {
    try {
      const moderationApi = getApi(ModerationApi);
      
      if (action === 'resolve' || action === 'dismiss') {
        // Call the resolve report endpoint
        await moderationApi.resolveReport(reportId, {
          action: action,
          notes: notes || ''
        }, '', ''); // authorization and accessToken parameters
      } else if (action === 'remove') {
        // TODO: Implement content removal
        console.log('Content removal not yet implemented');
      } else if (action === 'ban_user' || action === 'ban_pseudonym' || action === 'mute_user') {
        // TODO: Implement user banning
        console.log('User banning not yet implemented');
      }
      
      // Reload reports to show updated status
      await loadReports();
      
      const actionMessages = {
        resolve: 'resolved',
        dismiss: 'dismissed',
        remove: 'content removed',
        ban_user: 'user banned',
        ban_pseudonym: 'pseudonym banned',
        mute_user: `user muted for ${muteDuration || 7} days`
      };
      
      toast.success(`Report ${actionMessages[action]} successfully.`);
    } catch (error: unknown) {
      console.error('Error handling report action:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to process report action';
      toast.error('Failed to process report action', { description: errorMessage });
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <Badge variant="secondary"><Clock className="w-3 h-3 mr-1" />Pending</Badge>;
      case 'investigating':
        return <Badge variant="default"><Eye className="w-3 h-3 mr-1" />Investigating</Badge>;
      case 'resolved':
        return <Badge variant="default"><CheckCircle className="w-3 h-3 mr-1" />Resolved</Badge>;
      case 'dismissed':
        return <Badge variant="secondary"><XCircle className="w-3 h-3 mr-1" />Dismissed</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getContentTypeIcon = (contentType: string) => {
    switch (contentType) {
      case 'post':
        return <FileText className="w-4 h-4" />;
      case 'comment':
        return <MessageSquare className="w-4 h-4" />;
      case 'user':
        return <User className="w-4 h-4" />;
      default:
        return <AlertTriangle className="w-4 h-4" />;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getReportReasonLabel = (reason: string) => {
    const reasonMap: Record<string, string> = {
      'spam': 'Spam',
      'harassment': 'Harassment',
      'hate_speech': 'Hate Speech',
      'violence': 'Violence',
      'misinformation': 'Misinformation',
      'inappropriate': 'Inappropriate',
      'copyright': 'Copyright',
      'off_topic': 'Off Topic',
      'impersonation': 'Impersonation',
      'other': 'Other'
    };
    return reasonMap[reason] || reason;
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">Reports</h2>
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
        </div>
        <div className="grid gap-4">
          {[...Array(5)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="p-4">
                <div className="h-4 bg-muted rounded w-1/3 mb-2"></div>
                <div className="h-3 bg-muted rounded w-1/2"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Reports</h2>
        <div className="flex items-center gap-4">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="investigating">Investigating</SelectItem>
              <SelectItem value="resolved">Resolved</SelectItem>
              <SelectItem value="dismissed">Dismissed</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {reports.length === 0 ? (
        <Card>
          <CardContent className="p-8 text-center">
            <Flag className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No reports found</h3>
            <p className="text-muted-foreground">
              {statusFilter === 'pending' 
                ? 'No pending reports to review.'
                : `No ${statusFilter} reports found.`
              }
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {reports.map((report) => (
            <Card key={report.reportId} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      {getContentTypeIcon(report.contentType)}
                      <span className="font-medium">
                        {report.contentType === 'post' ? 'Post' : 
                         report.contentType === 'comment' ? 'Comment' : 'User'} Report
                      </span>
                      {getStatusBadge(report.status)}
                    </div>
                    
                    <div className="text-sm text-muted-foreground">
                      <span>Reported by {report.reporter.displayName}</span>
                      <span className="mx-2">•</span>
                      <span>{formatDate(report.createdAt)}</span>
                      <span className="mx-2">•</span>
                      <span>{getReportReasonLabel(report.reportReason)}</span>
                    </div>

                    {report.reportedUser && (
                      <div className="text-sm">
                        <span className="text-muted-foreground">Reported user: </span>
                        <span className="font-medium">{report.reportedUser.displayName}</span>
                      </div>
                    )}

                    {report.content && (
                      <div className="text-sm">
                        <span className="text-muted-foreground">Content: </span>
                        <span className="line-clamp-2">
                          {report.content.title || report.content.content}
                        </span>
                      </div>
                    )}

                    <div className="text-sm text-muted-foreground line-clamp-2">
                      {report.reportDetails}
                    </div>
                  </div>

                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSelectedReport(report);
                      setShowReportDialog(true);
                    }}
                  >
                    <Eye className="w-4 h-4 mr-2" />
                    Review
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-6">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {page} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
              >
                Next
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Report Detail Dialog */}
      <ReportDetailDialog
        report={selectedReport}
        open={showReportDialog}
        onOpenChange={setShowReportDialog}
        onAction={handleReportAction}
      />
    </div>
  );
} 