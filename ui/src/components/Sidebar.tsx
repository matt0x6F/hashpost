import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "./shadcn/sidebar";
import { Home, Radar, ChartNoAxesCombined, Users, Plus, X, Shield } from "lucide-react";
import { CreateForumDialog } from "./CreateForumDialog";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { getApi } from "@/lib/api-client";
import { SubscriptionsApi } from "@/generated/api/src/apis/SubscriptionsApi";
import { toast } from "sonner";
import type { Subforum } from "@/generated/api/src/models";
import Link from "next/link";
import { useCapabilities } from "@/lib/capabilities";

const items = [
  { title: "Home", url: "/", icon: Home },
  { title: "Forums", url: "/forums", icon: Users },
  { title: "Popular", url: "/popular", icon: ChartNoAxesCombined },
  { title: "All", url: "/all", icon: Radar },
];

export function AppSidebar() {
  const [subscriptions, setSubscriptions] = useState<Subforum[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasPlatformAdmin, setHasPlatformAdmin] = useState(false);
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const { hasRole } = useCapabilities();

  const checkPlatformAdminRole = async () => {
    if (!isAuthenticated) {
      setHasPlatformAdmin(false);
      return;
    }

    try {
      // Check for platform admin role (no subforumId means platform-level role)
      const isPlatformAdmin = await hasRole('platform_admin');
      setHasPlatformAdmin(isPlatformAdmin);
    } catch (error) {
      console.error('Error checking platform admin role:', error);
      setHasPlatformAdmin(false);
    }
  };

  const loadSubscriptions = async () => {
    if (!isAuthenticated) {
      return;
    }

    setIsLoading(true);
    try {
      const subscriptionsApi = getApi(SubscriptionsApi);
      const response = await subscriptionsApi.listMySubscriptions(20, 0);
      setSubscriptions(response.subscriptions?.map(sub => ({
        id: sub.subforumSlug,
        name: sub.subforumName || sub.subforumSlug,
        slug: sub.subforumSlug,
        description: undefined,
        createdBy: '',
        createdAt: new Date(sub.subscribedAt),
        updatedAt: undefined,
        subscriberCount: undefined,
        postCount: undefined
      })) || []);
    } catch (err: unknown) {
      console.error('Error loading subscriptions:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to load subscriptions';
      toast.error('Failed to load subscriptions', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleUnsubscribe = async (subforum: Subforum) => {
    if (!isAuthenticated) {
      toast.error('Please log in to manage subscriptions');
      return;
    }

    try {
      const subscriptionsApi = getApi(SubscriptionsApi);
      await subscriptionsApi.unsubscribeFromSubforum(subforum.slug);
      
      // Remove from local state
      setSubscriptions(prev => prev.filter(sub => sub.slug !== subforum.slug));
      toast.success('Unsubscribed from subforum');
    } catch (error) {
      console.error('Unsubscribe error:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to unsubscribe';
      toast.error(errorMessage);
    }
  };

  useEffect(() => {
    loadSubscriptions();
  }, [isAuthenticated]);

  useEffect(() => {
    checkPlatformAdminRole();
  }, [isAuthenticated]);

  return (
    <Sidebar>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <Link
                      href={item.url}
                      className="flex items-center gap-3 px-3 py-2 rounded transition-colors text-sidebar-foreground font-medium font-sans hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                    >
                      <item.icon className="w-5 h-5" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        
        <div className="mt-6 pt-6 border-t border-sidebar-border">
          <CreateForumDialog>
            <button className="w-full flex items-center gap-3 px-3 py-2 rounded transition-colors text-sidebar-foreground font-medium font-sans hover:bg-sidebar-accent hover:text-sidebar-accent-foreground">
              <Plus className="w-5 h-5" />
              <span>Create Forum</span>
            </button>
          </CreateForumDialog>
        </div>

        {/* Admin Section - Only show for users with platform admin role */}
        {!authLoading && isAuthenticated && hasPlatformAdmin && (
          <div className="mt-6 pt-6 border-t border-sidebar-border">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-medium text-sidebar-foreground">Administration</h3>
            </div>
            
            <Link
              href="/admin"
              className="flex items-center gap-3 px-3 py-2 rounded transition-colors text-sidebar-foreground font-medium font-sans hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
            >
              <Shield className="w-5 h-5" />
              <span>Platform Admin</span>
            </Link>
          </div>
        )}

        {/* Subscriptions Section */}
        {!authLoading && isAuthenticated && (
          <div className="mt-6 pt-6 border-t border-sidebar-border">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-medium text-sidebar-foreground">Subscriptions</h3>
              {isLoading && (
                <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
              )}
            </div>
            
            {subscriptions.length === 0 ? (
              <p className="text-xs text-sidebar-muted-foreground">
                No subscriptions yet
              </p>
            ) : (
              <div className="space-y-1">
                {subscriptions.map((subforum) => (
                  <div
                    key={subforum.slug}
                    className="group relative flex items-center gap-2 px-3 py-2 rounded transition-colors text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  >
                    <Link
                      href={`/h/${subforum.slug}`}
                      className="flex-1 flex items-center gap-2 min-w-0"
                    >
                      <span className="text-xs font-medium truncate">
                        h/{subforum.slug}
                      </span>
                    </Link>
                    
                    {/* Unsubscribe button - only visible on hover */}
                    <button
                      onClick={() => handleUnsubscribe(subforum)}
                      className="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded hover:bg-sidebar-accent-foreground/10"
                      title="Unsubscribe"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </SidebarContent>
    </Sidebar>
  );
} 