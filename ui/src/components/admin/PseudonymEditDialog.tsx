import React, { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/shadcn/dialog';
import { Button } from '@/components/shadcn/button';
import { Input } from '@/components/shadcn/input';
import { Label } from '@/components/shadcn/label';
import { Textarea } from '@/components/shadcn/textarea';
import { Switch } from '@/components/shadcn/switch';
import { Badge } from '@/components/shadcn/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/shadcn/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/shadcn/tabs';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Separator } from '@/components/shadcn/separator';
import { X, Save, Loader2, Plus, Globe, Building2, Shield, Key, Users, Settings } from 'lucide-react';
import { toast } from 'sonner';
import { getApi } from '@/lib/api-client';
import { AdminApi } from '@/generated/api/src/apis/AdminApi';
import { AdminPseudonymDetail } from '@/generated/api/src/models/AdminPseudonymDetail';

interface RoleKey {
  keyId: string;
  roleName: string;
  scope: string;
  capabilities: string[];
  subforumId?: number;
  subforumName?: string;
  expiresAt: string;
  isActive: boolean;
}

interface PseudonymEditDialogProps {
  isOpen: boolean;
  onClose: () => void;
  pseudonymId: string;
  onPseudonymUpdated: () => void;
}

export function PseudonymEditDialog({
  isOpen,
  onClose,
  pseudonymId,
  onPseudonymUpdated,
}: PseudonymEditDialogProps) {
  const [pseudonym, setPseudonym] = useState<AdminPseudonymDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState({
    displayName: '',
    slug: '',
    bio: '',
    websiteURL: '',
    isActive: true,
    showKarma: true,
    allowDirectMessages: true,
  });

  // Role keys management
  const [roleKeys, setRoleKeys] = useState<RoleKey[]>([]);
  const [newRoleKey, setNewRoleKey] = useState({
    roleName: '',
    scope: '',
    capabilities: [] as string[],
    subforumId: undefined as number | undefined,
    expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // 1 year from now
  });

  const adminApi = getApi(AdminApi);

  // Available roles and scopes for the form
  const availableRoles = [
    { value: 'user', label: 'User', description: 'Basic platform user' },
    { value: 'moderator', label: 'Moderator', description: 'Subforum content moderator' },
    { value: 'subforum_owner', label: 'Subforum Owner', description: 'Subforum administrator' },
    { value: 'trust_safety', label: 'Trust & Safety', description: 'Platform safety team' },
    { value: 'legal_team', label: 'Legal Team', description: 'Legal compliance team' },
    { value: 'platform_admin', label: 'Platform Admin', description: 'System administrator' },
  ];

  const availableScopes = [
    { value: 'authentication', label: 'Authentication', description: 'Login and session management', icon: Shield },
    { value: 'self_correlation', label: 'Self Correlation', description: 'Own pseudonym access', icon: Key },
    { value: 'moderation', label: 'Moderation', description: 'Content and user moderation', icon: Users },
    { value: 'correlation', label: 'Correlation', description: 'Cross-user identity correlation', icon: Building2 },
    { value: 'administration', label: 'Administration', description: 'System administration', icon: Settings },
  ];

  const availableCapabilities = [
    // Basic user capabilities
    { value: 'create_content', label: 'Create Content', description: 'Create posts and comments' },
    { value: 'vote', label: 'Vote', description: 'Vote on content' },
    { value: 'message', label: 'Message', description: 'Send direct messages' },
    { value: 'report', label: 'Report', description: 'Report content or users' },
    
    // Moderation capabilities
    { value: 'moderate_content', label: 'Moderate Content', description: 'Moderate posts and comments' },
    { value: 'ban_users', label: 'Ban Users', description: 'Ban users from subforum' },
    { value: 'remove_content', label: 'Remove Content', description: 'Remove posts and comments' },
    { value: 'review_reports', label: 'Review Reports', description: 'Review user reports' },
    { value: 'manage_subforum_settings', label: 'Manage Settings', description: 'Manage subforum settings' },
    { value: 'manage_subforum_rules', label: 'Manage Rules', description: 'Manage subforum rules' },
    
    // Administrative capabilities
    { value: 'system_admin', label: 'System Admin', description: 'Full system administration' },
    { value: 'user_management', label: 'User Management', description: 'Manage user accounts' },
    { value: 'correlate_identities', label: 'Correlate Identities', description: 'Cross-platform identity correlation' },
    { value: 'access_all_pseudonyms', label: 'Access All Pseudonyms', description: 'Access all user pseudonyms' },
  ];

  useEffect(() => {
    console.log('PseudonymEditDialog useEffect - isOpen:', isOpen, 'pseudonymId:', pseudonymId);
    if (isOpen && pseudonymId) {
      console.log('Loading pseudonym data...');
      loadPseudonym();
    }
  }, [isOpen, pseudonymId]);

  const loadPseudonym = async () => {
    setLoading(true);
    try {
      const response = await adminApi.adminGetPseudonym(pseudonymId);
      const pseudonymData = response;
      setPseudonym(pseudonymData);
      setFormData({
        displayName: pseudonymData.displayName || '',
        slug: pseudonymData.slug || '',
        bio: pseudonymData.bio || '',
        websiteURL: pseudonymData.websiteUrl || '',
        isActive: pseudonymData.isActive,
        showKarma: pseudonymData.showKarma,
        allowDirectMessages: pseudonymData.allowDirectMessages,
      });

      // Load role keys from the backend response
      if (pseudonymData.roleKeys && pseudonymData.roleKeys.length > 0) {
        const realRoleKeys: RoleKey[] = pseudonymData.roleKeys.map(roleKey => ({
          keyId: roleKey.keyId || '',
          roleName: roleKey.roleName,
          scope: roleKey.scope,
          capabilities: roleKey.capabilities || [],
          subforumId: roleKey.subforumId,
          subforumName: roleKey.subforumName || '',
          expiresAt: roleKey.expiresAt,
          isActive: roleKey.isActive,
        }));
        setRoleKeys(realRoleKeys);
      } else {
        // Fallback to empty array if no role keys
        setRoleKeys([]);
      }
    } catch (error) {
      console.error('Failed to load pseudonym:', error);
      toast.error('Failed to load pseudonym details');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      // Convert role keys to the format expected by the backend
      const roleKeysPayload = roleKeys.map(roleKey => ({
        roleName: roleKey.roleName,
        scope: roleKey.scope,
        capabilities: roleKey.capabilities,
        subforumId: roleKey.subforumId,
        expiresAt: roleKey.expiresAt,
        isActive: roleKey.isActive,
      }));

      await adminApi.adminUpdatePseudonym(pseudonymId, {
        displayName: formData.displayName || undefined,
        slug: formData.slug || undefined,
        bio: formData.bio || undefined,
        websiteUrl: formData.websiteURL || undefined,
        isActive: formData.isActive,
        showKarma: formData.showKarma,
        allowDirectMessages: formData.allowDirectMessages,
        roleKeys: roleKeysPayload.length > 0 ? roleKeysPayload : undefined,
      });

      toast.success('Pseudonym updated successfully');
      onPseudonymUpdated();
      onClose();
    } catch (error) {
      console.error('Failed to update pseudonym:', error);
      toast.error('Failed to update pseudonym');
    } finally {
      setSaving(false);
    }
  };

  const handleInputChange = (field: string, value: string | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const addRoleKey = () => {
    if (!newRoleKey.roleName || !newRoleKey.scope || newRoleKey.capabilities.length === 0) {
      toast.error('Please fill in all required fields for the role key');
      return;
    }

    const roleKey: RoleKey = {
      keyId: `new-${Date.now()}`,
      roleName: newRoleKey.roleName,
      scope: newRoleKey.scope,
      capabilities: [...newRoleKey.capabilities],
      subforumId: newRoleKey.subforumId,
      subforumName: newRoleKey.subforumId ? `Subforum ${newRoleKey.subforumId}` : undefined,
      expiresAt: new Date(newRoleKey.expiresAt).toISOString(),
      isActive: true,
    };

    setRoleKeys(prev => [...prev, roleKey]);
    
    // Reset form
    setNewRoleKey({
      roleName: '',
      scope: '',
      capabilities: [],
      subforumId: undefined,
      expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    });

    toast.success('Role key added (changes will be saved when you click Save)');
  };

  const removeRoleKey = (keyId: string) => {
    setRoleKeys(prev => prev.filter(rk => rk.keyId !== keyId));
    toast.success('Role key removed (changes will be saved when you click Save)');
  };

  const toggleCapability = (capability: string) => {
    setNewRoleKey(prev => ({
      ...prev,
      capabilities: prev.capabilities.includes(capability)
        ? prev.capabilities.filter(c => c !== capability)
        : [...prev.capabilities, capability]
    }));
  };

  const getScopeIcon = (scope: string) => {
    const scopeInfo = availableScopes.find(s => s.value === scope);
    return scopeInfo?.icon || Globe;
  };

  const getRoleLabel = (roleName: string) => {
    const role = availableRoles.find(r => r.value === roleName);
    return role?.label || roleName;
  };

  const getCapabilityLabel = (capability: string) => {
    const cap = availableCapabilities.find(c => c.value === capability);
    return cap?.label || capability;
  };

  if (loading) {
    return (
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="sm:max-w-[800px]">
          <DialogHeader>
            <DialogTitle>Loading Pseudonym</DialogTitle>
          </DialogHeader>
          <div className="flex items-center justify-center p-8">
            <Loader2 className="h-8 w-8 animate-spin" />
            <span className="ml-2">Loading pseudonym details...</span>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[900px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Pseudonym</DialogTitle>
          <DialogDescription>
            Modify pseudonym settings and manage role-based permissions
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="profile" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="profile">Profile</TabsTrigger>
            <TabsTrigger value="permissions">Permissions</TabsTrigger>
            <TabsTrigger value="statistics">Statistics</TabsTrigger>
          </TabsList>

          <TabsContent value="profile" className="space-y-6">
            {/* Basic Information */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Basic Information</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="displayName">Display Name</Label>
                  <Input
                    id="displayName"
                    value={formData.displayName}
                    onChange={(e) => handleInputChange('displayName', e.target.value)}
                    placeholder="Display name"
                  />
                </div>
                
                <div>
                  <Label htmlFor="slug">Slug</Label>
                  <Input
                    id="slug"
                    value={formData.slug}
                    onChange={(e) => handleInputChange('slug', e.target.value)}
                    placeholder="URL slug"
                  />
                </div>
              </div>

              <div>
                <Label htmlFor="bio">Bio</Label>
                <Textarea
                  id="bio"
                  value={formData.bio}
                  onChange={(e) => handleInputChange('bio', e.target.value)}
                  placeholder="User bio"
                  rows={3}
                />
              </div>

              <div>
                <Label htmlFor="websiteURL">Website URL</Label>
                <Input
                  id="websiteURL"
                  value={formData.websiteURL}
                  onChange={(e) => handleInputChange('websiteURL', e.target.value)}
                  placeholder="https://example.com"
                />
              </div>
            </div>

            {/* Settings */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Settings</h3>
              
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label htmlFor="isActive">Active</Label>
                  <Switch
                    id="isActive"
                    checked={formData.isActive}
                    onCheckedChange={(checked) => handleInputChange('isActive', checked)}
                  />
                </div>
                
                <div className="flex items-center justify-between">
                  <Label htmlFor="showKarma">Show Karma</Label>
                  <Switch
                    id="showKarma"
                    checked={formData.showKarma}
                    onCheckedChange={(checked) => handleInputChange('showKarma', checked)}
                  />
                </div>
                
                <div className="flex items-center justify-between">
                  <Label htmlFor="allowDirectMessages">Allow Direct Messages</Label>
                  <Switch
                    id="allowDirectMessages"
                    checked={formData.allowDirectMessages}
                    onCheckedChange={(checked) => handleInputChange('allowDirectMessages', checked)}
                  />
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="permissions" className="space-y-6">
            {/* Current Role Keys */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Current Role Keys</h3>
              
              {roleKeys.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No role keys assigned</p>
                  <p className="text-sm">Add role keys to grant specific permissions</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {roleKeys.map((roleKey) => (
                    <Card key={roleKey.keyId}>
                      <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            {React.createElement(getScopeIcon(roleKey.scope), { className: "h-5 w-5" })}
                            <div>
                              <CardTitle className="text-base">
                                {getRoleLabel(roleKey.roleName)} • {roleKey.scope}
                              </CardTitle>
                              <CardDescription>
                                {roleKey.subforumId ? `Subforum: ${roleKey.subforumName}` : 'Global scope'}
                              </CardDescription>
                            </div>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => removeRoleKey(roleKey.keyId)}
                            className="text-destructive hover:text-destructive"
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        </div>
                      </CardHeader>
                      <CardContent>
                        <div className="flex flex-wrap gap-2">
                          {roleKey.capabilities.map((capability) => (
                            <Badge key={capability} variant="secondary">
                              {getCapabilityLabel(capability)}
                            </Badge>
                          ))}
                        </div>
                        <div className="mt-3 text-sm text-muted-foreground">
                          Expires: {new Date(roleKey.expiresAt).toLocaleDateString()}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </div>

            <Separator />

            {/* Add New Role Key */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Add New Role Key</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="roleName">Role</Label>
                  <Select value={newRoleKey.roleName} onValueChange={(value) => setNewRoleKey(prev => ({ ...prev, roleName: value }))}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableRoles.map((role) => (
                        <SelectItem key={role.value} value={role.value}>
                          <div>
                            <div className="font-medium">{role.label}</div>
                            <div className="text-sm text-muted-foreground">{role.description}</div>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                
                <div>
                  <Label htmlFor="scope">Scope</Label>
                  <Select value={newRoleKey.scope} onValueChange={(value) => setNewRoleKey(prev => ({ ...prev, scope: value }))}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select scope" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableScopes.map((scope) => (
                        <SelectItem key={scope.value} value={scope.value}>
                          <div className="flex items-center gap-2">
                            {React.createElement(scope.icon, { className: "h-4 w-4" })}
                            <div>
                              <div className="font-medium">{scope.label}</div>
                              <div className="text-sm text-muted-foreground">{scope.description}</div>
                            </div>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="subforumId">Subforum ID (optional)</Label>
                  <Input
                    id="subforumId"
                    type="number"
                    placeholder="Leave empty for global scope"
                    value={newRoleKey.subforumId || ''}
                    onChange={(e) => setNewRoleKey(prev => ({
                      ...prev,
                      subforumId: e.target.value ? parseInt(e.target.value) : undefined
                    }))}
                  />
                </div>
                
                <div>
                  <Label htmlFor="expiresAt">Expires At</Label>
                  <Input
                    id="expiresAt"
                    type="date"
                    value={newRoleKey.expiresAt}
                    onChange={(e) => setNewRoleKey(prev => ({ ...prev, expiresAt: e.target.value }))}
                  />
                </div>
              </div>

              <div>
                <Label>Capabilities</Label>
                <div className="mt-2 grid grid-cols-2 gap-2">
                  {availableCapabilities.map((capability) => (
                    <div key={capability.value} className="flex items-center space-x-2">
                      <input
                        type="checkbox"
                        id={capability.value}
                        checked={newRoleKey.capabilities.includes(capability.value)}
                        onChange={() => toggleCapability(capability.value)}
                        className="rounded"
                      />
                      <Label htmlFor={capability.value} className="text-sm">
                        {capability.label}
                      </Label>
                    </div>
                  ))}
                </div>
              </div>

              <Button onClick={addRoleKey} className="w-full">
                <Plus className="mr-2 h-4 w-4" />
                Add Role Key
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="statistics" className="space-y-6">
            {/* Statistics */}
            {pseudonym && (
              <div className="space-y-4">
                <h3 className="text-lg font-medium">Statistics</h3>
                
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="font-medium">Karma Score:</span> {pseudonym.karmaScore}
                  </div>
                  <div>
                    <span className="font-medium">Posts:</span> {pseudonym.postCount}
                  </div>
                  <div>
                    <span className="font-medium">Comments:</span> {pseudonym.commentCount}
                  </div>
                  <div>
                    <span className="font-medium">Created:</span> {new Date(pseudonym.createdAt).toLocaleDateString()}
                  </div>
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>

        {/* Actions */}
        <div className="flex justify-end gap-3 pt-6">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
