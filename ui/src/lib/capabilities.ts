import { useCallback } from 'react';
import { getApi } from './api-client';
import { UserManagementApi } from '@/generated/api/src/apis/UserManagementApi';
import { UserRole } from '@/generated/api/src/models/UserRole';

export interface UserCapabilities {
  permissions: string[];
  roles: UserRole[];
}

export class CapabilitiesService {
  private static instance: CapabilitiesService;
  private capabilities: UserCapabilities | null = null;
  private lastFetch: number = 0;
  private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

  static getInstance(): CapabilitiesService {
    if (!CapabilitiesService.instance) {
      CapabilitiesService.instance = new CapabilitiesService();
    }
    return CapabilitiesService.instance;
  }

  async getUserCapabilities(subforumId?: string): Promise<UserCapabilities> {
    const now = Date.now();
    
    // Return cached capabilities if still valid
    if (this.capabilities && (now - this.lastFetch) < this.CACHE_DURATION) {
      return this.capabilities;
    }

    try {
      const userManagementApi = getApi(UserManagementApi);
      
      // Fetch both permissions and roles
      const [permissionsResponse, rolesResponse] = await Promise.all([
        userManagementApi.getMyPermissions(subforumId),
        userManagementApi.getMyRoles(subforumId)
      ]);

      this.capabilities = {
        permissions: permissionsResponse.permissions || [],
        roles: rolesResponse.roles || []
      };

      this.lastFetch = now;
      return this.capabilities;
    } catch (error) {
      console.error('Failed to fetch user capabilities:', error);
      // Return empty capabilities on error
      return {
        permissions: [],
        roles: []
      };
    }
  }

  async hasPermission(permission: string, subforumId?: string): Promise<boolean> {
    const capabilities = await this.getUserCapabilities(subforumId);
    return capabilities.permissions.includes(permission);
  }

  async hasRole(roleName: string, subforumId?: string): Promise<boolean> {
    const capabilities = await this.getUserCapabilities(subforumId);
    return capabilities.roles.some(role => 
      role.roleName === roleName && 
      (subforumId ? role.subforumId === subforumId : role.isPlatformRole)
    );
  }

  async isModerator(subforumId?: string): Promise<boolean> {
    return this.hasRole('moderator', subforumId);
  }

  async isAdmin(subforumId?: string): Promise<boolean> {
    return this.hasRole('admin', subforumId);
  }

  async isOwner(subforumId?: string): Promise<boolean> {
    return this.hasRole('owner', subforumId);
  }

  async canModerateContent(subforumId?: string): Promise<boolean> {
    return this.hasPermission('moderate_content', subforumId);
  }

  async canManageUsers(subforumId?: string): Promise<boolean> {
    return this.hasPermission('manage_users', subforumId);
  }

  async canManageRoles(subforumId?: string): Promise<boolean> {
    return this.hasPermission('manage_roles', subforumId);
  }

  async canCreatePosts(subforumId?: string): Promise<boolean> {
    return this.hasPermission('create_posts', subforumId);
  }

  async canCreateComments(subforumId?: string): Promise<boolean> {
    return this.hasPermission('create_comments', subforumId);
  }

  async canVote(subforumId?: string): Promise<boolean> {
    return this.hasPermission('vote', subforumId);
  }

  // Clear cache (useful for logout or role changes)
  clearCache(): void {
    this.capabilities = null;
    this.lastFetch = 0;
  }

  // Get cached capabilities without fetching
  getCachedCapabilities(): UserCapabilities | null {
    return this.capabilities;
  }
}

// Export singleton instance
export const capabilitiesService = CapabilitiesService.getInstance();

// Helper hooks for React components
export function useCapabilities() {
  const hasPermission = useCallback(capabilitiesService.hasPermission.bind(capabilitiesService), []);
  const hasRole = useCallback(capabilitiesService.hasRole.bind(capabilitiesService), []);
  const isModerator = useCallback(capabilitiesService.isModerator.bind(capabilitiesService), []);
  const isAdmin = useCallback(capabilitiesService.isAdmin.bind(capabilitiesService), []);
  const isOwner = useCallback(capabilitiesService.isOwner.bind(capabilitiesService), []);
  const canModerateContent = useCallback(capabilitiesService.canModerateContent.bind(capabilitiesService), []);
  const canManageUsers = useCallback(capabilitiesService.canManageUsers.bind(capabilitiesService), []);
  const canManageRoles = useCallback(capabilitiesService.canManageRoles.bind(capabilitiesService), []);
  const canCreatePosts = useCallback(capabilitiesService.canCreatePosts.bind(capabilitiesService), []);
  const canCreateComments = useCallback(capabilitiesService.canCreateComments.bind(capabilitiesService), []);
  const canVote = useCallback(capabilitiesService.canVote.bind(capabilitiesService), []);
  const getUserCapabilities = useCallback(capabilitiesService.getUserCapabilities.bind(capabilitiesService), []);
  const clearCache = useCallback(capabilitiesService.clearCache.bind(capabilitiesService), []);

  return {
    hasPermission,
    hasRole,
    isModerator,
    isAdmin,
    isOwner,
    canModerateContent,
    canManageUsers,
    canManageRoles,
    canCreatePosts,
    canCreateComments,
    canVote,
    getUserCapabilities,
    clearCache
  };
}
