// ShadCN Sidebar primitives (generated from https://ui.shadcn.com/docs/components/sidebar)
import * as React from "react";

// Context for sidebar open/close state
const SidebarContext = React.createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
} | null>(null);

export function SidebarProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(true);
  return (
    <SidebarContext.Provider value={{ open, setOpen }}>
      {children}
    </SidebarContext.Provider>
  );
}

export function useSidebar() {
  const context = React.useContext(SidebarContext);
  if (!context) throw new Error("useSidebar must be used within a SidebarProvider.");
  return context;
}

export function Sidebar({ children }: { children: React.ReactNode }) {
  const { open } = useSidebar();
  return (
    <aside
      data-open={open}
      className={`sidebar fixed md:static z-40 md:z-auto h-full w-64 bg-zinc-900 border-r border-zinc-800 transition-transform duration-200 ${open ? "translate-x-0" : "-translate-x-full md:translate-x-0"}`}
    >
      {children}
    </aside>
  );
}

export function SidebarContent({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col h-full overflow-y-auto">{children}</div>;
}

export function SidebarGroup({ children }: { children: React.ReactNode }) {
  return <div className="mb-6 last:mb-0">{children}</div>;
}

export function SidebarGroupLabel({ children }: { children: React.ReactNode }) {
  return <div className="text-xs font-semibold uppercase text-zinc-400 mb-2 px-4">{children}</div>;
}

export function SidebarGroupContent({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col gap-1">{children}</div>;
}

export function SidebarMenu({ children }: { children: React.ReactNode }) {
  return <ul className="flex flex-col gap-1">{children}</ul>;
}

export function SidebarMenuItem({ children }: { children: React.ReactNode }) {
  return <li>{children}</li>;
}

export function SidebarMenuButton({ asChild = false, children }: { asChild?: boolean; children: React.ReactNode }) {
  if (asChild) return <>{children}</>;
  return (
    <button className="w-full flex items-center gap-2 px-4 py-2 rounded hover:bg-zinc-800 text-zinc-100 transition-colors">
      {children}
    </button>
  );
}

export function SidebarTrigger() {
  const { open, setOpen } = useSidebar();
  return (
    <button
      className="md:hidden fixed top-4 left-4 z-50 w-10 h-10 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center shadow-lg"
      aria-label="Toggle sidebar"
      onClick={() => setOpen(!open)}
    >
      <span className="sr-only">Toggle sidebar</span>
      <svg width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-menu"><line x1="4" y1="12" x2="20" y2="12"/><line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="18" x2="20" y2="18"/></svg>
    </button>
  );
} 