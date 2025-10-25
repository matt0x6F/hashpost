"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import { Database, Eye, ExternalLink, Activity, Users, Clock } from "lucide-react";
import { getApi } from "@/lib/api-client";
import { PDSManagementApi } from "@/generated/api/src/apis/PDSManagementApi";
import { PDSServerStats } from "@/generated/api/src/models/PDSServerStats";
import { toast } from "sonner";

export function PDSServersTab() {
  const router = useRouter();
  
  // PDS servers state
  const [servers, setServers] = useState<PDSServerStats[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load PDS servers on mount
  useEffect(() => {
    loadPDSServers();
  }, []);

  const loadPDSServers = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const api = getApi(PDSManagementApi);
      const response = await api.listPDSServers();
      setServers(response.servers || []);
    } catch (error: unknown) {
      console.error('Failed to load PDS servers:', error);
      
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
        const status = (error.response as { status: number }).status;
        if (status === 401) {
          setError("Authentication required. Please log in again.");
        } else if (status === 403) {
          setError("Insufficient permissions for PDS management");
        } else if (status === 500) {
          setError("Server error. Please try again later.");
        } else {
          setError("Failed to load PDS servers. Please try again.");
        }
      } else {
        setError("Failed to load PDS servers. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getServerStatusBadge = (server: PDSServerStats) => {
    if (!server.lastActivity) {
      return <Badge variant="secondary">Unknown</Badge>;
    }
    
    const lastActivity = new Date(server.lastActivity);
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

  const handleViewDetails = (endpoint: string) => {
    const encodedEndpoint = encodeURIComponent(endpoint);
    router.push(`/admin/pds/${encodedEndpoint}`);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              PDS Servers
            </CardTitle>
            <CardDescription>
              Monitor connected Personal Data Servers
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8">
              <div className="text-muted-foreground">Loading PDS servers...</div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              PDS Servers
            </CardTitle>
            <CardDescription>
              Monitor connected Personal Data Servers
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8">
              <div className="text-destructive mb-4">{error}</div>
              <Button onClick={loadPDSServers} variant="outline">
                Try Again
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* PDS Servers Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            PDS Servers
          </CardTitle>
          <CardDescription>
            Monitor connected Personal Data Servers and their user activity
          </CardDescription>
        </CardHeader>
        <CardContent>
          {servers.length === 0 ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground mb-4">No PDS servers found</div>
              <div className="text-sm text-muted-foreground">
                PDS servers will appear here as users authenticate from external sources
              </div>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>PDS Endpoint</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Users</TableHead>
                  <TableHead>Active (24h)</TableHead>
                  <TableHead>Last Activity</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {servers.map((server, index) => (
                  <TableRow key={server.pdsEndpoint || index}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Database className="h-4 w-4 text-muted-foreground" />
                        <span className="font-mono text-sm">
                          <span className="truncate max-w-xs" title={server.pdsEndpoint}>
                            {server.pdsEndpoint}
                          </span>
                        </span>
                        <ExternalLink className="h-3 w-3 text-muted-foreground" />
                      </div>
                    </TableCell>
                    <TableCell>
                      {getServerStatusBadge(server)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Users className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{server.userCount}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Activity className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{server.activeUsers24h}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm text-muted-foreground">
                          {formatDate(server.lastActivity)}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleViewDetails(server.pdsEndpoint || '')}
                        title="View PDS details"
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
