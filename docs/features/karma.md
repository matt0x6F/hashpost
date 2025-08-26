# Karma System

## Overview

The karma system in Hashpost is a reputation mechanism that tracks the quality and engagement of user contributions through community voting. Karma represents the cumulative score of all posts and comments created by a pseudonym, reflecting their standing within the community.

## How Karma Works

### Calculation

Karma is calculated as the sum of all post and comment scores for a pseudonym:

```
Total Karma = Sum of Post Scores + Sum of Comment Scores
```

Where each post/comment score is:
```
Score = Upvotes - Downvotes
```

### Real-time Updates

Karma is automatically updated whenever:
- A post receives a vote (upvote/downvote/removal)
- A comment receives a vote (upvote/downvote/removal)
- Content is removed by moderators
- Content is restored by moderators

This ensures karma always reflects the current state of community feedback.

## Karma Calculation Details

### Content Filtering

The karma calculation automatically filters out:
- **Moderator-removed content**: Posts/comments removed by moderators (`is_removed = true`)

**Voting restrictions**: Deleted content cannot receive further votes to prevent manipulation.

**Karma preservation**: When users delete their own content, the karma contribution is preserved.

### Score Sources

#### Posts
- **Starting score**: 0 (new posts begin with neutral karma)
- **Upvote**: +1 to score
- **Downvote**: -1 to score
- **Vote removal**: Restores previous score

#### Comments
- **Starting score**: 0 (new comments begin with neutral karma)
- **Upvote**: +1 to score
- **Downvote**: -1 to score
- **Vote removal**: Restores previous score

## Implementation

### Karma Update Triggers

The system automatically calls `UpdateKarmaForPseudonym()` after:

1. **Voting operations**: When posts or comments receive upvotes/downvotes
2. **Profile views**: When a user views their profile (for lazy updates)

**Note**: Karma is NOT automatically updated when:
- Content is removed by moderators (moderator removal doesn't affect karma)

This design choice ensures that:
- Moderator actions don't penalize users unfairly
- Karma updates happen when they actually matter (during voting)

### Performance Considerations

- Karma updates are performed asynchronously to avoid blocking user interactions
- Updates are batched when possible to minimize database calls
- Failed karma updates don't break the main functionality (graceful degradation)

## User Experience

### Profile Display

Karma is prominently displayed on user profiles alongside:
- Post count
- Comment count
- Account creation date
- Last activity timestamp

### Privacy Controls

Users can control karma visibility:
- **Show karma**: Public display of karma score
- **Hide karma**: Karma remains hidden from other users
- **Default**: Karma is visible by default

### Karma Thresholds

While not currently implemented, the system is designed to support:
- **Posting restrictions**: Minimum karma requirements for certain actions
- **Moderation privileges**: Karma-based access to moderation tools
- **Community features**: Karma-gated access to premium features

## Moderation Integration

### Content Removal Impact

When content is removed by moderators:
- **Immediate effect**: Content is hidden from public view
- **Karma impact**: Removed content no longer contributes to karma
- **Audit trail**: Removal reason and moderator are recorded

### User Deletion Impact

When content is deleted by the author:
- **Visibility**: Content is hidden from public view
- **Voting restriction**: No further votes can be cast on deleted content
- **Audit trail**: Deletion reason and timestamp are recorded

### Restoration Process

When removed content is restored:
- **Visibility restored**: Content becomes visible again
- **Karma restored**: Content scores are recalculated into karma
- **History preserved**: Removal/restoration history is maintained

## Best Practices

### For Users

1. **Quality over quantity**: Focus on creating valuable content rather than gaming the system
2. **Engage constructively**: Thoughtful comments often receive positive feedback
3. **Respect community norms**: Following guidelines helps maintain positive karma

### For Moderators

1. **Consistent enforcement**: Apply rules fairly to maintain system integrity
2. **Document decisions**: Clear removal reasons help users understand expectations
3. **Consider context**: Not all rule violations require content removal

### For Developers

1. **Monitor performance**: Karma updates should not impact user experience
2. **Validate calculations**: Regular audits ensure karma accuracy
3. **Handle edge cases**: Graceful degradation when karma updates fail

## Future Enhancements

### Planned Features

- **Karma decay**: Gradual reduction of old karma over time
- **Weighted voting**: Different user types having different vote weights
- **Karma badges**: Recognition for reaching milestone karma levels
- **Community-specific karma**: Separate karma tracking per subforum

### Technical Improvements

- **Caching layer**: Redis-based karma caching for improved performance
- **Batch processing**: Bulk karma updates for efficiency
- **Analytics dashboard**: Detailed karma statistics and trends
- **API endpoints**: Programmatic access to karma data

## Troubleshooting

### Common Issues

#### Karma Not Updating
- Check if content is properly filtered (moderator-removed content doesn't count)
- Verify vote operations are completing successfully
- Ensure database connections are stable

#### Inconsistent Karma
- Run manual karma recalculation for affected users
- Check for orphaned votes or content
- Verify moderation actions are properly recorded

#### Performance Issues
- Monitor karma update query performance
- Consider implementing karma update queuing
- Review database indexing on karma-related tables

### Debugging Tools

- **Manual recalculation**: `UpdateKarmaForPseudonym()` can be called directly
- **Log analysis**: Karma updates are logged with detailed context
- **Database queries**: Direct inspection of karma calculation queries
- **Test coverage**: Comprehensive tests for karma edge cases

## Conclusion

The karma system provides a robust foundation for community reputation management. By automatically tracking user contributions and community feedback, it encourages quality content creation while maintaining system integrity. The real-time updates ensure accuracy, while the comprehensive filtering prevents manipulation.

As the platform evolves, the karma system will continue to adapt, providing users with meaningful feedback and communities with effective moderation tools.

## Testing Karma Locally

### Manual Testing

To test karma calculation locally:

1. **Create test content**: Create posts and comments with your test pseudonym
2. **Vote on content**: Have other users (or different pseudonyms) vote on your content
3. **Check karma updates**: View your profile to trigger karma recalculation

### Debugging Karma Issues

If karma isn't updating as expected:

1. **Check database state**:
   ```sql
   -- Check current karma score
   SELECT karma_score FROM pseudonyms WHERE id = 'your-pseudonym-id';
   
   -- Check post scores
   SELECT id, score, is_deleted, is_removed FROM posts WHERE pseudonym_id = 'your-pseudonym-id';
   
   -- Check comment scores  
   SELECT id, score, is_deleted, is_removed FROM comments WHERE pseudonym_id = 'your-pseudonym-id';
   ```

2. **Verify vote operations**: Check that votes are being recorded correctly
3. **Check content state**: Ensure content isn't marked as removed by moderators
4. **Manual recalculation**: Use the admin interface to manually recalculate karma

### Common Issues
