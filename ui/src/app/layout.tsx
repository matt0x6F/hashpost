"use client";
import { SidebarProvider, useSidebar } from "../components/shadcn/sidebar";
import { AppSidebar } from "../components/Sidebar";
import TopBar from "../components/TopBar";
import { Geist, Geist_Mono } from "next/font/google";
import { AuthProvider } from "../lib/auth-context";
import { Toaster } from "../components/shadcn/sonner";
import { ThemeProvider } from "next-themes";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

function LayoutContent({ children }: { children: React.ReactNode }) {
  const { open, setOpen } = useSidebar();
  
  return (
    <div className="flex flex-col h-screen">
      <TopBar onMenuClick={() => setOpen(!open)} />
      <div className="flex flex-1 overflow-hidden">
        <AppSidebar />
        <main className="flex-1 p-2 sm:p-4 md:p-6 bg-background overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem
          disableTransitionOnChange
        >
          <AuthProvider>
            <SidebarProvider>
              <LayoutContent>{children}</LayoutContent>
            </SidebarProvider>
          </AuthProvider>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
