"use client";

import { useState, useEffect, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/shadcn/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Input } from "@/components/shadcn/input";
import { Label } from "@/components/shadcn/label";
import { getApi } from "@/lib/api-client";
import { AuthenticationApi } from "@/generated/api/src/apis/AuthenticationApi";
import { toast } from "sonner";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  const { login } = useAuth();
  
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState("");
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  useEffect(() => {
    if (!token) {
      setError("Invalid verification link. Please check your email for the correct link.");
      return;
    }

    // Auto-verify the email when the page loads
    verifyEmail();
  }, [token]);

  const verifyEmail = async () => {
    if (!token) return;

    setIsLoading(true);
    setError("");

    try {
      // Email verification not available in atproto system
      toast.error("Email verification is not available in the atproto system");
    } catch (error) {
      console.error("Email verification failed:", error);
      setError("Failed to verify email. The link may have expired or be invalid.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoggingIn(true);
    setError("");

    try {
      const authApi = getApi(AuthenticationApi);
      const loginResponse = await authApi.loginUser({
        email: loginEmail,
        password: loginPassword
      });
      
      login(loginEmail, loginPassword);
      toast.success("Logged in successfully!");
      // Redirect to home or dashboard
      window.location.href = "/";
    } catch (error) {
      console.error("Login failed:", error);
      setError("Login failed. Please check your email and password.");
    } finally {
      setIsLoggingIn(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-lg">
          <CardHeader className="text-center">
            <CardTitle>Invalid Verification Link</CardTitle>
            <CardDescription>
              The email verification link is invalid or has expired.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground text-center">
              Please check your email for the correct verification link, or contact support if you need assistance.
            </p>
            <div className="text-center">
              <Link href="/" className="text-sm text-primary hover:underline">
                Return to Home
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isSuccess) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-lg">
          <CardHeader className="text-center">
            <CardTitle>Email Verified Successfully!</CardTitle>
            <CardDescription>
              Your email address has been verified. You can now log in to your account.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <form onSubmit={handleLogin} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={loginEmail}
                  onChange={(e) => setLoginEmail(e.target.value)}
                  placeholder="Enter your email"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={loginPassword}
                  onChange={(e) => setLoginPassword(e.target.value)}
                  placeholder="Enter your password"
                  required
                />
              </div>
              {error && (
                <div className="text-sm text-red-600 dark:text-red-400 text-center">
                  {error}
                </div>
              )}
              <Button type="submit" className="w-full" disabled={isLoggingIn}>
                {isLoggingIn ? "Logging in..." : "Log In"}
              </Button>
            </form>
            <div className="text-center">
              <Link href="/" className="text-sm text-primary hover:underline">
                Return to Home
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <CardTitle>Verifying Your Email</CardTitle>
          <CardDescription>
            Please wait while we verify your email address...
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading && (
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-sm text-muted-foreground">Verifying email...</p>
            </div>
          )}
          
          {error && (
            <div className="text-sm text-red-600 dark:text-red-400 text-center">
              {error}
            </div>
          )}
          
          {!isLoading && !error && (
            <div className="text-center">
              <Button onClick={verifyEmail} className="w-full">
                Try Again
              </Button>
            </div>
          )}
          
          <div className="text-center">
            <Link href="/" className="text-sm text-primary hover:underline">
              Return to Home
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-lg">
          <CardHeader className="text-center">
            <CardTitle>Loading...</CardTitle>
            <CardDescription>
              Please wait while we load the verification page...
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    }>
      <VerifyEmailContent />
    </Suspense>
  );
} 