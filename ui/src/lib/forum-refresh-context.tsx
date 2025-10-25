'use client';

import React, { createContext, useContext, useState, ReactNode } from 'react';

interface ForumRefreshContextType {
  refreshTrigger: number;
  triggerRefresh: () => void;
}

const ForumRefreshContext = createContext<ForumRefreshContextType | undefined>(undefined);

export function ForumRefreshProvider({ children }: { children: ReactNode }) {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const triggerRefresh = () => {
    setRefreshTrigger(prev => prev + 1);
  };

  return (
    <ForumRefreshContext.Provider value={{ refreshTrigger, triggerRefresh }}>
      {children}
    </ForumRefreshContext.Provider>
  );
}

export function useForumRefresh() {
  const context = useContext(ForumRefreshContext);
  if (context === undefined) {
    throw new Error('useForumRefresh must be used within a ForumRefreshProvider');
  }
  return context;
}
