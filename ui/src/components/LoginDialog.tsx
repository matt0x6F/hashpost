"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./shadcn/dialog";
import { Button } from "./shadcn/button";
import { Input } from "./shadcn/input";
import { Label } from "./shadcn/label";
import { getApi } from "@/lib/api-client";
import { AuthenticationApi } from "@/generated/api/src/apis/AuthenticationApi";
import { useAuth } from "@/lib/auth-context";

interface LoginDialogProps {
  children: React.ReactNode;
  onLoginSuccess?: () => void;
  onSignupSuccess?: () => void;
}

type DialogMode = "login" | "signup";

export function LoginDialog({ children, onLoginSuccess, onSignupSuccess }: LoginDialogProps) {
  const { login } = useAuth();
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState<DialogMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");

    try {
      if (mode === "login") {
        const authApi = getApi(AuthenticationApi);
        const response = await authApi.loginUser({ email, password });
        console.log("Login successful:", response);
        
        // Store user data in global state
        login(response);
        
        setOpen(false);
        onLoginSuccess?.();
      } else {
        const authApi = getApi(AuthenticationApi);
        const signupResponse = await authApi.registerUser({
          email,
          password,
          displayName,
        });
        console.log("Signup successful:", signupResponse);
        
        // Store user data in global state
        login(signupResponse);
        
        setOpen(false);
        onSignupSuccess?.();
      }
    } catch (err) {
      setError(mode === "login" ? "Invalid email or password" : "Failed to create account");
      console.error(`${mode} failed:`, err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      // Reset form when dialog closes
      setEmail("");
      setPassword("");
      setConfirmPassword("");
      setDisplayName("");
      setError("");
      setMode("login");
    }
  };

  const switchMode = (newMode: DialogMode) => {
    setMode(newMode);
    setError("");
    setPassword("");
    setConfirmPassword("");
    setDisplayName("");
  };

  const isSignupValid = mode === "signup" && password !== confirmPassword;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>
            {mode === "login" ? "Login to HashPost" : "Create Account"}
          </DialogTitle>
          <DialogDescription>
            {mode === "login" 
              ? "Enter your credentials to access your account."
              : "Create a new account to get started with HashPost."
            }
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isLoading}
            />
          </div>
          
          {mode === "signup" && (
            <div className="space-y-2">
              <Label htmlFor="displayName">Display Name</Label>
              <Input
                id="displayName"
                type="text"
                placeholder="Enter your display name"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                required
                disabled={isLoading}
              />
            </div>
          )}
          
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={isLoading}
            />
          </div>
          
          {mode === "signup" && (
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                type="password"
                placeholder="Confirm your password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                disabled={isLoading}
                className={isSignupValid ? "border-red-500" : ""}
              />
              {isSignupValid && (
                <div className="text-sm text-red-600 dark:text-red-400">
                  Passwords do not match
                </div>
              )}
            </div>
          )}
          
          {error && (
            <div className="text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}
          
          <DialogFooter>
            <Button
              type="submit"
              disabled={isLoading || (mode === "signup" && isSignupValid)}
              className="w-full"
            >
              {isLoading 
                ? (mode === "login" ? "Signing in..." : "Creating account...")
                : (mode === "login" ? "Sign In" : "Create Account")
              }
            </Button>
          </DialogFooter>
          <div className="text-center text-sm mt-2">
            {mode === "login" ? (
              <span>
                Don&apos;t have an account?{" "}
                <button
                  type="button"
                  onClick={() => switchMode("signup")}
                  className="text-primary hover:underline focus:outline-none"
                >
                  Sign up
                </button>
              </span>
            ) : (
              <span>
                Already have an account?{" "}
                <button
                  type="button"
                  onClick={() => switchMode("login")}
                  className="text-primary hover:underline focus:outline-none"
                >
                  Sign in
                </button>
              </span>
            )}
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
} 