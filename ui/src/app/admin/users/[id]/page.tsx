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
  Eye,
  Edit,
  User,
  Mail,
  Calendar,
  MoreHorizontal,
  Shield,
  Database,
  ExternalLink
} from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { PlatformAdminApi } from '@/generated/api/src/apis/PlatformAdminApi';
import { UserWithRoles } from '@/generated/api/src/models/UserWithRoles';
// Updated to use PlatformAdminApi for user listing

export default function UserDetailPage() {
  const router = useRouter();
  const params = useParams();
  const userId = params.id as string;
  
  const [user, setUser] = useState<UserWithRoles | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [isPasswordResetLoading, setIsPasswordResetLoading] = useState(false);
  


  useEffect(() => {
    if (userId) {
      loadUserDetails();
    }
  }, [userId]);

  const loadUserDetails = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const api = getApi(PlatformAdminApi);
      const response = await api.listAllUsers(100, 0);
      const user = response.users?.find(u => u.userDid === userId);
      setUser(user || null);
    } catch (error: unknown) {
      console.error('Failed to load user details:', error);
      
      // Handle specific error cases
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
        const status = (error.response as { status: number }).status;
        if (status === 401) {
          setError("Authentication required. Please log in again.");
        } else if (status === 403) {
          setError("Insufficient permissions for user management");
        } else if (status === 404) {
          setError("User not found");
        } else if (status === 500) {
          setError("Server error. Please try again later.");
        } else {
          setError("Failed to load user details. Please try again.");
        }
      } else {
        setError("Failed to load user details. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleBack = () => {
    router.push('/admin?tab=users');
  };

  const handleTriggerPasswordReset = async () => {
    if (!user) return;
    
    setIsPasswordResetLoading(true);
    try {
      // Password reset not available in atproto system
      toast.error("Password reset is not available in the atproto system");
    } catch (error: unknown) {
      console.error('Failed to trigger password reset:', error);
      toast.error("Failed to trigger password reset. Please try again.");
    } finally {
      setIsPasswordResetLoading(false);
    }
  };

  const handleUserUpdated = (updatedUser: UserWithRoles) => {
    setUser(updatedUser);
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getUserStatusBadge = (user: UserWithRoles) => {
    // In atproto system, users are active by default
    return <Badge variant="secondary">Active</Badge>;
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-6 max-w-7xl">
        <div className="text-center py-8">
          <div className="text-muted-foreground">Loading user details...</div>
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
            Back to User List
          </Button>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container mx-auto py-6 max-w-7xl">
        <div className="text-center py-8">
          <div className="text-muted-foreground mb-4">User not found</div>
          <Button onClick={handleBack} variant="outline">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to User List
          </Button>
        </div>
      </div>
    );
  }



  return (
    <div className="container mx-auto py-6 max-w-7xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Button onClick={handleBack} variant="outline" size="sm">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to User List
        </Button>
        <div>
          <h1 className="text-3xl font-bold">User Details</h1>
          <p className="text-muted-foreground">Manage user account and permissions</p>
        </div>
      </div>

      <div className="grid gap-6">
        {/* User Information Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Account Information
            </CardTitle>
            <CardDescription>
              Basic user account details and status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">DID:</span>
                  <span className="font-mono text-sm">{user.userDid}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Created:</span>
                  <span>N/A</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Status:</span>
                  {getUserStatusBadge(user)}
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Handle:</span>
                  <span className="font-mono text-sm">@{user.handle}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Shield className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Roles:</span>
                  <Badge variant="outline">{user.roleCount}</Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Database className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">PDS Source:</span>
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
                </div>
                {user.lastSeenAt && (
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">Last Seen:</span>
                    <span className="text-sm text-muted-foreground">
                      {formatDate(user.lastSeenAt)}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Actions Card */}
        <Card>
          <CardHeader>
            <CardTitle>User Actions</CardTitle>
            <CardDescription>
              Administrative actions for this user
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              <Button 
                variant="outline" 
                onClick={() => setShowEditDialog(true)}
              >
                <Edit className="h-4 w-4 mr-2" />
                Edit User
              </Button>

              <Button 
                variant="outline" 
                onClick={handleTriggerPasswordReset}
                disabled={isPasswordResetLoading}
              >
                <Shield className="h-4 w-4 mr-2" />
                {isPasswordResetLoading ? "Sending..." : "Trigger Password Reset"}
              </Button>

              <Button variant="outline">
                <Eye className="h-4 w-4 mr-2" />
                View Activity
              </Button>
            </div>
          </CardContent>
        </Card>



        {/* User Edit Dialog not available in atproto system */}
      </div>
    </div>
  );
}
