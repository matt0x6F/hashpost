'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/shadcn/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card'
import { Badge } from '@/components/shadcn/badge'
import { ArrowLeft, User, Activity, Calendar } from 'lucide-react'
import { useEffect, useState, use } from 'react'

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

  const formatDate = (dateString: string) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  useEffect(() => {
    // Debug: Log what we're looking for
    		// Debug: UserDetailPage useEffect with user ID
    
    // Try to get user data from sessionStorage first (stored by search results)
    const userDataKey = `user_detail_${id}`;
    		// Debug: sessionStorage key for user data
    
    const storedUserData = sessionStorage.getItem(userDataKey);
    		// Debug: retrieved user data from sessionStorage
    
    if (storedUserData) {
      try {
        const decodedUserData = JSON.parse(storedUserData);
        		// Debug: parsed user data with email and pseudonyms
        
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
      		// Debug: no user data in sessionStorage, attempting API fetch
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
    // Get return parameters from URL
    const returnTab = searchParams.get('returnTab') || 'pseudonyms'
    const returnQuery = searchParams.get('returnQuery') || ''
    const returnPage = searchParams.get('returnPage') || '1'
    
    // Navigate back to the correct tab with preserved state
    const backUrl = `/admin?tab=${returnTab}&query=${encodeURIComponent(returnQuery)}&page=${returnPage}`
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

  if (error || !pseudonymProfile) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" size="sm" onClick={handleBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Search
          </Button>
        </div>
        <div className="mt-8 text-center">
          <p className="text-destructive">{error || 'Pseudonym not found'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6 max-w-7xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Button onClick={handleBack} variant="outline" size="sm">
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Pseudonym List
        </Button>
        <div>
          <h1 className="text-3xl font-bold">Pseudonym Details</h1>
          <p className="text-muted-foreground">View pseudonym profile and activity</p>
        </div>
      </div>

      <div className="grid gap-6">

        {/* Pseudonym Information Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Pseudonym Information
            </CardTitle>
            <CardDescription>
              Basic pseudonym profile details and status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Display Name:</span>
                  <span>{pseudonymProfile.displayName}</span>
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Pseudonym ID:</span>
                  <span className="font-mono text-sm break-all">{pseudonymProfile.pseudonymId}</span>
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Slug:</span>
                  <span>{pseudonymProfile?.slug || 'N/A'}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Created:</span>
                  <span>{formatDate(pseudonymProfile.createdAt)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Last Active:</span>
                  <span>{pseudonymProfile?.lastActiveAt ? formatDate(pseudonymProfile.lastActiveAt) : 'N/A'}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Last Updated:</span>
                  <span>{pseudonymProfile?.updatedAt ? formatDate(pseudonymProfile.updatedAt) : 'N/A'}</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Status:</span>
                  <Badge variant={pseudonymProfile?.isActive ? "secondary" : "outline"}>
                    {pseudonymProfile?.isActive ? 'Active' : 'Inactive'}
                  </Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Karma Score:</span>
                  <Badge variant="outline">{pseudonymProfile.karmaScore}</Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Posts:</span>
                  <Badge variant="outline">{pseudonymProfile?.postCount || 0}</Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Comments:</span>
                  <Badge variant="outline">{pseudonymProfile?.commentCount || 0}</Badge>
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Show Karma:</span>
                  <span>{pseudonymProfile?.showKarma ? 'Yes' : 'No'}</span>
                </div>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Allow DMs:</span>
                  <span>{pseudonymProfile?.allowDirectMessages ? 'Yes' : 'No'}</span>
                </div>
              </div>
            </div>
          
          {/* Additional fields */}
          {pseudonymProfile?.bio && (
            <div className="mt-4 pt-4 border-t">
              <div className="flex items-center gap-2 mb-2">
                <User className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">Bio</span>
              </div>
              <p className="text-sm text-foreground">{pseudonymProfile.bio}</p>
            </div>
          )}
          
          {pseudonymProfile?.websiteUrl && (
            <div className="mt-4 pt-4 border-t">
              <div className="flex items-center gap-2 mb-2">
                <User className="h-4 w-4 text-muted-foreground" />
                <span className="font-medium">Website</span>
              </div>
              <a 
                href={pseudonymProfile.websiteUrl} 
                target="_blank" 
                rel="noopener noreferrer"
                className="text-primary hover:underline text-sm"
              >
                {pseudonymProfile.websiteUrl}
              </a>
            </div>
          )}


        </CardContent>
      </Card>

      {/* Recent Activity */}
      <Card>
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
    </div>
  )
}
