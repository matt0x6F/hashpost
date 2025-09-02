"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import { Label } from "@/components/shadcn/label";
import { Users, Search, Eye } from "lucide-react";
import { getApi } from "@/lib/api-client";
import { AdminApi } from "@/generated/api/src/apis/AdminApi";
import { AdminUserInfo } from "@/generated/api/src/models/AdminUserInfo";
import { toast } from "sonner";
import { DEFAULT_PAGE_SIZE } from "@/lib/constants";

export function UserListTab() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // User management state
  const [users, setUsers] = useState<AdminUserInfo[]>([]);
  const [isUserListLoading, setIsUserListLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);
  const [userSearchQuery, setUserSearchQuery] = useState("");
  
  // Restore state from URL params
  useEffect(() => {
    const userQueryFromUrl = searchParams.get('userQuery');
    if (userQueryFromUrl && userQueryFromUrl !== userSearchQuery) {
      setUserSearchQuery(userQueryFromUrl);
    }

    const returnUserQueryFromUrl = searchParams.get('returnUserQuery');
    if (returnUserQueryFromUrl && returnUserQueryFromUrl !== userSearchQuery) {
      setUserSearchQuery(returnUserQueryFromUrl);
    }

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
  }, [searchParams, userSearchQuery]);

  // Load users on component mount and when page changes
  useEffect(() => {
    loadUsers();
  }, [currentPage]);

  const loadUsers = async () => {
    setIsUserListLoading(true);
    try {
      const api = getApi(AdminApi);
      const response = await api.adminListUsers(undefined, undefined, currentPage, DEFAULT_PAGE_SIZE, userSearchQuery || undefined);
      
      if (response.users) {
        setUsers(response.users);
        setTotalPages(response.pagination?.pages || 1);
        setTotalUsers(response.pagination?.total || 0);
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

  const handleUserSearch = async () => {
    setCurrentPage(1); // Reset to first page when searching
    updateUserSearchURL(userSearchQuery, 1);
    await loadUsers();
  };

  const handleUserSearchClear = () => {
    setUserSearchQuery("");
    setCurrentPage(1);
    updateUserSearchURL("", 1);
    loadUsers();
  };

  const updateUserSearchURL = (query: string, page: number = 1) => {
    const newSearchParams = new URLSearchParams(searchParams);
    if (query.trim()) {
      newSearchParams.set('userQuery', query);
    } else {
      newSearchParams.delete('userQuery');
    }
    newSearchParams.set('page', page.toString());
    router.replace(`/admin?tab=users&${newSearchParams.toString()}`);
  };

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
    updateUserSearchURL(userSearchQuery, newPage);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className="space-y-6">
      {/* Search Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Search Users
          </CardTitle>
          <CardDescription>
            Search for users by ID or email address
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={(e) => { e.preventDefault(); handleUserSearch(); }} className="flex gap-2">
            <div className="space-y-4 w-full">
              <div>
                <Label htmlFor="user-search" className="text-sm font-medium">
                  Search for users by ID or email address
                </Label>
                <div className="mt-2 flex gap-2">
                  <Input
                    id="user-search"
                    placeholder="Search by ID or email..."
                    value={userSearchQuery}
                    onChange={(e) => setUserSearchQuery(e.target.value)}
                    className="flex-1"
                  />
                  <Button type="submit" disabled={isUserListLoading || !userSearchQuery.trim()}>
                    {isUserListLoading ? "Searching..." : "Search"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleUserSearchClear}
                  >
                    Show All
                  </Button>
                </div>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* User List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            User Management
          </CardTitle>
          <CardDescription>
            {userSearchQuery.trim() ? `Search results for "${userSearchQuery}"` : 'Manage platform users and view account details'}
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
                  {userSearchQuery.trim() ? `Found ${users.length} users` : `Total Users: ${totalUsers}`}
                </div>
                <div className="text-sm text-muted-foreground">
                  Page {currentPage} of {totalPages}
                </div>
              </div>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User ID</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Pseudonyms</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((user) => (
                    <TableRow key={user.userId}>
                      <TableCell>
                        <Badge variant="outline">{user.userId}</Badge>
                      </TableCell>
                      <TableCell>{user.email}</TableCell>
                      <TableCell>
                        {user.isActive ? (
                          <Badge variant="secondary">Active</Badge>
                        ) : (
                          <Badge variant="outline">Inactive</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{user.pseudonymCount}</Badge>
                      </TableCell>
                      <TableCell>{formatDate(user.createdAt)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => router.push(`/admin/users/${user.userId}?returnTab=users&returnUserQuery=${encodeURIComponent(userSearchQuery)}&returnUserPage=${currentPage}`)}
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
