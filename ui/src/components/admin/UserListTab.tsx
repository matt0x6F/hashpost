"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import { Users, Eye, ExternalLink } from "lucide-react";
import { getApi } from "@/lib/api-client";
import { PlatformAdminApi } from "@/generated/api/src/apis/PlatformAdminApi";
import { UserWithRoles } from "@/generated/api/src/models/UserWithRoles";
import { toast } from "sonner";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function UserListTab() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // User management state
  const [users, setUsers] = useState<UserWithRoles[]>([]);
  const [isUserListLoading, setIsUserListLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);
  
  // Restore state from URL params
  useEffect(() => {
    const pageFromUrl = searchParams.get('page');
    if (pageFromUrl) {
      const page = parseInt(pageFromUrl);
      if (!isNaN(page) && page > 0) {
        setCurrentPage(page);
      }
    }

    const returnUserPageFromUrl = searchParams.get('returnUserPage');
    if (returnUserPageFromUrl) {
      const page = parseInt(returnUserPageFromUrl);
      if (!isNaN(page) && page > 0) {
        setCurrentPage(page);
      }
    }
  }, [searchParams]);

  // Load users when currentPage changes
  useEffect(() => {
    const loadUsers = async () => {
      setIsUserListLoading(true);
      try {
        const api = getApi(PlatformAdminApi);
        const offset = (currentPage - 1) * DEFAULT_PAGE_SIZE;
        const response = await api.listAllUsers(DEFAULT_PAGE_SIZE, offset);
        
        if (response.users) {
          setUsers(response.users);
          // In atproto system, pagination is handled differently
          // For now, assume we have all users if we get a response
          setTotalPages(1);
          setTotalUsers(response.users.length);
        } else {
          setUsers([]);
          setTotalPages(1);
          setTotalUsers(0);
        }
      } catch (error: unknown) {
        console.error('Failed to load users:', error);
        
        // Handle specific error cases
        if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
          const status = (error.response as { status: number }).status;
          if (status === 401) {
            toast.error("Authentication required. Please log in again.");
          } else if (status === 403) {
            toast.error("Insufficient permissions for user management");
          } else if (status === 500) {
            toast.error("Server error. Please try again later.");
          } else {
            toast.error("Failed to load users. Please try again.");
          }
        } else {
          toast.error("Failed to load users. Please try again.");
        }
        
        setUsers([]);
      } finally {
        setIsUserListLoading(false);
      }
    };

    loadUsers();
  }, [currentPage]);

  const updateUserPageURL = (page: number = 1) => {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('page', page.toString());
    router.replace(`/admin?tab=users&${newSearchParams.toString()}`);
  };

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
    updateUserPageURL(newPage);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* User List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            User Management
          </CardTitle>
          <CardDescription>
            Manage platform users and view account details
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isUserListLoading ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground">Loading users...</div>
            </div>
          ) : (
            <>
              <div className="mb-4 flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  Total Users: {totalUsers}
                </div>
                <div className="text-sm text-muted-foreground">
                  Page {currentPage} of {totalPages}
                </div>
              </div>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>DID</TableHead>
                    <TableHead>Handle</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Email Verified</TableHead>
                    <TableHead>Roles</TableHead>
                    <TableHead>PDS Source</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((user, index) => (
                    <TableRow key={user.userDid || user.handle || index}>
                      <TableCell>
                        <Badge variant="outline">{user.userDid || user.handle || 'N/A'}</Badge>
                      </TableCell>
                      <TableCell>@{user.handle}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">Active</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="default">N/A</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{user.roleCount || 0}</Badge>
                      </TableCell>
                      <TableCell>
                        {user.isLocal ? (
                          <Badge variant="secondary">HashPost PDS</Badge>
                        ) : user.pdsSource ? (
                          <div className="flex items-center gap-2">
                            <Badge variant="outline" className="truncate max-w-xs" title={user.pdsSource}>
                              {user.pdsSource}
                            </Badge>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => window.open(`/admin/pds/${encodeURIComponent(user.pdsSource)}`, '_blank')}
                              title="View PDS details"
                            >
                              <ExternalLink className="h-4 w-4" />
                            </Button>
                          </div>
                        ) : (
                          <Badge variant="outline">Unknown</Badge>
                        )}
                      </TableCell>
                      <TableCell>N/A</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => router.push(`/admin/users/${user.userDid || user.handle}?returnTab=users&returnUserPage=${currentPage}`)}
                            title="View user details"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center space-x-2 mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage <= 1}
                  >
                    Previous
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {currentPage} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage >= totalPages}
                  >
                    Next
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
