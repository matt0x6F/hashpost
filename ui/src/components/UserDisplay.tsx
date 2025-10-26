import React from 'react';
import Link from 'next/link';

interface Author {
  did: string;
  handle: string;
  displayName?: string | null;
  avatarUrl?: string | null;
}

interface UserDisplayProps {
  author: Author;
  linkToProfile?: boolean;
}

function truncateDID(did: string): string {
  if (did.length <= 20) {
    return did;
  }
  return `${did.substring(0, 8)}...${did.substring(did.length - 8)}`;
}

export function UserDisplay({ author, linkToProfile = true }: UserDisplayProps) {
  const displayText = author.displayName || author.handle || truncateDID(author.did);
  const secondaryText = author.displayName ? `@${author.handle}` : null;
  
  const content = (
    <div className="inline-flex items-center gap-1">
      <span className="font-medium">{displayText}</span>
      {secondaryText && (
        <span className="text-muted-foreground text-sm">{secondaryText}</span>
      )}
    </div>
  );

  if (linkToProfile && author.handle) {
    return (
      <Link 
        href={`/u/${author.handle}`}
        className="hover:underline transition-colors"
      >
        {content}
      </Link>
    );
  }

  return content;
}

