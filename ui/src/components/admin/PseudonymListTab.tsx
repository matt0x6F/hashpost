"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import { Users, Eye, Edit, MoreHorizontal } from "lucide-react";
import { getApi } from "@/lib/api-client";
import { AdminApi } from "@/generated/api/src/apis/AdminApi";
import { AdminPseudonymInfo } from "@/generated/api/src/models/AdminPseudonymInfo";
import { toast } from "sonner";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";
import { PseudonymEditDialog } from "@/components/admin/PseudonymEditDialog";

export function PseudonymListTab() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // Pseudonyms state
  const [allPseudonyms, setAllPseudonyms] = useState<AdminPseudonymInfo[]>([]);
  const [isPseudonymsLoading, setIsPseudonymsLoading] = useState(false);
  const [pseudonymsPage, setPseudonymsPage] = useState(1);
  const [pseudonymsTotalPages, setPseudonymsTotalPages] = useState(1);
  const [totalPseudonyms, setTotalPseudonyms] = useState(0);
  
  // Edit dialog state
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedPseudonymId, setSelectedPseudonymId] = useState<string>("");
  
  // Restore state from URL params
  useEffect(() => {
    const pseudonymsPageFromUrl = searchParams.get('pseudonymsPage');
    if (pseudonymsPageFromUrl) {
      const page = parseInt(pseudonymsPageFromUrl);
      if (!isNaN(page) && page > 0) {
        setPseudonymsPage(page);
      }
    }

    // Handle return page from detail pages
    const returnPageFromUrl = searchParams.get('returnPage');
    if (returnPageFromUrl) {
      const page = parseInt(returnPageFromUrl);
      if (!isNaN(page) && page > 0) {
        setPseudonymsPage(page);
      }
    }
  }, [searchParams]);

  // Load pseudonyms when component mounts
  useEffect(() => {
    loadAllPseudonyms();
  }, [pseudonymsPage]);

  const loadAllPseudonyms = async () => {
    setIsPseudonymsLoading(true);
    try {
      const api = getApi(AdminApi);
      const response = await api.adminListPseudonyms(undefined, undefined, pseudonymsPage, DEFAULT_PAGE_SIZE);
      
      if (response.pseudonyms) {
        setAllPseudonyms(response.pseudonyms);
        setTotalPseudonyms(response.pagination?.total || 0);
        setPseudonymsTotalPages(response.pagination?.pages || 1);
      } else {
        setAllPseudonyms([]);
        setTotalPseudonyms(0);
        setPseudonymsTotalPages(1);
      }
    } catch (error: unknown) {
      console.error('Failed to load pseudonyms:', error);
      
      // Handle specific error cases
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
        const status = (error.response as { status: number }).status;
        if (status === 401) {
          toast.error("Authentication required. Please log in again.");
        } else if (status === 403) {
          toast.error("Insufficient permissions for pseudonym management");
        } else if (status === 500) {
          toast.error("Server error. Please try again later.");
        } else {
          toast.error("Failed to load pseudonyms. Please try again.");
        }
      } else {
        toast.error("Failed to load pseudonyms. Please try again.");
      }
      
      setAllPseudonyms([]);
      setTotalPseudonyms(0);
      setPseudonymsTotalPages(1);
    } finally {
      setIsPseudonymsLoading(false);
    }
  };

  const handlePseudonymsPageChange = (newPage: number) => {
    setPseudonymsPage(newPage);
    updatePseudonymsPageURL(newPage);
  };

  const updatePseudonymsPageURL = (page: number = 1) => {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('pseudonymsPage', page.toString());
    router.replace(`/admin?tab=pseudonyms&${newSearchParams.toString()}`);
  };

  const handleEditPseudonym = (pseudonymId: string) => {
    setSelectedPseudonymId(pseudonymId);
    setEditDialogOpen(true);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Pseudonym List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Pseudonym Management
          </CardTitle>
          <CardDescription>
            Manage platform pseudonyms and view details
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isPseudonymsLoading ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground">Loading pseudonyms...</div>
            </div>
          ) : (
            <>
              <div className="mb-4 flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  Total Pseudonyms: {totalPseudonyms}
                </div>
                <div className="text-sm text-muted-foreground">
                  Page {pseudonymsPage} of {pseudonymsTotalPages}
                </div>
              </div>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Pseudonym ID</TableHead>
                    <TableHead>Display Name</TableHead>
                    <TableHead>Slug</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Karma</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {allPseudonyms.map((pseudonym) => (
                    <TableRow key={pseudonym.pseudonymId}>
                      <TableCell>
                        <Badge variant="outline" className="font-mono text-xs">
                          {pseudonym.pseudonymId.substring(0, 8)}...
                        </Badge>
                      </TableCell>
                      <TableCell>{pseudonym.displayName}</TableCell>
                      <TableCell>{pseudonym.slug}</TableCell>
                      <TableCell>
                        {pseudonym.isActive ? (
                          <Badge variant="secondary">Active</Badge>
                        ) : (
                          <Badge variant="outline">Inactive</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{pseudonym.karmaScore}</Badge>
                      </TableCell>
                      <TableCell>{formatDate(pseudonym.createdAt)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button 
                            variant="ghost" 
                            size="sm"
                            onClick={() => router.push(`/admin/pseudonyms/${pseudonym.pseudonymId}?returnTab=pseudonyms&returnPage=${pseudonymsPage}`)}
                            title="View pseudonym details"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button 
                            variant="ghost" 
                            size="sm"
                            onClick={() => handleEditPseudonym(pseudonym.pseudonymId)}
                            title="Edit pseudonym"
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              {pseudonymsTotalPages > 1 && (
                <div className="flex items-center justify-center space-x-2 mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePseudonymsPageChange(pseudonymsPage - 1)}
                    disabled={pseudonymsPage <= 1}
                  >
                    Previous
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {pseudonymsPage} of {pseudonymsTotalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePseudonymsPageChange(pseudonymsPage + 1)}
                    disabled={pseudonymsPage >= pseudonymsTotalPages}
                  >
                    Next
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Edit Pseudonym Dialog */}
      <PseudonymEditDialog
        isOpen={editDialogOpen}
        onClose={() => setEditDialogOpen(false)}
        pseudonymId={selectedPseudonymId}
        onPseudonymUpdated={() => {
          setEditDialogOpen(false);
          loadAllPseudonyms();
        }}
      />
    </div>
  );
}
