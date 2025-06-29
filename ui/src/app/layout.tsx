"use client";
import { SidebarProvider, SidebarTrigger } from "../components/shadcn/sidebar";
import { AppSidebar } from "../components/Sidebar";
import TopBar from "../components/TopBar";
import { Geist, Geist_Mono } from "next/font/google";
import { AuthProvider } from "../lib/auth-context";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}>
        <AuthProvider>
          <SidebarProvider>
            <div className="flex flex-col h-screen">
              <TopBar />
              <div className="flex flex-1 overflow-hidden">
                <AppSidebar />
                <main className="flex-1 p-6 md:p-10 bg-background overflow-y-auto">
                  <SidebarTrigger />
                  {children}
                </main>
              </div>
            </div>
          </SidebarProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
