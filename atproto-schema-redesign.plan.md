# Atproto Schema Redesign and Frontend Integration Plan

## Overview

Redesign the API schema to be atproto-compliant by separating immutable records from computed/stateful data, then update the frontend to work with the new architecture.

## Key Principles

1. **Immutable Records**: Core content (posts, comments) stored as atproto records
2. **Computed Data**: Vote counts, scores, user states fetched separately
3. **Event-Driven**: Voting creates separate vote records, not mutable fields
4. **Session-Specific**: User vote state not stored in post records

## Implementation Plan

### Phase 1: Redesign OpenAPI Schema

#### 1.1 Update Post Schema (Atproto-Compliant)

**Current (Problematic):**
```yaml
Post:
  properties:
    score: integer          # ❌ Mutable state
    upvotes: integer        # ❌ Mutable state  
    downvotes: integer      # ❌ Mutable state
    userVote: integer       # ❌ Session-specific
    isSticky: boolean       # ❌ Moderation state
    isRemoved: boolean      # ❌ Moderation state
```

**New (Atproto-Compliant):**
```yaml
Post:
  properties:
    id: string              # ✅ atproto record URI
    author: string          # ✅ DID
    subforum: string        # ✅ Subforum identifier
    title: string           # ✅ Content
    content: string         # ✅ Content
    created_at: string      # ✅ Record timestamp
    updated_at: string      # ✅ Last modification
    # Remove: score, upvotes, downvotes, userVote, isSticky, isRemoved
```

#### 1.2 Add Computed Data Endpoints

**New Endpoints:**
```yaml
/api/v1/posts/{id}/metrics:
  get:
    summary: Get post metrics
    description: Get computed vote counts and score
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PostMetrics'

/api/v1/posts/{id}/user-vote:
  get:
    summary: Get user's vote on post
    description: Get the authenticated user's current vote
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserVote'

/api/v1/posts/{id}/moderation:
  get:
    summary: Get post moderation state
    description: Get moderation flags (pinned, locked, etc.)
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ModerationState'
```

**New Schemas:**
```yaml
PostMetrics:
  type: object
  properties:
    upvotes: integer
    downvotes: integer
    score: integer
    comment_count: integer

UserVote:
  type: object
  properties:
    vote_type: string
    enum: [up, down, null]

ModerationState:
  type: object
  properties:
    is_pinned: boolean
    is_locked: boolean
    is_removed: boolean
```

#### 1.3 Implement Vote Records System

**Vote Record Schema:**
```yaml
VoteRecord:
  type: object
  required:
    - id
    - voter
    - target
    - vote_type
    - created_at
  properties:
    id: string              # atproto record URI
    voter: string          # DID of voter
    target: string         # Post/comment ID
    vote_type: string      # up, down
    created_at: string     # Record timestamp
```

**Voting Endpoints:**
```yaml
/api/v1/posts/{id}/vote:
  post:
    summary: Vote on post
    description: Create a vote record
    requestBody:
      content:
        application/json:
          schema:
            type: object
            properties:
              vote_type:
                type: string
                enum: [up, down]
  responses:
    '200':
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/VoteRecord'
```

### Phase 2: Update Backend Implementation

#### 2.1 Separate Record Storage from Computed Data

**Backend Changes:**
- Store posts as atproto records (immutable)
- Store votes as separate atproto records
- Compute metrics on-demand from vote records
- Store moderation state separately from post records

#### 2.2 Implement Metrics Computation

**Metrics Service:**
```go
type MetricsService struct {
    voteRepo VoteRepository
}

func (s *MetricsService) GetPostMetrics(postID string) (*PostMetrics, error) {
    votes, err := s.voteRepo.GetVotesForPost(postID)
    if err != nil {
        return nil, err
    }
    
    upvotes := 0
    downvotes := 0
    for _, vote := range votes {
        if vote.VoteType == "up" {
            upvotes++
        } else {
            downvotes++
        }
    }
    
    return &PostMetrics{
        Upvotes: upvotes,
        Downvotes: downvotes,
        Score: upvotes - downvotes,
    }, nil
}
```

### Phase 3: Update Frontend Architecture

#### 3.1 Create Data Fetching Hooks

**Post Data Hook:**
```typescript
// ui/src/hooks/usePostData.ts
export function usePostData(postId: string) {
  const [post, setPost] = useState<Post | null>(null);
  const [metrics, setMetrics] = useState<PostMetrics | null>(null);
  const [userVote, setUserVote] = useState<UserVote | null>(null);
  const [moderation, setModeration] = useState<ModerationState | null>(null);
  
  const fetchPostData = async () => {
    const [postRes, metricsRes, voteRes, modRes] = await Promise.all([
      postsApi.getPost(postId),
      postsApi.getPostMetrics(postId),
      postsApi.getUserVote(postId),
      postsApi.getModerationState(postId)
    ]);
    
    setPost(postRes);
    setMetrics(metricsRes);
    setUserVote(voteRes);
    setModeration(modRes);
  };
  
  return { post, metrics, userVote, moderation, refetch: fetchPostData };
}
```

#### 3.2 Update PostCard Component

**New PostCard Structure:**
```typescript
// ui/src/components/PostCard.tsx
export function PostCard({ postId }: { postId: string }) {
  const { post, metrics, userVote, moderation, refetch } = usePostData(postId);
  const [isVoting, setIsVoting] = useState(false);
  
  const handleVote = async (voteType: 'up' | 'down' | null) => {
    setIsVoting(true);
    try {
      if (voteType) {
        await votingApi.voteOnPost(postId, { vote_type: voteType });
      } else {
        await votingApi.removeVoteFromPost(postId);
      }
      await refetch(); // Refresh all data
    } finally {
      setIsVoting(false);
    }
  };
  
  if (!post || !metrics) return <Loading />;
  
  return (
    <div>
      <h3>{post.title}</h3>
      <p>{post.content}</p>
      <div>
        <Button onClick={() => handleVote(userVote?.vote_type === 'up' ? null : 'up')}>
          ↑ {metrics.upvotes}
        </Button>
        <span>{metrics.score}</span>
        <Button onClick={() => handleVote(userVote?.vote_type === 'down' ? null : 'down')}>
          ↓ {metrics.downvotes}
        </Button>
      </div>
    </div>
  );
}
```

#### 3.3 Update PostList Component

**New PostList Structure:**
```typescript
// ui/src/components/PostList.tsx
export function PostList({ subforumSlug }: { subforumSlug: string }) {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  
  const loadPosts = async () => {
    setLoading(true);
    try {
      const response = await postsApi.listPosts(subforumSlug, 20, 0);
      setPosts(response.posts || []);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div>
      {posts.map(post => (
        <PostCard key={post.id} postId={post.id} />
      ))}
    </div>
  );
}
```

### Phase 4: Update Authentication & Capabilities

#### 4.1 Enhance Capabilities System

**Update capabilities.ts:**
```typescript
// ui/src/lib/capabilities.ts
export class CapabilitiesService {
  async canVoteOnPost(postId: string): Promise<boolean> {
    const capabilities = await this.getUserCapabilities();
    return capabilities.permissions.includes('vote') && 
           capabilities.permissions.includes('create_posts');
  }
  
  async canModeratePost(postId: string): Promise<boolean> {
    const capabilities = await this.getUserCapabilities();
    return capabilities.permissions.includes('moderate_content');
  }
}
```

#### 4.2 Update Auth Context

**Enhanced auth-context.tsx:**
```typescript
// ui/src/lib/auth-context.tsx
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [capabilities, setCapabilities] = useState<UserCapabilities | null>(null);
  
  const refreshCapabilities = async () => {
    if (user) {
      const caps = await capabilitiesService.getUserCapabilities();
      setCapabilities(caps);
    }
  };
  
  // Refresh capabilities when user changes
  useEffect(() => {
    refreshCapabilities();
  }, [user]);
  
  return (
    <AuthContext.Provider value={{ user, capabilities, refreshCapabilities }}>
      {children}
    </AuthContext.Provider>
  );
}
```

### Phase 5: Testing & Validation

#### 5.1 Test Atproto Compliance

**Validation Checklist:**
- [ ] Post records are immutable (no mutable fields)
- [ ] Vote records are separate atproto records
- [ ] Metrics are computed from vote records
- [ ] User vote state is session-specific
- [ ] Moderation state is separate from post records
- [ ] All data is content-addressed and verifiable

#### 5.2 Test Frontend Integration

**Integration Tests:**
- [ ] PostCard displays post data correctly
- [ ] Voting updates metrics in real-time
- [ ] User vote state persists across page loads
- [ ] Moderation controls work for authorized users
- [ ] Capabilities system restricts actions appropriately

## Implementation Timeline

### Week 1: Backend Schema Redesign
- Update OpenAPI spec with new schema
- Implement vote records system
- Add metrics computation endpoints
- Update backend handlers

### Week 2: Frontend Architecture Update
- Create data fetching hooks
- Update PostCard component
- Update PostList component
- Implement capabilities integration

### Week 3: Testing & Refinement
- Test atproto compliance
- Fix integration issues
- Performance optimization
- Documentation updates

## Success Criteria

1. **Atproto Compliance**: All data stored as immutable, verifiable records
2. **Performance**: Metrics computed efficiently from vote records
3. **User Experience**: Real-time updates without page refreshes
4. **Security**: Proper capability-based access control
5. **Maintainability**: Clean separation of concerns between records and computed data

## Migration Strategy

1. **Backward Compatibility**: Keep old endpoints during transition
2. **Gradual Migration**: Update components one by one
3. **Feature Flags**: Enable new architecture per component
4. **Rollback Plan**: Ability to revert to old system if needed
