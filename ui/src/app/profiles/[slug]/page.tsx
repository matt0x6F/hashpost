'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/shadcn/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/card';
import { Badge } from '@/components/shadcn/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/shadcn/avatar';
import { Button } from '@/components/shadcn/button';
import { ExternalLink, Globe, MessageSquare, FileText, Calendar, Clock, Edit, TrendingUp } from 'lucide-react';
import { pseudonymsApi, extractApiErrorMessage } from '@/lib/api-client';
import { GetPostsByPseudonymSortEnum, GetCommentsByPseudonymSortEnum } from '@/generated/api/src/apis/PseudonymsApi';
import { useAuth } from '@/lib/auth-context';

interface PseudonymProfile {
  pseudonymId: string;
  displayName: string;
  karmaScore: number;
  createdAt: string;
  lastActiveAt: string;
  isActive: boolean;
  bio: string;
  websiteUrl: string;
  showKarma: boolean;
  allowDirectMessages: boolean;
  postCount: number;
  commentCount: number;
  slug: string;
}

import type { Post as APIPost, Comment as APIComment } from '@/generated/api/src/models';

// Use the API types directly since they already have the correct structure
type Post = APIPost;
type Comment = APIComment;

export default function ProfilePage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const slug = params.slug as string;
  const [profile, setProfile] = useState<PseudonymProfile | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isOwnProfile, setIsOwnProfile] = useState(false);
  const [postsLoading, setPostsLoading] = useState(false);
  const [commentsLoading, setCommentsLoading] = useState(false);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        setLoading(true);
        const profileData = await pseudonymsApi.getPseudonymProfileBySlug(slug);
        setProfile(profileData);
        
        // Check if this is the user's own profile
        if (user && user.pseudonyms.some(p => p.pseudonymId === profileData.pseudonymId)) {
          setIsOwnProfile(true);
        }
      } catch (err) {
        setError('Failed to load profile');
        console.error('Error fetching profile:', err);
      } finally {
        setLoading(false);
      }
    };

    if (slug) {
      fetchProfile();
    }
  }, [slug]);

  // Fetch posts and comments when profile loads
  useEffect(() => {
    if (profile) {
      fetchPosts();
      fetchComments();
    }
  }, [profile]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getCommunityPrefix = (communityType: string) => {
    switch (communityType) {
      case 't': return 't/';
      case 'g': return 'g/';
      case 'b': return 'b/';
      case 'c': return 'c/';
      case 'h': return 'h/';
      default: return 'r/';
    }
  };

  const formatTimeAgo = (dateString: string) => {
    if (!dateString) return 'unknown time ago';
    
    const now = new Date();
    const date = new Date(dateString);
    
    // Check if the date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid date string:', dateString);
      return 'unknown time ago';
    }
    
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
    return `${Math.floor(diffInSeconds / 86400)}d ago`;
  };

  const fetchPosts = async () => {
    if (!profile) return;
    try {
      setPostsLoading(true);
      const response = await pseudonymsApi.getPostsByPseudonym(slug, 1, 25, GetPostsByPseudonymSortEnum.CREATED_AT);
      setPosts(response.posts || []);
    } catch (err) {
      console.error('Error fetching posts:', err);
    } finally {
      setPostsLoading(false);
    }
  };

  const fetchComments = async () => {
    if (!profile) return;
    try {
      setCommentsLoading(true);
      const response = await pseudonymsApi.getCommentsByPseudonym(slug, 1, 25, GetCommentsByPseudonymSortEnum.CREATED_AT);
      setComments(response.comments || []);
    } catch (err) {
      console.error('Error fetching comments:', err);
    } finally {
      setCommentsLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="animate-pulse">
          <div className="h-32 bg-muted rounded-lg mb-6"></div>
          <div className="space-y-4">
            <div className="h-4 bg-muted rounded w-1/4"></div>
            <div className="h-4 bg-muted rounded w-1/2"></div>
            <div className="h-4 bg-muted rounded w-3/4"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <Card>
          <CardContent className="pt-6">
            <div className="text-center">
              <h2 className="text-2xl font-bold text-muted-foreground mb-2">Profile Not Found</h2>
              <p className="text-muted-foreground">The profile you&apos;re looking for doesn&apos;t exist or is no longer available.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl">
      {/* Profile Header */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-start space-x-4">
            <Avatar className="h-20 w-20">
              <AvatarFallback className="text-lg">
                {profile.displayName.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1">
              <div className="flex items-center space-x-2 mb-2">
                <h1 className="text-3xl font-bold">{profile.displayName}</h1>
                {profile.showKarma && (
                  <Badge variant="secondary">
                    {profile.karmaScore} karma
                  </Badge>
                )}
                {isOwnProfile && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => router.push(`/profiles/${slug}/edit`)}
                  >
                    <Edit className="w-4 h-4 mr-2" />
                    Edit Profile
                  </Button>
                )}
              </div>
              {profile.bio && (
                <p className="text-muted-foreground mb-3">{profile.bio}</p>
              )}
              <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                <div className="flex items-center space-x-1">
                  <Calendar className="h-4 w-4" />
                  <span>Joined {formatDate(profile.createdAt)}</span>
                </div>
                <div className="flex items-center space-x-1">
                  <Clock className="h-4 w-4" />
                  <span>Last active {formatTimeAgo(profile.lastActiveAt)}</span>
                </div>
                {profile.websiteUrl && (
                  <Button variant="ghost" size="sm" asChild>
                    <a href={profile.websiteUrl} target="_blank" rel="noopener noreferrer">
                      <Globe className="h-4 w-4 mr-1" />
                      Website
                    </a>
                  </Button>
                )}
              </div>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-2xl font-bold">{profile.postCount}</p>
                <p className="text-sm text-muted-foreground">Posts</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <MessageSquare className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-2xl font-bold">{profile.commentCount}</p>
                <p className="text-sm text-muted-foreground">Comments</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <TrendingUp className="h-5 w-5 text-muted-foreground" />
              <div>
                <p className="text-2xl font-bold">{profile.karmaScore}</p>
                <p className="text-sm text-muted-foreground">Karma</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Content Tabs */}
      <Tabs defaultValue="posts" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="posts" className="flex items-center space-x-2">
            <FileText className="h-4 w-4" />
            <span>Posts ({profile.postCount})</span>
          </TabsTrigger>
          <TabsTrigger value="comments" className="flex items-center space-x-2">
            <MessageSquare className="h-4 w-4" />
            <span>Comments ({profile.commentCount})</span>
          </TabsTrigger>
        </TabsList>
        
        <TabsContent value="posts" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Posts by {profile.displayName}</CardTitle>
            </CardHeader>
            <CardContent>
              {posts.length === 0 ? (
                <div className="text-center py-8">
                  <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No posts yet</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {posts.map((post) => (
                    <div key={post.postId} className="border-b border-border pb-4 last:border-b-0">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h3 className="font-semibold mb-1">
                            <a href={`/posts/${post.postId}`} className="hover:underline">
                              {post.title}
                            </a>
                          </h3>
                          <p className="text-sm text-muted-foreground mb-2">
                            Posted in {getCommunityPrefix(post.subforum?.communityType || '')}{post.subforum?.displayName || 'unknown'} • {formatTimeAgo(post.createdAt)}
                          </p>
                          <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                            <span>{post.score} points</span>
                            <span>{post.commentCount} comments</span>
                            <span>{post.viewCount || 0} views</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="comments" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Comments by {profile.displayName}</CardTitle>
            </CardHeader>
            <CardContent>
              {comments.length === 0 ? (
                <div className="text-center py-8">
                  <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No comments yet</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.commentId} className="border-b border-border pb-4 last:border-b-0">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <p className="text-sm text-muted-foreground mb-1">
                            Comment on &quot;{comment.postTitle || 'Unknown Post'}&quot; in {getCommunityPrefix(comment.communityType || '')}{comment.subforumDisplayName || 'unknown'}
                          </p>
                          <p className="mb-2">{comment.content}</p>
                          <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                            <span>{comment.score} points</span>
                            <span>{formatTimeAgo(comment.createdAt)}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
} 