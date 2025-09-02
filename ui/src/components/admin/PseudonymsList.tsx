"use client";

import { useState, useEffect } from 'react';
import { Button } from '@/components/shadcn/button';
import { Badge } from '@/components/shadcn/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Edit, Eye, User, Calendar, MessageSquare, FileText, Star } from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { AdminApi } from '@/generated/api/src/apis/AdminApi';
import { AdminPseudonymInfo } from '@/generated/api/src/models/AdminPseudonymInfo';
import { PseudonymEditDialog } from "@/components/admin/PseudonymEditDialog";

interface PseudonymsListProps {
  userId: number;
}

export default function PseudonymsList({ userId }: PseudonymsListProps) {
  console.log('PseudonymsList component rendering with userId:', userId);
  
  const [pseudonyms, setPseudonyms] = useState<AdminPseudonymInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedPseudonymId, setSelectedPseudonymId] = useState<string>('');

  const adminApi = getApi(AdminApi);

  useEffect(() => {
    console.log('PseudonymsList useEffect running, userId:', userId);
    if (userId) {
      loadPseudonyms();
    }
  }, [userId]);

  const loadPseudonyms = async () => {
    setLoading(true);
    try {
      const response = await adminApi.adminListPseudonyms();
      setPseudonyms(response.pseudonyms || []);
    } catch (error) {
      console.error('Failed to load pseudonyms:', error);
      toast.error('Failed to load pseudonyms');
    } finally {
      setLoading(false);
    }
  };

  const handleEditPseudonym = (pseudonymId: string) => {
    setSelectedPseudonymId(pseudonymId);
    setEditDialogOpen(true);
  };

  const handlePseudonymUpdated = () => {
    loadPseudonyms();
  };

  const handleCloseEditDialog = () => {
    setEditDialogOpen(false);
    setSelectedPseudonymId('');
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Pseudonyms</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center p-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span className="ml-2">Loading pseudonyms...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Pseudonyms ({pseudonyms.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {pseudonyms.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <User className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No pseudonyms found</p>
            </div>
          ) : (
            <div className="space-y-4">
              {pseudonyms.map((pseudonym) => (
                <div
                  key={pseudonym.pseudonymId}
                  className="border rounded-lg p-4 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h4 className="font-medium text-lg">{pseudonym.displayName}</h4>
                        {pseudonym.slug && (
                          <Badge variant="outline" className="text-xs">
                            @{pseudonym.slug}
                          </Badge>
                        )}
                        <Badge variant={pseudonym.isActive ? "default" : "secondary"}>
                          {pseudonym.isActive ? "Active" : "Inactive"}
                        </Badge>
                      </div>
                      
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Star className="h-4 w-4" />
                          <span>Karma: {pseudonym.karmaScore}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <FileText className="h-4 w-4" />
                          <span>Posts: {pseudonym.postCount || 0}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <MessageSquare className="h-4 w-4" />
                          <span>Comments: {pseudonym.commentCount || 0}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <Calendar className="h-4 w-4" />
                          <span>{new Date(pseudonym.createdAt).toLocaleDateString()}</span>
                        </div>
                      </div>
                      
                      <div className="mt-2">
                        <span className="text-xs text-muted-foreground">
                          Real Identity: {pseudonym.realIdentity}
                        </span>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleEditPseudonym(pseudonym.pseudonymId)}
                      >
                        <Edit className="h-4 w-4 mr-1" />
                        Edit
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <PseudonymEditDialog
        isOpen={editDialogOpen}
        onClose={handleCloseEditDialog}
        pseudonymId={selectedPseudonymId}
        onPseudonymUpdated={handlePseudonymUpdated}
      />
    </>
  );
}
