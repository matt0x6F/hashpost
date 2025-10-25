'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Badge } from '@/components/shadcn/badge';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/shadcn/select';
import { Checkbox } from '@/components/shadcn/checkbox';
import { Plus, Edit, Trash2, ArrowLeft, Crown, Shield } from 'lucide-react';
import { DebugUserInfo } from '@/components/DebugUserInfo';
import { COMMUNITY_CONFIG, type CommunityType } from '@/lib/community-config';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import Link from 'next/link';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
// Removed ModeratorTeamMember and ModeratorTeamResponseBody - not available in atproto system
import { ConfirmationDialog } from '@/components/ConfirmationDialog';

// Use placeholder types since moderator team models are not available in atproto system
type ModeratorTeam = any;

export default function ModeratorTeamPage() {
  const params = useParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const communityType = params.community as CommunityType;
  const subforumName = params.subforum as string;
  const fullSubforumPath = `${communityType}/${subforumName}`;

  const communityConfig = COMMUNITY_CONFIG[communityType];
  const { login } = useAuth();
  const [subforumContextLoaded, setSubforumContextLoaded] = useState(false);
  const [moderatorTeam, setModeratorTeam] = useState<ModeratorTeam | null>(null);
  const [loading, setLoading] = useState(true);
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingModerator, setEditingModerator] = useState<string | null>(null);
  const [showRemoveDialog, setShowRemoveDialog] = useState(false);
  const [moderatorToRemove, setModeratorToRemove] = useState<string | null>(null);

  // Form state for adding/editing moderators
  const [formData, setFormData] = useState({
    pseudonym_id: '',
    role: 'moderator',
    capabilities: [] as string[],
  });

  // Available capabilities
  const availableCapabilities = [
    'moderate_content',
    'ban_users',
    'remove_content',
    'review_reports',
    'forward_reports',
    'manage_subforum_rules',
    'manage_subforum_settings',
  ];

  // Available roles
  const availableRoles = [
    { value: 'moderator', label: 'Moderator' },
    { value: 'junior_moderator', label: 'Junior Moderator' },
    { value: 'senior_moderator', label: 'Senior Moderator' },
  ];

  // Load subforum-specific user context
  useEffect(() => {
    if (fullSubforumPath && isAuthenticated) {
      loadSubforumUserContext();
    }
  }, [fullSubforumPath, isAuthenticated]);

  const loadSubforumUserContext = async () => {
    try {
      // In atproto system, capabilities are handled globally via RBAC
      // No need for subforum-specific authentication
      console.log('Subforum context loading not needed in atproto system');
      setSubforumContextLoaded(true);
    } catch (error) {
      console.error('Error loading subforum user context:', error);
      setSubforumContextLoaded(true); // Mark as loaded even on error
    }
  };

  // Load moderator team
  useEffect(() => {
    if (isAuthenticated && subforumContextLoaded) {
      loadModeratorTeam();
    }
  }, [isAuthenticated, subforumContextLoaded]);

  const loadModeratorTeam = async () => {
    try {
      setLoading(true);
      const subforumsApi = getApi(SubforumsApi);
      
      try {
        // Moderator team management not available in atproto system
        setModeratorTeam(null);
      } catch (error: unknown) {
        if (error && typeof error === 'object' && 'status' in error && error.status === 403) {
          toast.error('You do not have permission to view the moderator team');
          router.push(`/${fullSubforumPath}/moderation`);
          return;
        }
        throw new Error('Failed to load moderator team');
      }
    } catch (error) {
      console.error('Error loading moderator team:', error);
      toast.error('Failed to load moderator team');
    } finally {
      setLoading(false);
    }
  };

  const addModerator = async () => {
    try {
      const subforumsApi = getApi(SubforumsApi);
      
      const addModeratorBody = {
        pseudonymId: formData.pseudonym_id,
        role: formData.role,
        capabilities: formData.capabilities
      };
      
      // Moderator team management not available in atproto system
      toast.error('Moderator team management is not available in the atproto system');
      setShowAddForm(false);
      setFormData({ pseudonym_id: '', role: 'moderator', capabilities: [] });
      loadModeratorTeam();
    } catch (error) {
      console.error('Error adding moderator:', error);
      toast.error('Failed to add moderator');
    }
  };

  const updateModerator = async (pseudonymId: string) => {
    try {
      const subforumsApi = getApi(SubforumsApi);
      
      const updateModeratorBody = {
        role: formData.role,
        capabilities: formData.capabilities,
        isActive: true // TODO: Add isActive to formData
      };
      
      // Moderator team management not available in atproto system
      toast.error('Moderator team management is not available in the atproto system');
      setEditingModerator(null);
      setFormData({ pseudonym_id: '', role: 'moderator', capabilities: [] });
      loadModeratorTeam();
    } catch (error) {
      console.error('Error updating moderator:', error);
      toast.error('Failed to update moderator');
    }
  };

  const removeModerator = async (pseudonymId: string) => {
    setModeratorToRemove(pseudonymId);
    setShowRemoveDialog(true);
  };

  const handleRemoveModerator = async () => {
    if (!moderatorToRemove) return;

    try {
      const subforumsApi = getApi(SubforumsApi);
      
      // Moderator team management not available in atproto system
      toast.error('Moderator team management is not available in the atproto system');
      loadModeratorTeam();
    } catch (error) {
      console.error('Error removing moderator:', error);
      toast.error('Failed to remove moderator');
    }
  };

  const startEditing = (moderator: any) => {
    setEditingModerator(moderator.pseudonymId);
    setFormData({
      pseudonym_id: moderator.pseudonymId,
      role: moderator.role,
      capabilities: moderator.capabilities || [],
    });
  };

  const handleCapabilityChange = (capability: string, checked: boolean) => {
    if (checked) {
      setFormData({
        ...formData,
        capabilities: [...formData.capabilities, capability],
      });
    } else {
      setFormData({
        ...formData,
        capabilities: formData.capabilities.filter(c => c !== capability),
      });
    }
  };

  // Check if user has moderator permissions and redirect if not
  useEffect(() => {
    if (!isLoading && isAuthenticated && user && subforumContextLoaded) {
      // In atproto system, permissions are handled via RBAC - for now, assume no permissions
      const hasManageModerators = false;
      
      if (!hasManageModerators) {
        toast.error('You do not have permission to manage the moderator team');
        router.push(`/${fullSubforumPath}`);
      }
    } else if (!isLoading && !isAuthenticated) {
      // Redirect unauthenticated users
      toast.error('You must be logged in to access moderator team management');
      router.push(`/${fullSubforumPath}`);
    }
  }, [user, isAuthenticated, isLoading, subforumContextLoaded, fullSubforumPath, router]);

  // Show loading state while checking permissions or loading subforum context
  if (isLoading || (isAuthenticated && !subforumContextLoaded) || loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading moderator team...</p>
        </div>
      </div>
    );
  }

  // Check if user has moderator permissions
  // In atproto system, permissions are handled via RBAC - for now, assume no permissions
  const hasManageModerators = false;

  // Don't render the page if user doesn't have permissions (redirect will happen)
  if (!isAuthenticated || !hasManageModerators) {
    return null;
  }

  if (!moderatorTeam) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Failed to load moderator team</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Link href={`/${fullSubforumPath}/moderation`}>
            <Button variant="ghost" size="sm">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Moderation
            </Button>
          </Link>
        </div>
        <h1 className="text-3xl font-bold mb-2">Moderator Team</h1>
        <p className="text-muted-foreground">
          Manage the moderator team for{' '}
          <Badge variant="secondary" className={communityConfig.color}>
            {fullSubforumPath}
          </Badge>
          {' '}• Moderation Dashboard
        </p>
      </div>

      {/* Debug Info */}
      <DebugUserInfo />

      <div className="space-y-6">
        {/* Owner */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Crown className="w-5 h-5 text-yellow-600" />
              Owner
            </CardTitle>
            <CardDescription>
              The owner has full control over this subforum
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
              <div>
                <div className="font-medium">{moderatorTeam.owner.displayName || moderatorTeam.owner.pseudonymId}</div>
                <div className="text-sm text-muted-foreground">
                  Added: {new Date(moderatorTeam.owner.addedAt).toLocaleDateString()}
                </div>
              </div>
              <Badge variant="secondary">Owner</Badge>
            </div>
          </CardContent>
        </Card>

        {/* Moderators */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Shield className="w-5 h-5" />
                  Moderators ({moderatorTeam.members?.length || 0})
                </CardTitle>
                <CardDescription>
                  Manage the moderator team and their permissions
                </CardDescription>
              </div>
              <Button onClick={() => setShowAddForm(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Add Moderator
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {!moderatorTeam.members || moderatorTeam.members.length === 0 ? (
              <p className="text-muted-foreground text-center py-8">
                No moderators yet. Add the first moderator to get started.
              </p>
            ) : (
              <div className="space-y-4">
                {moderatorTeam.members.map((moderator: any) => (
                  <div key={moderator.pseudonymId} className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <div className="font-medium">{moderator.displayName || moderator.pseudonymId}</div>
                      <div className="text-sm text-muted-foreground">
                        Role: {moderator.role} • Added: {new Date(moderator.addedAt).toLocaleDateString()}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Capabilities: {moderator.capabilities?.join(', ') || 'None'}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => startEditing(moderator)}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeModerator(moderator.pseudonymId)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Add/Edit Moderator Form */}
        {(showAddForm || editingModerator) && (
          <Card>
            <CardHeader>
              <CardTitle>
                {editingModerator ? 'Edit Moderator' : 'Add New Moderator'}
              </CardTitle>
              <CardDescription>
                Configure the moderator&apos;s role and permissions
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="pseudonym-id">Pseudonym ID</Label>
                <Input
                  id="pseudonym-id"
                  value={formData.pseudonym_id}
                  onChange={(e) =>
                    setFormData({ ...formData, pseudonym_id: e.target.value })
                  }
                  placeholder="Enter pseudonym ID"
                  disabled={!!editingModerator}
                />
              </div>

              <div>
                <Label htmlFor="role">Role</Label>
                <Select
                  value={formData.role}
                  onValueChange={(value) =>
                    setFormData({ ...formData, role: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {availableRoles.map((role) => (
                      <SelectItem key={role.value} value={role.value}>
                        {role.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label>Capabilities</Label>
                <div className="grid grid-cols-2 gap-2 mt-2">
                  {availableCapabilities.map((capability) => (
                    <div key={capability} className="flex items-center space-x-2">
                      <Checkbox
                        id={capability}
                        checked={formData.capabilities.includes(capability)}
                        onCheckedChange={(checked) =>
                          handleCapabilityChange(capability, checked as boolean)
                        }
                      />
                      <Label htmlFor={capability} className="text-sm">
                        {capability.replace(/_/g, ' ')}
                      </Label>
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex gap-2">
                <Button
                  onClick={editingModerator ? () => updateModerator(editingModerator) : addModerator}
                >
                  {editingModerator ? 'Update Moderator' : 'Add Moderator'}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowAddForm(false);
                    setEditingModerator(null);
                    setFormData({ pseudonym_id: '', role: 'moderator', capabilities: [] });
                  }}
                >
                  Cancel
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Confirmation Dialog for Removing Moderators */}
        <ConfirmationDialog
          open={showRemoveDialog}
          onOpenChange={setShowRemoveDialog}
          title="Remove Moderator"
          description="Are you sure you want to remove this moderator? This action cannot be undone."
          confirmText="Remove Moderator"
          cancelText="Cancel"
          onConfirm={handleRemoveModerator}
          variant="destructive"
        />
      </div>
    </div>
  );
} 