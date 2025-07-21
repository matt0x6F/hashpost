"use client";

import React, { useState } from "react";
import { Button } from "./shadcn/button";
import { useAuth } from "@/lib/auth-context";

interface SimpleModalProps {
  children: React.ReactNode;
  onOpen?: () => void;
}

export function SimpleModal({ children, onOpen }: SimpleModalProps) {
  const { user } = useAuth();
  const [isOpen, setIsOpen] = useState(false);

  if (!user) {
    return null;
  }

  const handleOpen = () => {
    console.log("SimpleModal - opening modal");
    setIsOpen(true);
    if (onOpen) {
      onOpen();
    }
  };

  const handleClose = () => {
    console.log("SimpleModal - closing modal");
    setIsOpen(false);
  };

  return (
    <>
      <div onClick={handleOpen}>
        {children}
      </div>
      
      {isOpen && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 bg-black/50 z-[9999]"
            onClick={handleClose}
          />
          
          {/* Modal */}
          <div className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-white border-2 border-red-500 p-6 rounded-lg z-[10000] min-w-[300px]">
            <h2 className="text-lg font-bold mb-4">Simple Modal Test</h2>
            <p className="mb-4">User: {user.displayName}</p>
            <p className="mb-4">Pseudonyms: {user.pseudonyms.length}</p>
            <Button onClick={handleClose}>Close</Button>
          </div>
        </>
      )}
    </>
  );
} 