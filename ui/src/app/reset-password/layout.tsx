"use client";

import { Toaster } from "../../components/shadcn/sonner";
import { ThemeProvider } from "next-themes";
import "../globals.css";

export default function ResetPasswordLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <title>Reset Password - HashPost</title>
      </head>
      <body className="antialiased bg-background text-foreground">
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem
          disableTransitionOnChange
        >
          {/* No AuthProvider - completely isolated */}
          <div className="min-h-screen flex items-center justify-center bg-background">
            {children}
          </div>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
} 