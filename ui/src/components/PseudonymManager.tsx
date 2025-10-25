"use client";

import React, { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./shadcn/dialog";
import { Button } from "./shadcn/button";
import { useAuth } from "@/lib/auth-context";

interface PseudonymManagerProps {
  children: React.ReactNode;
  onOpen?: () => void;
}

export function PseudonymManager({ children, onOpen }: PseudonymManagerProps) {
  const { user } = useAuth();
  const [isOpen, setIsOpen] = useState(false);

  if (!user) {
    return null;
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => {
      setIsOpen(open);
      if (open && onOpen) {
        onOpen();
      }
    }}>
      <DialogTrigger asChild>
        <div>
          {children}
        </div>
      </DialogTrigger>
      <DialogContent className="z-[9999]">
        <DialogHeader>
          <DialogTitle>Pseudonym Manager</DialogTitle>
        </DialogHeader>
        <div className="p-4">
          <h3 className="text-lg font-medium mb-4">Manage your pseudonyms</h3>
          <p className="text-muted-foreground mb-2">User: {user.handle}</p>
          <p className="text-muted-foreground mb-4">Pseudonyms: N/A (not available in atproto system)</p>
          <Button onClick={() => {
            setIsOpen(false);
          }}>Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
} 