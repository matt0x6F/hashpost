import React from 'react';
import { Badge } from './shadcn/badge';
import { Lock, Pin, EyeOff } from 'lucide-react';

interface PostBadgesProps {
  isSticky?: boolean;
  isLocked?: boolean;
  isRemoved?: boolean;
  isSpoiler?: boolean;
  isNsfw?: boolean;
  className?: string;
}

export function PostBadges({ 
  isSticky, 
  isLocked, 
  isRemoved, 
  isSpoiler, 
  isNsfw, 
  className = "" 
}: PostBadgesProps) {
  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {isSpoiler && (
        <Badge variant="secondary" className="text-xs">
          Spoiler
        </Badge>
      )}
      {isNsfw && (
        <Badge variant="destructive" className="text-xs">
          NSFW
        </Badge>
      )}
      {isSticky && (
        <Badge variant="secondary" className="text-xs">
          <Pin className="w-3 h-3 mr-1" />
          Sticky
        </Badge>
      )}
      {isLocked && (
        <Badge variant="destructive" className="text-xs">
          <Lock className="w-3 h-3 mr-1" />
          Locked
        </Badge>
      )}
      {isRemoved && (
        <Badge variant="destructive" className="text-xs">
          <EyeOff className="w-3 h-3 mr-1" />
          Removed
        </Badge>
      )}
    </div>
  );
} 