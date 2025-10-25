"use client";

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Badge } from '@/components/shadcn/badge';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/shadcn/table';
import { 
  ArrowLeft,
  Database,
  Users,
  Activity,
  Clock,
  Calendar,
  ExternalLink,
  Eye
} from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { PDSManagementApi } from '@/generated/api/src/apis/PDSManagementApi';
import { PDSServerDetails } from '@/generated/api/src/models/PDSServerDetails';
import { UserWithRoles } from '@/generated/api/src/models/UserWithRoles';

export default function PDSDetailPage() {
  const router = useRouter();
  const params = useParams();
  const endpoint = params.endpoint as string;
  
  const [pdsDetails, setPdsDetails] = useState<PDSServerDetails | null>(null);
  const [users, setUsers] = useState<UserWithRoles[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUsersLoading, setIsUsersLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  
  const limit = 20;

  useEffect(() => {
    if (endpoint) {
      loadPDSDetails();
      loadUsers();
    }
  }, [endpoint, currentPage]);

  const loadPDSDetails = async () => {
    try {
      const api = getApi(PDSManagementApi);
      const decodedEndpoint = decodeURIComponent(endpoint);
      const response = await api.getPDSServerDetails(decodedEndpoint);
      setPdsDetails(response);
    } catch (error: unknown) {
      console.error('Failed to load PDS details:', error);
      setError("Failed to load PDS server details");
    }
  };

  const loadUsers = async () => {
    setIsUsersLoading(true);
    try {
      const api = getApi(PDSManagementApi);
      const decodedEndpoint = decodeURIComponent(endpoint);
      const offset = (currentPage - 1) * limit;
      const response = await api.listPDSServerUsers(decodedEndpoint, limit, offset);
      setUsers(response.users || []);
      
      // Calculate total pages (rough estimate)
      const totalUsers = pdsDetails?.userCount || 0;
      setTotalPages(Math.ceil(totalUsers / limit));
    } catch (error: unknown) {
      console.error('Failed to load users:', error);
      toast.error("Failed to load users from PDS server");
    } finally {
      setIsUsersLoading(false);
    }
  };

  const handleBack = () => {
    router.push('/admin?tab=pds');
  };

  const handleUserClick = (userDid: string) => {
    router.push(`/admin/users/${userDid}?returnTab=pds&returnPdsPage=${currentPage}`);
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getServerStatusBadge = (details: PDSServerDetails) => {
    if (!details.lastActivity) {
      return <Badge variant="secondary">Unknown</Badge>;
    }
    
    const lastActivity = new Date(details.lastActivity);
    const now = new Date();
    const hoursSinceActivity = (now.getTime() - lastActivity.getTime()) / (1000 * 60 * 60);
    
    if (hoursSinceActivity < 24) {
      return <Badge variant="default">Active</Badge>;
    } else if (hoursSinceActivity < 168) { // 7 days
      return <Badge variant="secondary">Stale</Badge>;
    } else {
      return <Badge variant="destructive">Inactive</Badge>;
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-6 max-w-7xl">
        <div className="text-center py-8">
          <div className="text-muted-foreground">Loading PDS server details...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-6 max-w-7xl">
        <div className="text-center py-8">
          <div className="text-destructive mb-4">{error}</div>
          <Button onClick={handleBack} variant="outline">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to PDS Servers
          </Button>
        </div>
      </div>
    );
  }

  if (!pdsDetails) {
    return (
      <div className="container mx-auto py-6 max-w-7xl">
        <div className="text-center py-8">
          <div className="text-muted-foreground mb-4">PDS server not found</div>
          <Button onClick={handleBack} variant="outline">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to PDS Servers
          </Button>
        </div>
      </div>
    );
  }

  const decodedEndpoint = decodeURIComponent(endpoint);

  return (
    <div className="container mx-auto py-6 max-w-7xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Button onClick={handleBack} variant="outline" size="sm">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to PDS Servers
        </Button>
        <div>
          <h1 className="text-3xl font-bold">PDS Server Details</h1>
          <p className="text-muted-foreground">Monitor and manage PDS server activity</p>
        </div>
      </div>

      <div className="grid gap-6">
        {/* PDS Server Overview */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Server Overview
            </CardTitle>
            <CardDescription>
              Basic information about this PDS server
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Database className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Endpoint:</span>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm">{decodedEndpoint}</span>
                    {decodedEndpoint !== 'HashPost PDS' && (
                      <ExternalLink className="h-3 w-3 text-muted-foreground" />
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Status:</span>
                  {getServerStatusBadge(pdsDetails)}
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Users className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Total Users:</span>
                  <span className="font-medium">{pdsDetails.userCount}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Last Activity:</span>
                  <span className="text-sm text-muted-foreground">
                    {formatDate(pdsDetails.lastActivity)}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Activity Metrics */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Activity Metrics
            </CardTitle>
            <CardDescription>
              User activity statistics for this PDS server
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{pdsDetails.activeUsers24h}</div>
                <div className="text-sm text-muted-foreground">Active (24h)</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{pdsDetails.activeUsers7d}</div>
                <div className="text-sm text-muted-foreground">Active (7d)</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{pdsDetails.activeUsers30d}</div>
                <div className="text-sm text-muted-foreground">Active (30d)</div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Users from this PDS */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Users from this PDS
            </CardTitle>
            <CardDescription>
              Users authenticated from this PDS server
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isUsersLoading ? (
              <div className="text-center py-8">
                <div className="text-muted-foreground">Loading users...</div>
              </div>
            ) : users.length === 0 ? (
              <div className="text-center py-8">
                <div className="text-muted-foreground">No users found from this PDS server</div>
              </div>
            ) : (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>DID</TableHead>
                      <TableHead>Handle</TableHead>
                      <TableHead>Display Name</TableHead>
                      <TableHead>Roles</TableHead>
                      <TableHead>Last Seen</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.map((user, index) => (
                      <TableRow key={user.userDid || user.handle || index}>
                        <TableCell>
                          <Badge variant="outline" className="font-mono text-xs">
                            {user.userDid || 'N/A'}
                          </Badge>
                        </TableCell>
                        <TableCell>@{user.handle}</TableCell>
                        <TableCell>{user.displayName || 'N/A'}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{user.roleCount || 0}</Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1">
                            <Clock className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm text-muted-foreground">
                              {formatDate(user.lastSeenAt)}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleUserClick(user.userDid)}
                            title="View user details"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4">
                    <div className="text-sm text-muted-foreground">
                      Page {currentPage} of {totalPages}
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                        disabled={currentPage === 1}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                        disabled={currentPage === totalPages}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
