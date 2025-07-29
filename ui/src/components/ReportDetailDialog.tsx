"use client";

import React, { useState } from 'react';
import { 
  Dialog, 
  DialogContent, 
  DialogDescription, 
  DialogFooter, 
  DialogHeader, 
  DialogTitle 
} from '@/components/shadcn/dialog';
import { Button } from '@/components/shadcn/button';
import { Textarea } from '@/components/shadcn/textarea';
import { Badge } from '@/components/shadcn/badge';
import { 
  AlertTriangle, 
  CheckCircle, 
  XCircle, 
  Trash2, 
  Ban, 
  User,
  FileText,
  MessageSquare,
  Clock,
  Eye
} from 'lucide-react';
import { toast } from 'sonner';
import type { Report } from '@/generated/api/src/models/Report';

interface ReportDetailDialogProps {
  report: Report | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAction: (reportId: number, action: 'resolve' | 'dismiss' | 'remove' | 'ban_user' | 'ban_pseudonym' | 'mute_user', notes?: string, muteDuration?: number) => Promise<void>;
}

export function ReportDetailDialog({ report, open, onOpenChange, onAction }: ReportDetailDialogProps) {
  const [action, setAction] = useState<'resolve' | 'dismiss' | 'remove' | 'ban_user' | 'ban_pseudonym' | 'mute_user'>('resolve');
  const [muteDuration, setMuteDuration] = useState<number>(7);
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAction = async () => {
    if (!report) return;

    setLoading(true);
    try {
      const muteDurationParam = action === 'mute_user' ? muteDuration : undefined;
      await onAction(report.reportId, action, notes, muteDurationParam);
      
      const actionMessages = {
        resolve: 'resolved',
        dismiss: 'dismissed', 
        remove: 'content removed',
        ban_user: 'user banned',
        ban_pseudonym: 'pseudonym banned',
        mute_user: `user muted for ${muteDuration} days`
      };
      
      toast.success(`Report ${actionMessages[action]} successfully.`);
      onOpenChange(false);
      setNotes('');
      setAction('resolve');
      setMuteDuration(7);
    } catch (error) {
      console.error('Error handling report action:', error);
      toast.error('Failed to process report action.');
    } finally {
      setLoading(false);
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

  if (!report) return null;

  

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="!w-[95vw] !max-w-none max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {getContentTypeIcon(report.contentType)}
            Report Details
          </DialogTitle>
          <DialogDescription>
            Review and take action on this report
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Report Status */}
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Status:</span>
            {getStatusBadge(report.status)}
          </div>

          {/* Report Information */}
          <div className="space-y-3">
            <div>
              <span className="text-sm font-medium">Reported by:</span>
              <p className="text-sm text-muted-foreground">{report.reporter.displayName}</p>
            </div>

            <div>
              <span className="text-sm font-medium">Reported on:</span>
              <p className="text-sm text-muted-foreground">
                {new Date(report.createdAt).toLocaleString()}
              </p>
            </div>

            <div>
              <span className="text-sm font-medium">Reason:</span>
              <p className="text-sm text-muted-foreground">{report.reportReason}</p>
            </div>

            <div>
              <span className="text-sm font-medium">Details:</span>
              <p className="text-sm text-muted-foreground">{report.reportDetails}</p>
            </div>
          </div>

          {/* Reported Content/User */}
          {report.contentType === 'user' ? (
            <div className="p-3 bg-muted rounded-lg">
              <h4 className="font-medium mb-2">Reported User:</h4>
              <p className="text-sm">{report.reportedUser?.displayName}</p>
            </div>
          ) : report.content ? (
            <div className="p-3 bg-muted rounded-lg">
              <h4 className="font-medium mb-2">Reported Content:</h4>
              {report.content.title && (
                <p className="text-sm font-medium mb-1">{report.content.title}</p>
              )}
              <p className="text-sm">{report.content.content}</p>
            </div>
          ) : null}

          {/* Resolution Information */}
          {report.resolvedBy && (
            <div className="p-3 bg-green-50 dark:bg-green-950/20 rounded-lg">
              <h4 className="font-medium mb-2 text-green-700 dark:text-green-300">
                Resolved by {report.resolvedBy.displayName}
              </h4>
              {report.resolvedAt && (
                <p className="text-sm text-green-600 dark:text-green-400">
                  {new Date(report.resolvedAt).toLocaleString()}
                </p>
              )}
              {report.resolutionNotes && (
                <p className="text-sm text-green-600 dark:text-green-400 mt-2">
                  {report.resolutionNotes}
                </p>
              )}
            </div>
          )}

          {/* Action Section - Only show for pending reports */}
          {report.status === 'pending' && (
            <div className="space-y-4">
              <div>
                <span className="text-sm font-medium">Action:</span>
                <div className="flex gap-2 mt-2">
                  <Button
                    variant={action === 'resolve' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setAction('resolve')}
                  >
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Resolve
                  </Button>
                  <Button
                    variant={action === 'dismiss' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setAction('dismiss')}
                  >
                    <XCircle className="w-4 h-4 mr-2" />
                    Dismiss
                  </Button>
                  {report.contentId && (
                    <Button
                      variant={action === 'remove' ? 'default' : 'outline'}
                      size="sm"
                      onClick={() => setAction('remove')}
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      Remove Content
                    </Button>
                  )}
                  {report.reportedPseudonymId && (
                    <>
                      <Button
                        variant={action === 'ban_user' ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setAction('ban_user')}
                      >
                        <Ban className="w-4 h-4 mr-2" />
                        Ban User
                      </Button>
                      <Button
                        variant={action === 'ban_pseudonym' ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setAction('ban_pseudonym')}
                      >
                        <User className="w-4 h-4 mr-2" />
                        Ban Pseudonym
                      </Button>
                      <Button
                        variant={action === 'mute_user' ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setAction('mute_user')}
                      >
                        <Clock className="w-4 h-4 mr-2" />
                        Mute User
                      </Button>
                    </>
                  )}
                </div>
              </div>

              <div>
                <span className="text-sm font-medium">Notes (optional):</span>
                <Textarea
                  placeholder="Add notes about your decision..."
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="mt-2"
                  rows={3}
                />
              </div>

              {action === 'mute_user' && (
                <div>
                  <span className="text-sm font-medium">Mute Duration (days):</span>
                  <div className="flex gap-2 mt-2">
                    <input
                      type="number"
                      min="1"
                      max="365"
                      value={muteDuration}
                      onChange={(e) => setMuteDuration(parseInt(e.target.value) || 7)}
                      className="flex h-9 w-20 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                    />
                    <span className="text-sm text-muted-foreground self-center">days</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {report.status === 'pending' && (
            <Button 
              onClick={handleAction}
              disabled={loading}
            >
              {loading ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              ) : null}
              {action === 'resolve' && 'Resolve Report'}
              {action === 'dismiss' && 'Dismiss Report'}
              {action === 'remove' && 'Remove Content'}
              {action === 'ban_user' && 'Ban User'}
              {action === 'ban_pseudonym' && 'Ban Pseudonym'}
              {action === 'mute_user' && `Mute User (${muteDuration} days)`}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
} 