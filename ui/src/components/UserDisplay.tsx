import React from 'react';

interface Author {
  did: string;
  handle: string;
  display_name?: string | null;
  avatar_url?: string | null;
}

interface UserDisplayProps {
  author: Author;
}

function truncateDID(did: string): string {
  if (did.length <= 20) {
    return did;
  }
  return `${did.substring(0, 8)}...${did.substring(did.length - 8)}`;
}

export function UserDisplay({ author }: UserDisplayProps) {
  const displayText = author.display_name || author.handle || truncateDID(author.did);
  const secondaryText = author.display_name ? `@${author.handle}` : null;
  
  return (
    <div className="inline-flex items-center gap-1">
      <span className="font-medium">{displayText}</span>
      {secondaryText && (
        <span className="text-muted-foreground text-sm">{secondaryText}</span>
      )}
    </div>
  );
}

