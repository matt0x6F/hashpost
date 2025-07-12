import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "./shadcn/sidebar";
import { Home, Radar, ChartNoAxesCombined, Users, Plus } from "lucide-react";
import { CreateForumDialog } from "./CreateForumDialog";

const items = [
  { title: "Home", url: "/", icon: Home },
  { title: "Forums", url: "/forums", icon: Users },
  { title: "Popular", url: "/popular", icon: ChartNoAxesCombined },
  { title: "All", url: "/all", icon: Radar },
];

export function AppSidebar() {
  return (
    <aside className="hidden md:block w-64 bg-[var(--slate-night)] border-r border-[var(--graphite-gray)] flex flex-col">
      <div className="flex-1 overflow-y-auto p-4">
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <a
                      href={item.url}
                      className="flex items-center gap-3 px-3 py-2 rounded transition-colors text-[var(--chalk-white)] font-medium font-sans hover:bg-[var(--mint-hover)] hover:text-[var(--slate-night)]"
                    >
                      <item.icon className="w-5 h-5" />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        
        <div className="mt-6 pt-6 border-t border-[var(--graphite-gray)]">
          <CreateForumDialog>
            <button className="w-full flex items-center gap-3 px-3 py-2 rounded transition-colors text-[var(--chalk-white)] font-medium font-sans hover:bg-[var(--mint-hover)] hover:text-[var(--slate-night)]">
              <Plus className="w-5 h-5" />
              <span>Create Forum</span>
            </button>
          </CreateForumDialog>
        </div>
      </div>
    </aside>
  );
} 