'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/shadcn/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card'
import { Badge } from '@/components/shadcn/badge'
import { ArrowLeft, User, Shield, Calendar, Mail, MessageSquare, FileText, Activity } from 'lucide-react'
import { useEffect, useState, use } from 'react'
import { SearchUser } from '@/generated/api/src/models/SearchUser'
import { PseudonymProfileResponseBody } from '@/generated/api/src/models/PseudonymProfileResponseBody'
import { Post } from '@/generated/api/src/models/Post'
import { Comment } from '@/generated/api/src/models/Comment'
import { getApi } from '@/lib/api-client'
import { PseudonymsApi } from '@/generated/api/src/apis/PseudonymsApi'

interface UserDetailPageProps {
  params: Promise<{
    id: string
  }>
}

export default function UserDetailPage({ params }: UserDetailPageProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [user, setUser] = useState<SearchUser | null>(null)
  const [pseudonymProfile, setPseudonymProfile] = useState<PseudonymProfileResponseBody | null>(null)
  const [recentPosts, setRecentPosts] = useState<Post[]>([])
  const [recentComments, setRecentComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Unwrap the params Promise using React.use()
  const { id } = use(params)

  // Get search context from URL params
  const query = searchParams.get('query') || ''
  const page = searchParams.get('page') || '1'

  useEffect(() => {
    // Debug: Log what we're looking for
    console.log('=== UserDetailPage useEffect ===');
    console.log('Looking for user ID:', id);
    
    // Try to get user data from sessionStorage first (stored by search results)
    const userDataKey = `user_detail_${id}`;
    console.log('SessionStorage key:', userDataKey);
    
    const storedUserData = sessionStorage.getItem(userDataKey);
    console.log('Retrieved from sessionStorage:', storedUserData);
    
    if (storedUserData) {
      try {
        const decodedUserData = JSON.parse(storedUserData);
        console.log('Parsed user data:', decodedUserData);
        console.log('User email:', decodedUserData.email);
        console.log('User pseudonyms:', decodedUserData.pseudonyms);
        
        setUser(decodedUserData);
        // Fetch additional pseudonym details
        fetchPseudonymDetails(decodedUserData.pseudonymId);
        // Clean up the stored data after using it
        sessionStorage.removeItem(userDataKey);
        return;
      } catch (err) {
        console.warn('Failed to parse user data from sessionStorage:', err);
        sessionStorage.removeItem(userDataKey);
        setError('Failed to parse user data');
        setLoading(false);
      }
    } else {
      // No user data found in sessionStorage - try to fetch from API
      console.log('No user data found in sessionStorage, attempting API fetch...');
      fetchUserFromAPI();
    }
  }, [id])

  const fetchUserFromAPI = async () => {
    try {
      setLoading(true);
      
      // Get real user data from the pseudonym profile API
      await fetchPseudonymDetails(id);
      
    } catch (err) {
      console.error('Failed to fetch user from API:', err);
      setError('Failed to load user details. Please return to search results and try again.');
      setLoading(false);
    }
  };

  const fetchPseudonymDetails = async (pseudonymId: string) => {
    try {
      const api = getApi(PseudonymsApi);
      
      // Get pseudonym profile - this gives us the real pseudonym data
      const profileResponse = await api.getPseudonymProfileBySlug(pseudonymId);
      setPseudonymProfile(profileResponse);
      
      // Create user object from real API data
      const realUser: SearchUser = {
        pseudonymId: profileResponse.pseudonymId,
        displayName: profileResponse.displayName,
        karmaScore: profileResponse.karmaScore,
        createdAt: profileResponse.createdAt,
        email: 'Email not available', // We can't get this from pseudonym profile API
        userId: 0, // We can't get this from pseudonym profile API
        pseudonyms: [{
          id: profileResponse.pseudonymId,
          displayName: profileResponse.displayName,
          isDefault: true,
          createdAt: profileResponse.createdAt
        }]
      };
      
      setUser(realUser);
      
      // Get recent posts
      try {
        const postsResponse = await api.getPostsByPseudonym(pseudonymId, 1, 5);
        setRecentPosts(postsResponse.posts || []);
      } catch (postsErr) {
        console.warn('Failed to fetch posts:', postsErr);
        setRecentPosts([]);
      }
      
      // Get recent comments
      try {
        const commentsResponse = await api.getCommentsByPseudonym(pseudonymId, 1, 5);
        setRecentComments(commentsResponse.comments || []);
      } catch (commentsErr) {
        console.warn('Failed to fetch comments:', commentsErr);
        setRecentComments([]);
      }
      
      setLoading(false);
      
    } catch (err) {
      console.error('Failed to fetch pseudonym details:', err);
      setError('Failed to load pseudonym details');
      setLoading(false);
    }
  };

  const handleBack = () => {
    // Navigate back to search results with preserved query parameters
    const backUrl = `/admin?tab=users&query=${encodeURIComponent(query)}&page=${page}`
    router.push(backUrl)
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Search
          </Button>
        </div>
        <div className="mt-8 text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p className="mt-2 text-muted-foreground">Loading user details...</p>
        </div>
      </div>
    )
  }

  if (error || !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Search
          </Button>
        </div>
        <div className="mt-8 text-center">
          <p className="text-destructive">{error || 'User not found'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      {/* Header with back button */}
      <div className="flex items-center space-x-4 mb-8">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Search
        </Button>
        <h1 className="text-2xl font-bold tracking-tight">Pseudonym Details</h1>
      </div>

      {/* User Information */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <User className="h-5 w-5" />
            <span>Pseudonym Information</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-foreground">Display Name</label>
              <p className="text-sm text-foreground">{user.displayName}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Pseudonym ID</label>
              <p className="text-sm text-foreground font-mono">{user.pseudonymId}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Slug</label>
              <p className="text-sm text-foreground font-mono">{pseudonymProfile?.slug || 'N/A'}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Account Created</label>
              <p className="text-sm text-foreground">
                {new Date(user.createdAt).toLocaleDateString()}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Last Updated</label>
              <p className="text-sm text-foreground">
                {pseudonymProfile?.updatedAt ? new Date(pseudonymProfile.updatedAt).toLocaleDateString() : 'N/A'}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Last Active</label>
              <p className="text-sm text-foreground">
                {pseudonymProfile?.lastActiveAt ? new Date(pseudonymProfile.lastActiveAt).toLocaleDateString() : 'N/A'}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Karma Score</label>
              <p className="text-sm text-foreground">{user.karmaScore}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Show Karma</label>
              <p className="text-sm text-foreground">
                {pseudonymProfile?.showKarma ? 'Yes' : 'No'}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Status</label>
              <p className="text-sm text-foreground">
                <Badge variant={pseudonymProfile?.isActive ? "default" : "secondary"}>
                  {pseudonymProfile?.isActive ? 'Active' : 'Inactive'}
                </Badge>
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Allow Direct Messages</label>
              <p className="text-sm text-foreground">
                {pseudonymProfile?.allowDirectMessages ? 'Yes' : 'No'}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Total Posts</label>
              <p className="text-sm text-foreground">{pseudonymProfile?.postCount || 0}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-foreground">Total Comments</label>
              <p className="text-sm text-foreground">{pseudonymProfile?.commentCount || 0}</p>
            </div>
          </div>
          
          {pseudonymProfile?.bio && (
            <div className="mt-4">
              <label className="text-sm font-medium text-foreground">Bio</label>
              <p className="text-sm text-foreground mt-1">{pseudonymProfile.bio}</p>
            </div>
          )}
          
          {pseudonymProfile?.websiteUrl && (
            <div className="mt-4">
              <label className="text-sm font-medium text-foreground">Website</label>
              <p className="text-sm text-foreground mt-1">
                <a 
                  href={pseudonymProfile.websiteUrl} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {pseudonymProfile.websiteUrl}
                </a>
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Recent Activity */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5" />
            <span>Recent Activity</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentPosts.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold mb-2">Posts</h3>
                <div className="space-y-2">
                  {recentPosts.map((post) => (
                    <div key={post.postId} className="bg-muted/50 p-3 rounded-lg">
                      <p className="text-sm text-foreground">{post.content}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Posted on {new Date(post.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}
            {recentComments.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold mb-2">Comments</h3>
                <div className="space-y-2">
                  {recentComments.map((comment) => (
                    <div key={comment.commentId} className="bg-muted/50 p-3 rounded-lg">
                      <p className="text-sm text-foreground">{comment.content}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Commented on {new Date(comment.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )}
            {recentPosts.length === 0 && recentComments.length === 0 && (
              <p className="text-sm text-muted-foreground">No recent activity found.</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
