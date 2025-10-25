import { useState, useCallback, useEffect } from 'react';
import { getApi } from '@/lib/api-client';
import { VotingApi } from '@/generated/api/src/apis';
import { useAuth } from '@/lib/auth-context';

interface CommentVote {
  vote_type: 'up' | 'down' | null;
  upvotes: number;
  downvotes: number;
  score: number;
}

export function useCommentVote(commentId: string) {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [userVote, setUserVote] = useState<CommentVote>({
    vote_type: null,
    upvotes: 0,
    downvotes: 0,
    score: 0
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchVote = useCallback(async () => {
    if (!commentId || authLoading) return;
    
    // Only fetch votes if user is authenticated
    if (!isAuthenticated) {
      setUserVote({
        vote_type: null,
        upvotes: 0,
        downvotes: 0,
        score: 0
      });
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      const votingApi = getApi(VotingApi);
      const response = await votingApi.getUserVoteOnComment(commentId);
      
      setUserVote({
        vote_type: response.voteType || null,
        upvotes: response.upvotes || 0,
        downvotes: response.downvotes || 0,
        score: response.score || 0
      });
    } catch (err) {
      console.error('Error fetching comment vote:', err);
      // Don't set error for 401 - user might not be logged in
      if (err instanceof Error && !err.message.includes('401')) {
        setError(err.message);
      }
    } finally {
      setIsLoading(false);
    }
  }, [commentId, isAuthenticated, authLoading]);

  const updateVote = useCallback((newVote: CommentVote) => {
    setUserVote(newVote);
  }, []);

  const refetch = useCallback(() => {
    fetchVote();
  }, [fetchVote]);

  useEffect(() => {
    fetchVote();
  }, [fetchVote]);

  return {
    userVote,
    isLoading,
    error,
    refetch,
    updateVote
  };
}
