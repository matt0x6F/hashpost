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
  Shield
} from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { AdminApi } from '@/generated/api/src/apis/AdminApi';
import { AdminUserInfo } from '@/generated/api/src/models/AdminUserInfo';

export default function UserDetailPage() {
  const router = useRouter();
  const params = useParams();
  const userId = params.id as string;
  
  const [user, setUser] = useState<AdminUserInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (userId) {
      loadUserDetails();
    }
  }, [userId]);

  const loadUserDetails = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const api = getApi(AdminApi);
      const response = await api.adminGetUser(parseInt(userId));
      setUser(response);
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

  const getUserStatusBadge = (user: AdminUserInfo) => {
    if (user.isSuspended) {
      return <Badge variant="destructive">Suspended</Badge>;
    } else if (user.isActive) {
      return <Badge variant="secondary">Active</Badge>;
    } else {
      return <Badge variant="outline">Inactive</Badge>;
    }
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
                  <Mail className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Email:</span>
                  <span>{user.email}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Created:</span>
                  <span>{formatDate(user.createdAt)}</span>
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
                  <span className="font-medium">Pseudonyms:</span>
                  <Badge variant="outline">{user.pseudonymCount}</Badge>
                </div>
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
              <Button variant="outline">
                <Edit className="h-4 w-4 mr-2" />
                Edit User
              </Button>

              <Button variant="outline">
                <Eye className="h-4 w-4 mr-2" />
                View Activity
              </Button>
              <Button variant="outline">
                <MoreHorizontal className="h-4 w-4 mr-2" />
                More Actions
              </Button>
            </div>
          </CardContent>
        </Card>


      </div>
    </div>
  );
}
