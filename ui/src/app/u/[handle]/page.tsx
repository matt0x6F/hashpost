'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Button } from '@/components/shadcn/button';
import { Badge } from '@/components/shadcn/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/shadcn/tabs';
import { 
  User, 
  Mail, 
  Calendar, 
  MessageSquare, 
  FileText,
  Settings,
  Lock,
  Users,
  Eye
} from 'lucide-react';
import { getApiWithRefresh } from '@/lib/api-client';
import { ProfilesApi } from '@/generated/api/src/apis/ProfilesApi';
import { UserProfile, UserPost, UserComment } from '@/generated/api/src/models';
import { PostCard } from '@/components/PostCard';
import Comment from '@/components/Comment';
import { useAuth } from '@/lib/auth-context';
import { toast } from 'sonner';
import Link from 'next/link';

interface ProfilePageProps {}

export default function ProfilePage({}: ProfilePageProps) {
  const params = useParams();
  const handle = params.handle as string;
  const { user: currentUser, isLoading: authLoading } = useAuth();
  
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [posts, setPosts] = useState<UserPost[]>([]);
  const [comments, setComments] = useState<UserComment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('posts');
  const [postsLoading, setPostsLoading] = useState(false);
  const [commentsLoading, setCommentsLoading] = useState(false);

  const isOwnProfile = currentUser?.handle === handle;

  useEffect(() => {
    if (handle && !authLoading) {
      loadProfile();
    }
  }, [handle, authLoading]);

  const loadProfile = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const api = getApiWithRefresh(ProfilesApi);
      const profileData = await api.getUserProfile(handle);
      setProfile(profileData);
      
      // Load initial posts
      await loadPosts();
    } catch (err: unknown) {
      console.error('Failed to load profile:', err);
      
      if (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'status' in err.response) {
        const status = (err.response as { status: number }).status;
        if (status === 401) {
          setError("Please log in to view this profile");
        } else if (status === 403) {
          // Check if user is viewing their own profile
          console.log('403 error - debugging info:', {
            currentUserHandle: currentUser?.handle,
            profileHandle: handle,
            isOwnProfile,
            authLoading
          });
          
          if (isOwnProfile) {
            // If it's their own profile, this shouldn't happen since backend allows owner access
            // Log warning but don't show error - might be auth issue
            console.warn('Profile owner received 403 error - possible auth issue');
            setError("Authentication issue - please try refreshing the page");
          } else {
            setError("This profile is private");
          }
        } else if (status === 404) {
          setError("User not found");
        } else {
          setError("Failed to load profile");
        }
      } else {
        setError("Failed to load profile");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const loadPosts = async () => {
    if (!handle) return;
    
    setPostsLoading(true);
    try {
      const api = getApiWithRefresh(ProfilesApi);
      const response = await api.getUserPosts(handle, { limit: 20, offset: 0 });
      setPosts(response.posts || []);
    } catch (err) {
      console.error('Failed to load posts:', err);
      
      // Handle visibility errors for posts
      if (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'status' in err.response) {
        const status = (err.response as { status: number }).status;
        if (status === 403 && !isOwnProfile) {
          // For non-owners, this is expected for private profiles
          console.log('Posts not accessible due to private profile');
        } else if (status === 403 && isOwnProfile) {
          console.warn('Profile owner cannot access own posts - possible auth issue');
        }
      }
      // Don't show toast error for visibility-related failures
    } finally {
      setPostsLoading(false);
    }
  };

  const loadComments = async () => {
    if (!handle) return;
    
    setCommentsLoading(true);
    try {
      const api = getApiWithRefresh(ProfilesApi);
      const response = await api.getUserComments(handle, { limit: 20, offset: 0 });
      setComments(response.comments || []);
    } catch (err) {
      console.error('Failed to load comments:', err);
      
      // Handle visibility errors for comments
      if (err && typeof err === 'object' && 'response' in err && err.response && typeof err.response === 'object' && 'status' in err.response) {
        const status = (err.response as { status: number }).status;
        if (status === 403 && !isOwnProfile) {
          // For non-owners, this is expected for private profiles
          console.log('Comments not accessible due to private profile');
        } else if (status === 403 && isOwnProfile) {
          console.warn('Profile owner cannot access own comments - possible auth issue');
        }
      }
      // Don't show toast error for visibility-related failures
    } finally {
      setCommentsLoading(false);
    }
  };

  const handleTabChange = (value: string) => {
    setActiveTab(value);
    if (value === 'comments' && comments.length === 0) {
      loadComments();
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const getVisibilityIcon = (visibility: string) => {
    switch (visibility) {
      case 'public':
        return <Eye className="h-4 w-4" />;
      case 'authenticated':
        return <Users className="h-4 w-4" />;
      case 'private':
        return <Lock className="h-4 w-4" />;
      default:
        return <Eye className="h-4 w-4" />;
    }
  };

  const getVisibilityLabel = (visibility: string) => {
    switch (visibility) {
      case 'public':
        return 'Public';
      case 'authenticated':
        return 'Authenticated users only';
      case 'private':
        return 'Private';
      default:
        return 'Public';
    }
  };

  if (isLoading || authLoading) {
    return (
      <div className="container mx-auto py-6 max-w-4xl">
        <div className="flex items-center justify-center min-h-64">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-6 max-w-4xl">
        <Card>
          <CardContent className="p-8 text-center">
            <div className="text-muted-foreground mb-4">{error}</div>
            {error === "Please log in to view this profile" && (
              <Button asChild>
                <Link href="/auth/login">Log In</Link>
              </Button>
            )}
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="container mx-auto py-6 max-w-4xl">
        <Card>
          <CardContent className="p-8 text-center">
            <div className="text-muted-foreground">Profile not found</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 max-w-4xl">
      {/* Profile Header */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center space-x-4">
              <div className="w-20 h-20 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-2xl font-bold">
                {profile.displayName ? profile.displayName.charAt(0).toUpperCase() : profile.handle.charAt(0).toUpperCase()}
              </div>
              <div>
                <CardTitle className="text-2xl">
                  {profile.displayName || profile.handle}
                </CardTitle>
                <CardDescription className="text-lg">
                  @{profile.handle}
                </CardDescription>
                {profile.bio && (
                  <p className="mt-2 text-sm text-muted-foreground">{profile.bio}</p>
                )}
              </div>
            </div>
            
            <div className="flex items-center space-x-2">
              <Badge variant="outline" className="flex items-center gap-1">
                {getVisibilityIcon(profile.profileVisibility)}
                {getVisibilityLabel(profile.profileVisibility)}
              </Badge>
              {isOwnProfile && (
                <Button asChild variant="outline" size="sm">
                  <Link href={`/u/${handle}/edit`}>
                    <Settings className="h-4 w-4 mr-2" />
                    Edit Profile
                  </Link>
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold">{profile.postCount}</div>
              <div className="text-sm text-muted-foreground">Posts</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{profile.commentCount}</div>
              <div className="text-sm text-muted-foreground">Comments</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">{profile.reputation}</div>
              <div className="text-sm text-muted-foreground">Reputation</div>
            </div>
            <div className="text-center">
              <div className="text-sm text-muted-foreground">
                Joined {formatDate(profile.createdAt)}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Content Tabs */}
      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="posts" className="flex items-center gap-2">
            <FileText className="h-4 w-4" />
            Posts ({profile.postCount})
          </TabsTrigger>
          <TabsTrigger value="comments" className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Comments ({profile.commentCount})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="posts" className="mt-6">
          {postsLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : posts.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <FileText className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <div className="text-muted-foreground">No posts yet</div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {posts.map((post) => (
                <PostCard key={post.id} postId={post.id} />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="comments" className="mt-6">
          {commentsLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : comments.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <MessageSquare className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <div className="text-muted-foreground">No comments yet</div>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <Comment 
                  key={comment.id} 
                  comment={{
                    id: comment.id,
                    content: comment.content,
                    author: {
                      did: comment.authorDid,
                      handle: comment.authorHandle,
                      displayName: comment.authorDisplayName || null,
                      avatarUrl: comment.authorAvatarUrl || null,
                    },
                    createdAt: comment.createdAt,
                    upvotes: comment.upvotes,
                    downvotes: comment.downvotes,
                    score: comment.score,
                    postId: comment.postId,
                    parentId: comment.parentId,
                  }}
                  postId={comment.postId}
                />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
