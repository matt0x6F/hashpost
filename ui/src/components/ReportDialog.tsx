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
import { Label } from '@/components/shadcn/label';
import { RadioGroup, RadioGroupItem } from '@/components/shadcn/radio-group';
import { 
  AlertTriangle, 
  Flag,
  FileText,
  MessageSquare,
  User
} from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { ModerationApi } from '@/generated/api/src/apis/ModerationApi';

interface ReportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  contentType: 'post' | 'comment' | 'user';
  contentId?: number;
  reportedPseudonymId?: string;
  contentTitle?: string;
  contentPreview?: string;
  reportedUserDisplayName?: string;
}

const REPORT_REASONS = {
  post: [
    { value: 'spam', label: 'Spam or unwanted commercial content' },
    { value: 'harassment', label: 'Harassment or bullying' },
    { value: 'hate_speech', label: 'Hate speech or discrimination' },
    { value: 'violence', label: 'Violence or threats' },
    { value: 'misinformation', label: 'Misinformation or false claims' },
    { value: 'inappropriate', label: 'Inappropriate or offensive content' },
    { value: 'copyright', label: 'Copyright violation' },
    { value: 'other', label: 'Other' }
  ],
  comment: [
    { value: 'spam', label: 'Spam or unwanted commercial content' },
    { value: 'harassment', label: 'Harassment or bullying' },
    { value: 'hate_speech', label: 'Hate speech or discrimination' },
    { value: 'violence', label: 'Violence or threats' },
    { value: 'misinformation', label: 'Misinformation or false claims' },
    { value: 'inappropriate', label: 'Inappropriate or offensive content' },
    { value: 'off_topic', label: 'Off-topic or irrelevant' },
    { value: 'other', label: 'Other' }
  ],
  user: [
    { value: 'harassment', label: 'Harassment or bullying' },
    { value: 'hate_speech', label: 'Hate speech or discrimination' },
    { value: 'spam', label: 'Spam or unwanted commercial content' },
    { value: 'impersonation', label: 'Impersonation or fake identity' },
    { value: 'inappropriate', label: 'Inappropriate behavior' },
    { value: 'other', label: 'Other' }
  ]
};

export function ReportDialog({ 
  open, 
  onOpenChange, 
  contentType, 
  contentId, 
  reportedPseudonymId, 
  contentTitle, 
  contentPreview, 
  reportedUserDisplayName 
}: ReportDialogProps) {
  console.log('ReportDialog props:', { open, contentType, contentId, reportedPseudonymId, contentTitle, contentPreview, reportedUserDisplayName });
  const [reportReason, setReportReason] = useState('');
  const [reportDetails, setReportDetails] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const reasons = REPORT_REASONS[contentType];

  const getContentTypeIcon = (type: string) => {
    switch (type) {
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

  const getContentTypeLabel = (type: string) => {
    switch (type) {
      case 'post':
        return 'Post';
      case 'comment':
        return 'Comment';
      case 'user':
        return 'User';
      default:
        return 'Content';
    }
  };

  const handleSubmit = async () => {
    if (!reportReason) {
      toast.error('Please select a reason for reporting');
      return;
    }

    if (!reportDetails.trim()) {
      toast.error('Please provide details about your report');
      return;
    }

    setIsSubmitting(true);
    try {
      const moderationApi = getApi(ModerationApi);
      
      await moderationApi.reportContent({
        contentType,
        contentId: contentId || null,
        reportedPseudonymId: reportedPseudonymId || '',
        reportReason,
        reportDetails: reportDetails.trim()
      });

      toast.success('Report submitted successfully. Thank you for helping keep our community safe.');
      onOpenChange(false);
      
      // Reset form
      setReportReason('');
      setReportDetails('');
    } catch (error: unknown) {
      console.error('Error submitting report:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to submit report';
      toast.error('Failed to submit report', { description: errorMessage });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    onOpenChange(false);
    setReportReason('');
    setReportDetails('');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {getContentTypeIcon(contentType)}
            Report {getContentTypeLabel(contentType)}
          </DialogTitle>
          <DialogDescription>
            Help us understand what&apos;s wrong with this {contentType}. Your report will be reviewed by our moderation team.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Content Preview */}
          {(contentTitle || contentPreview || reportedUserDisplayName) && (
            <div className="p-4 bg-muted rounded-lg">
              <h4 className="font-medium mb-2">Content being reported:</h4>
              {contentType === 'user' && reportedUserDisplayName ? (
                <div className="flex items-center gap-2">
                  <User className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm">{reportedUserDisplayName}</span>
                </div>
              ) : (
                <div className="space-y-2">
                  {contentTitle && (
                    <div className="font-medium text-sm">{contentTitle}</div>
                  )}
                  {contentPreview && (
                    <div className="text-sm text-muted-foreground line-clamp-3">
                      {contentPreview}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          {/* Report Reason */}
          <div className="space-y-3">
            <Label className="text-base font-medium">What&apos;s wrong with this {contentType}?</Label>
            <RadioGroup value={reportReason} onValueChange={setReportReason}>
              <div className="space-y-3">
                {reasons.map((reason) => (
                  <div key={reason.value} className="flex items-center space-x-2">
                    <RadioGroupItem value={reason.value} id={reason.value} />
                    <Label htmlFor={reason.value} className="text-sm cursor-pointer">
                      {reason.label}
                    </Label>
                  </div>
                ))}
              </div>
            </RadioGroup>
          </div>

          {/* Report Details */}
          <div className="space-y-3">
            <Label htmlFor="report-details" className="text-base font-medium">
              Additional details (required)
            </Label>
            <Textarea
              id="report-details"
              placeholder={`Please provide specific details about why you&apos;re reporting this ${contentType}...`}
              value={reportDetails}
              onChange={(e) => setReportDetails(e.target.value)}
              rows={4}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground">
              The more specific you can be, the better we can address the issue.
            </p>
          </div>

          {/* Report Guidelines */}
          <div className="p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
            <h4 className="font-medium text-blue-900 dark:text-blue-100 mb-2 flex items-center gap-2">
              <Flag className="w-4 h-4" />
              Reporting Guidelines
            </h4>
            <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1">
              <li>• Only report content that violates our community guidelines</li>
              <li>• False reports may result in account restrictions</li>
              <li>• Reports are reviewed by our moderation team</li>
              <li>• You can report anonymously</li>
            </ul>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button 
            onClick={handleSubmit}
            disabled={isSubmitting || !reportReason || !reportDetails.trim()}
          >
            {isSubmitting ? (
              <div className="flex items-center gap-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                Submitting...
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <Flag className="w-4 h-4" />
                Submit Report
              </div>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
} 