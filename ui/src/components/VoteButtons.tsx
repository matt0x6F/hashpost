import { Button } from '@/components/shadcn/button';
import { ArrowUp, ArrowDown } from 'lucide-react';
import React from 'react';

interface VoteButtonsProps {
  score: number;
  userVote: number; // 1 = upvoted, -1 = downvoted, 0 = no vote
  onVote: (vote: number) => void;
  disabled?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export const VoteButtons: React.FC<VoteButtonsProps> = ({
  score,
  userVote,
  onVote,
  disabled = false,
  size = 'md',
}) => {
  // Tailwind size classes
  const arrowSize = size === 'sm' ? 'w-3 h-3' : size === 'lg' ? 'w-6 h-6' : 'w-4 h-4';
  const buttonSize = size === 'sm' ? 'h-6 px-2' : size === 'lg' ? 'h-10 px-4' : 'h-8 px-2';
  const scoreText = size === 'sm' ? 'text-xs' : size === 'lg' ? 'text-lg' : 'text-sm';

  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size={size === 'sm' ? 'sm' : size === 'lg' ? 'lg' : 'sm'}
        className={`${buttonSize}`}
        disabled={disabled}
        onClick={() => onVote(userVote === 1 ? 0 : 1)}
        aria-label={userVote === 1 ? 'Remove upvote' : 'Upvote'}
      >
        <ArrowUp className={`${arrowSize} ${userVote === 1 ? 'text-emerald-500' : 'text-muted-foreground'}`} />
      </Button>
      <span className={`${scoreText} font-medium min-w-[2rem] text-center`}>
        {score >= 1000 ? `${(score / 1000).toFixed(1)}k` : score}
      </span>
      <Button
        variant="ghost"
        size={size === 'sm' ? 'sm' : size === 'lg' ? 'lg' : 'sm'}
        className={`${buttonSize}`}
        disabled={disabled}
        onClick={() => onVote(userVote === -1 ? 0 : -1)}
        aria-label={userVote === -1 ? 'Remove downvote' : 'Downvote'}
      >
        <ArrowDown className={`${arrowSize} ${userVote === -1 ? 'text-emerald-500' : 'text-muted-foreground'}`} />
      </Button>
    </div>
  );
};

export default VoteButtons; 