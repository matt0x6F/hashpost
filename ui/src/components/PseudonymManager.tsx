"use client";

import React, { useState } from "react";
import { Button } from "./shadcn/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "./shadcn/dialog";
import { useAuth } from "@/lib/auth-context";

interface PseudonymManagerProps {
  children: React.ReactNode;
  onOpen?: () => void;
}

export function PseudonymManager({ children, onOpen }: PseudonymManagerProps) {
  const { user } = useAuth();
  const [isOpen, setIsOpen] = useState(false);

  console.log("PseudonymManager - user:", user);
  
  if (!user) {
    console.log("PseudonymManager - no user, returning null");
    return null;
  }

  console.log("PseudonymManager - rendering dialog, isOpen:", isOpen);
  
  console.log("PseudonymManager - about to render dialog, isOpen:", isOpen);
  
  return (
    <Dialog open={isOpen} onOpenChange={(open) => {
      console.log("PseudonymManager - dialog onOpenChange:", open);
      setIsOpen(open);
      if (open && onOpen) {
        onOpen();
      }
    }}>
      <DialogTrigger asChild>
        <div onClick={() => console.log("PseudonymManager - trigger clicked")}>
          {children}
        </div>
      </DialogTrigger>
      <DialogContent className="z-[9999]">
        <DialogHeader>
          <DialogTitle>Test Dialog</DialogTitle>
        </DialogHeader>
        <div className="p-4 bg-red-100 border-2 border-red-500">
          <h3 className="text-lg font-medium text-red-800">Test Dialog - Should be visible!</h3>
          <p className="text-red-800">User: {user.displayName}</p>
          <p className="text-red-800">Pseudonyms: {user.pseudonyms.length}</p>
          <p className="text-red-800">Dialog is open: {isOpen ? 'YES' : 'NO'}</p>
          <Button onClick={() => {
            console.log("Close button clicked");
            setIsOpen(false);
          }}>Close</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
} 