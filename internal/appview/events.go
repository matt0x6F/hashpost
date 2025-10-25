package appview

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	"github.com/matt0x6f/hashpost/internal/lexicons"
	"github.com/nats-io/nats.go"
)

// EventConsumer handles consuming atproto events from NATS Jetstream
type EventConsumer struct {
	nc               *nats.Conn
	js               nats.JetStreamContext
	logger           *slog.Logger
	db               *Database
	identityResolver *IdentityResolver
	profileFetcher   *ProfileFetcher

	// Enhanced error handling
	maxRetries    int
	retryDelay    time.Duration
	maxRetryDelay time.Duration
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(natsURL string, db *Database, logger *slog.Logger) (*EventConsumer, error) {
	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create identity resolver
	identityResolver := NewIdentityResolver(logger)

	// Create profile fetcher
	profileFetcher := NewProfileFetcher(db.queries, logger)

	consumer := &EventConsumer{
		nc:               nc,
		js:               js,
		logger:           logger,
		db:               db,
		identityResolver: identityResolver,
		profileFetcher:   profileFetcher,

		// Enhanced error handling configuration
		maxRetries:    3,
		retryDelay:    1 * time.Second,
		maxRetryDelay: 30 * time.Second,
	}

	return consumer, nil
}

// StartConsuming starts consuming events from the stream
func (ec *EventConsumer) StartConsuming(ctx context.Context) error {
	// Subscribe to all hashpost events
	subject := "hashpost.events.*" // Subscribe to all event types
	streamName := "HASHPOST_EVENTS"

	ec.logger.Info("Subscribing to atproto events", "subject", subject, "stream", streamName)

	// Create pull subscription with explicit stream binding
	sub, err := ec.js.PullSubscribe(subject, "hashpost-appview",
		nats.BindStream(streamName),
	)
	if err != nil {
		ec.logger.Error("Failed to create subscription", "error", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Log subscription details for debugging
	ec.logger.Info("Created NATS subscription",
		"subject", subject,
		"stream", streamName,
		"consumer", "hashpost-appview",
	)

	ec.logger.Info("Started consuming atproto events", "subject", subject)

	// Process messages in a loop
	emptyFetchCount := 0
	for {
		select {
		case <-ctx.Done():
			ec.logger.Info("Stopping event consumer")
			sub.Unsubscribe()
			return nil
		default:
			// Fetch messages
			msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					// No messages, continue
					emptyFetchCount++
					// Only log every 10th empty fetch to reduce spam
					if emptyFetchCount%10 == 0 {
						ec.logger.Debug("No messages received, continuing...", "empty_fetches", emptyFetchCount)
					}
					continue
				}
				ec.logger.Error("Failed to fetch messages", "error", err)
				continue
			}

			// Reset empty fetch counter when we get messages
			emptyFetchCount = 0
			ec.logger.Info("Received messages from NATS", "count", len(msgs))

			// Process each message with enhanced error handling
			for _, msg := range msgs {
				if err := ec.processMessageWithRetry(msg); err != nil {
					ec.logger.Error("Failed to process message after retries", "error", err)
					// Send to dead letter queue for manual intervention
					ec.sendToDeadLetterQueue(msg, err)
					continue
				}

				// Acknowledge the message
				if err := msg.Ack(); err != nil {
					ec.logger.Error("Failed to ack message", "error", err)
				}
			}
		}
	}
}

// processMessageWithRetry processes a message with retry logic
func (ec *EventConsumer) processMessageWithRetry(msg *nats.Msg) error {
	// Check for idempotency
	eventID := ec.generateEventID(msg)
	if ec.isEventProcessed(eventID) {
		ec.logger.Debug("Event already processed", "event_id", eventID)
		return nil
	}

	// Retry logic with exponential backoff
	for attempt := 0; attempt < ec.maxRetries; attempt++ {
		if err := ec.processMessage(msg); err != nil {
			if attempt == ec.maxRetries-1 {
				return fmt.Errorf("failed after %d attempts: %w", ec.maxRetries, err)
			}

			// Calculate exponential backoff delay
			delay := ec.calculateBackoffDelay(attempt)
			ec.logger.Warn("Message processing failed, retrying",
				"attempt", attempt+1,
				"max_retries", ec.maxRetries,
				"delay", delay,
				"error", err,
			)
			time.Sleep(delay)
			continue
		}

		// Success - mark as processed
		sequence := int64(0)
		if seqStr := msg.Header.Get("Nats-Sequence"); seqStr != "" {
			if seq, err := strconv.ParseInt(seqStr, 10, 64); err == nil {
				sequence = seq
			}
		}
		ec.markEventProcessed(eventID, msg.Subject, sequence)
		return nil
	}

	return fmt.Errorf("exhausted all retry attempts")
}

// generateEventID creates a unique identifier for an event
func (ec *EventConsumer) generateEventID(msg *nats.Msg) string {
	// Use subject + sequence for uniqueness
	sequence := msg.Header.Get("Nats-Sequence")
	if sequence == "" {
		sequence = "unknown"
	}
	return fmt.Sprintf("%s-%s", msg.Subject, sequence)
}

// isEventProcessed checks if an event has already been processed
func (ec *EventConsumer) isEventProcessed(eventID string) bool {
	exists, err := ec.db.queries.IsEventProcessed(context.Background(), eventID)
	if err != nil {
		ec.logger.Error("Failed to check if event is processed", "error", err, "event_id", eventID)
		return false // Assume not processed if we can't check
	}
	return exists
}

// markEventProcessed marks an event as processed
func (ec *EventConsumer) markEventProcessed(eventID string, subject string, sequence int64) {
	_, err := ec.db.queries.CreateProcessedEvent(context.Background(), &generated.CreateProcessedEventParams{
		EventID:  eventID,
		Subject:  subject,
		Sequence: sequence,
	})
	if err != nil {
		ec.logger.Error("Failed to mark event as processed", "error", err, "event_id", eventID)
		// Don't fail the event processing, just log the error
	}
}

// calculateBackoffDelay calculates exponential backoff delay
func (ec *EventConsumer) calculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(attempt+1) * ec.retryDelay
	if delay > ec.maxRetryDelay {
		delay = ec.maxRetryDelay
	}
	return delay
}

// sendToDeadLetterQueue sends failed messages to a dead letter queue
func (ec *EventConsumer) sendToDeadLetterQueue(msg *nats.Msg, err error) {
	dlqSubject := "hashpost.events.dlq"

	// Create DLQ message with original data and error info
	dlqData := map[string]interface{}{
		"original_subject": msg.Subject,
		"original_data":    string(msg.Data),
		"error":            err.Error(),
		"timestamp":        time.Now(),
	}

	dlqBytes, marshalErr := json.Marshal(dlqData)
	if marshalErr != nil {
		ec.logger.Error("Failed to marshal DLQ data", "error", marshalErr)
		return
	}

	// Publish to dead letter queue
	if _, pubErr := ec.js.Publish(dlqSubject, dlqBytes); pubErr != nil {
		ec.logger.Error("Failed to publish to dead letter queue", "error", pubErr)
		return
	}

	ec.logger.Error("Message sent to dead letter queue",
		"dlq_subject", dlqSubject,
		"original_subject", msg.Subject,
		"error", err,
	)
}

// processMessage processes a single NATS message
func (ec *EventConsumer) processMessage(msg *nats.Msg) error {
	ec.logger.Info("Processing NATS message", "subject", msg.Subject, "data_length", len(msg.Data))

	// Parse the event
	var event AtprotoEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		ec.logger.Error("Failed to unmarshal event", "error", err, "data", string(msg.Data))
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	ec.logger.Info("Processing atproto event",
		"type", event.Type,
		"repo", event.Repo,
		"collection", event.Collection,
		"subject", msg.Subject,
	)

	// Route event to appropriate handler
	switch event.Type {
	case "record.created":
		return ec.handleRecordCreated(event)
	case "record.updated":
		return ec.handleRecordUpdated(event)
	case "record.deleted":
		return ec.handleRecordDeleted(event)
	case "identity.resolved":
		return ec.handleIdentityResolved(event)
	case "session.created":
		return ec.handleSessionCreated(event)
	default:
		ec.logger.Warn("Unknown event type", "type", event.Type)
		return nil
	}
}

// AtprotoEvent represents an atproto event from the stream
type AtprotoEvent struct {
	Type       string                 `json:"type"`
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection,omitempty"`
	Record     map[string]interface{} `json:"record,omitempty"`
	URI        string                 `json:"uri,omitempty"`
	CID        string                 `json:"cid,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// handleRecordCreated processes record created events
func (ec *EventConsumer) handleRecordCreated(event AtprotoEvent) error {
	ec.logger.Info("Record created",
		"repo", event.Repo,
		"collection", event.Collection,
		"uri", event.URI,
	)

	// Process based on collection type
	switch event.Collection {
	case lexicons.CollectionFeedPost:
		return ec.handlePostCreated(event)
	case lexicons.CollectionFeedSubforum:
		return ec.handleSubforumCreated(event)
	case lexicons.CollectionFeedComment:
		return ec.handleCommentCreated(event)
	case "com.hashpost.feed.vote":
		return ec.handleVoteCreated(event)
	case "com.hashpost.graph.subscription":
		return ec.handleSubscriptionCreated(event)
	default:
		ec.logger.Debug("Unknown collection type", "collection", event.Collection)
	}

	return nil
}

// handleRecordUpdated processes record updated events
func (ec *EventConsumer) handleRecordUpdated(event AtprotoEvent) error {
	ec.logger.Info("Record updated",
		"repo", event.Repo,
		"collection", event.Collection,
		"uri", event.URI,
	)

	// Update denormalized data in AppView database based on collection type
	switch event.Collection {
	case lexicons.CollectionFeedPost:
		return ec.handlePostUpdated(event)
	case lexicons.CollectionFeedSubforum:
		return ec.handleSubforumUpdated(event)
	case lexicons.CollectionFeedComment:
		return ec.handleCommentUpdated(event)
	default:
		ec.logger.Debug("Unknown collection type for update", "collection", event.Collection)
	}

	return nil
}

// handleRecordDeleted processes record deleted events
func (ec *EventConsumer) handleRecordDeleted(event AtprotoEvent) error {
	ec.logger.Info("Record deleted",
		"repo", event.Repo,
		"collection", event.Collection,
		"uri", event.URI,
	)

	// Remove denormalized data from AppView database based on collection type
	switch event.Collection {
	case lexicons.CollectionFeedPost:
		return ec.handlePostDeleted(event)
	case lexicons.CollectionFeedSubforum:
		return ec.handleSubforumDeleted(event)
	case lexicons.CollectionFeedComment:
		return ec.handleCommentDeleted(event)
	default:
		ec.logger.Debug("Unknown collection type for deletion", "collection", event.Collection)
	}

	return nil
}

// handleIdentityResolved processes identity resolved events
func (ec *EventConsumer) handleIdentityResolved(event AtprotoEvent) error {
	ec.logger.Info("Identity resolved",
		"repo", event.Repo,
		"handle", event.Metadata["handle"],
	)

	// Update user information in AppView database
	did := event.Repo
	handle, ok := event.Metadata["handle"].(string)
	if !ok {
		ec.logger.Warn("Invalid handle in identity resolved event", "metadata", event.Metadata)
		return nil
	}

	// Create or update user in AppView database with basic info
	user := &AppViewUser{
		ID:          uuid.New(),
		DID:         did,
		Handle:      handle,
		DisplayName: handle, // Use handle as display name initially - will be updated by profile fetcher
		CreatedAt:   event.Timestamp,
		UpdatedAt:   event.Timestamp,
	}

	if err := ec.db.CreateUser(user); err != nil {
		ec.logger.Error("Failed to update user from identity resolution", "error", err, "did", did)
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Fetch full profile data from PDS (async, don't block event processing)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		profileData, err := ec.profileFetcher.FetchProfileFromPDS(ctx, did)
		if err != nil {
			ec.logger.Warn("Failed to fetch profile data", "error", err, "did", did)
			return
		}

		ec.logger.Info("Profile data fetched and cached", "did", did, "display_name", profileData.DisplayName)
	}()

	ec.logger.Info("User information updated from identity resolution", "did", did, "handle", handle)
	return nil
}

// handleSessionCreated processes session created events
func (ec *EventConsumer) handleSessionCreated(event AtprotoEvent) error {
	ec.logger.Info("Session created",
		"repo", event.Repo,
		"session_id", event.Metadata["session_id"],
	)

	// Update session information in AppView database
	sessionID, ok := event.Metadata["session_id"].(string)
	if !ok {
		ec.logger.Warn("Invalid session_id in session created event", "metadata", event.Metadata)
		return nil
	}

	// Resolve handle from DID
	handle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve handle for session", "error", err, "did", event.Repo)
		handle = "unknown"
	}

	// Create session record in AppView database
	// Note: This would require a sessions table in AppView database
	// For now, we'll just log the session creation
	ec.logger.Info("Session tracked in AppView", "session_id", sessionID, "did", event.Repo, "handle", handle)
	return nil
}

// handlePostCreated processes HashPost feed post creation events
func (ec *EventConsumer) handlePostCreated(event AtprotoEvent) error {
	ec.logger.Info("HashPost feed post created",
		"repo", event.Repo,
		"uri", event.URI,
		"text", event.Record["text"],
	)

	// Extract post data from the record
	text, ok := event.Record["text"].(string)
	if !ok {
		return fmt.Errorf("invalid text field in post record")
	}

	createdAtStr, ok := event.Record["createdAt"].(string)
	if !ok {
		return fmt.Errorf("invalid createdAt field in post record")
	}

	// Try multiple date formats
	var createdAt time.Time
	var err error

	// Try RFC3339 first
	createdAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		// Try RFC3339Nano
		createdAt, err = time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			// Try custom format with milliseconds
			createdAt, err = time.Parse("2006-01-02T15:04:05.000Z", createdAtStr)
			if err != nil {
				// Fallback to current time
				createdAt = time.Now()
				ec.logger.Warn("Failed to parse createdAt, using current time",
					"createdAt", createdAtStr, "error", err)
			}
		}
	}

	// Resolve author handle from DID
	authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
		authorHandle = "unknown"
	}

	// Extract subforum slug from record context or use default
	subforumSlug := "general" // Default subforum
	if subforum, ok := event.Record["subforum"].(string); ok && subforum != "" {
		subforumSlug = subforum
	}

	// Create AppView post data
	post := &AppViewPost{
		ID:           uuid.New(),
		AtprotoURI:   event.URI,
		AuthorDID:    event.Repo,
		AuthorHandle: authorHandle,
		SubforumSlug: subforumSlug,
		Title:        text, // For now, use text as title
		Content:      text,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}

	// Store in AppView database
	if err := ec.db.CreatePost(post); err != nil {
		return fmt.Errorf("failed to create post in AppView: %w", err)
	}

	// Update subforum post count
	if err := ec.db.queries.UpdateSubforumPostCount(context.Background(), &generated.UpdateSubforumPostCountParams{
		Slug:      post.SubforumSlug,
		PostCount: &[]int32{1}[0],
	}); err != nil {
		ec.logger.Warn("Failed to update subforum post count", "error", err, "subforum", post.SubforumSlug)
		// Don't fail the event processing, just log the error
	}

	// Update user's last seen time
	if err := ec.db.queries.UpdateUserLastSeen(context.Background(), event.Repo); err != nil {
		ec.logger.Warn("Failed to update user last seen time", "error", err, "did", event.Repo)
	}

	ec.logger.Info("Post stored in AppView database",
		"uri", event.URI,
		"text", text,
		"created_at", createdAt,
	)

	return nil
}

// handleSubforumCreated processes HashPost subforum creation events
func (ec *EventConsumer) handleSubforumCreated(event AtprotoEvent) error {
	ec.logger.Info("HashPost subforum created",
		"repo", event.Repo,
		"uri", event.URI,
		"name", event.Record["name"],
	)

	// Extract subforum data from the record
	name, ok := event.Record["name"].(string)
	if !ok {
		return fmt.Errorf("invalid name field in subforum record")
	}

	slug, ok := event.Record["slug"].(string)
	if !ok {
		return fmt.Errorf("invalid slug field in subforum record")
	}

	description, ok := event.Record["description"].(string)
	if !ok {
		return fmt.Errorf("invalid description field in subforum record")
	}

	// Resolve creator handle from DID
	createdByHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve creator handle", "error", err, "did", event.Repo)
		createdByHandle = "unknown"
	}

	// Create AppView subforum data
	subforum := &AppViewSubforum{
		ID:              uuid.New(),
		Name:            name,
		Slug:            slug,
		Description:     description,
		CreatedByDID:    event.Repo,
		CreatedByHandle: createdByHandle,
		CreatedAt:       event.Timestamp,
		UpdatedAt:       event.Timestamp,
	}

	// Store in AppView database
	if err := ec.db.CreateSubforum(subforum); err != nil {
		return fmt.Errorf("failed to create subforum in AppView: %w", err)
	}

	ec.logger.Info("Subforum stored in AppView database",
		"name", name,
		"slug", slug,
		"description", description,
	)

	return nil
}

// handlePostUpdated processes post update events
func (ec *EventConsumer) handlePostUpdated(event AtprotoEvent) error {
	ec.logger.Info("HashPost feed post updated",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Extract post data from the record
	text, ok := event.Record["text"].(string)
	if !ok {
		return fmt.Errorf("invalid text field in post record")
	}

	// Resolve author handle from DID
	authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
		authorHandle = "unknown"
	}

	// Extract subforum slug from record context or use default
	subforumSlug := "general" // Default subforum
	if subforum, ok := event.Record["subforum"].(string); ok && subforum != "" {
		subforumSlug = subforum
	}

	// Create updated post data
	post := &AppViewPost{
		AtprotoURI:   event.URI,
		AuthorDID:    event.Repo,
		AuthorHandle: authorHandle,
		SubforumSlug: subforumSlug,
		Title:        text, // For now, use text as title
		Content:      text,
		UpdatedAt:    event.Timestamp,
	}

	// Update post in AppView database
	if err := ec.db.UpdatePost(event.URI, post); err != nil {
		return fmt.Errorf("failed to update post in AppView: %w", err)
	}

	ec.logger.Info("Post updated in AppView database", "uri", event.URI)
	return nil
}

// handleSubforumUpdated processes subforum update events
func (ec *EventConsumer) handleSubforumUpdated(event AtprotoEvent) error {
	ec.logger.Info("HashPost subforum updated",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Extract subforum data from the record
	name, ok := event.Record["name"].(string)
	if !ok {
		return fmt.Errorf("invalid name field in subforum record")
	}

	slug, ok := event.Record["slug"].(string)
	if !ok {
		return fmt.Errorf("invalid slug field in subforum record")
	}

	description, ok := event.Record["description"].(string)
	if !ok {
		return fmt.Errorf("invalid description field in subforum record")
	}

	// Resolve creator handle from DID
	createdByHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve creator handle", "error", err, "did", event.Repo)
		createdByHandle = "unknown"
	}

	// Create updated subforum data
	subforum := &AppViewSubforum{
		Name:            name,
		Slug:            slug,
		Description:     description,
		CreatedByDID:    event.Repo,
		CreatedByHandle: createdByHandle,
		UpdatedAt:       event.Timestamp,
	}

	// Update subforum in AppView database
	if err := ec.db.CreateSubforum(subforum); err != nil {
		return fmt.Errorf("failed to update subforum in AppView: %w", err)
	}

	ec.logger.Info("Subforum updated in AppView database", "slug", slug)
	return nil
}

// handlePostDeleted processes post deletion events
func (ec *EventConsumer) handlePostDeleted(event AtprotoEvent) error {
	ec.logger.Info("HashPost feed post deleted",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Delete post from AppView database
	if err := ec.db.DeletePost(event.URI); err != nil {
		return fmt.Errorf("failed to delete post from AppView: %w", err)
	}

	ec.logger.Info("Post deleted from AppView database", "uri", event.URI)
	return nil
}

// handleSubforumDeleted processes subforum deletion events
func (ec *EventConsumer) handleSubforumDeleted(event AtprotoEvent) error {
	ec.logger.Info("HashPost subforum deleted",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Extract slug from URI for deletion
	// This is a simplified approach - in production, we'd need proper URI parsing
	// For now, we'll need to implement subforum deletion by slug
	ec.logger.Info("Subforum deletion not yet implemented", "uri", event.URI)
	return nil
}

// handleCommentCreated processes HashPost comment creation events
func (ec *EventConsumer) handleCommentCreated(event AtprotoEvent) error {
	ec.logger.Info("HashPost comment created",
		"repo", event.Repo,
		"uri", event.URI,
		"text", event.Record["text"],
	)

	// Extract comment data from the record
	text, ok := event.Record["text"].(string)
	if !ok {
		return fmt.Errorf("invalid text field in comment record")
	}

	postURI, ok := event.Record["post"].(string)
	if !ok {
		return fmt.Errorf("invalid post field in comment record")
	}

	createdAtStr, ok := event.Record["createdAt"].(string)
	if !ok {
		return fmt.Errorf("invalid createdAt field in comment record")
	}

	// Parse timestamp
	var createdAt time.Time
	var err error
	createdAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			createdAt = time.Now()
			ec.logger.Warn("Failed to parse createdAt, using current time",
				"createdAt", createdAtStr, "error", err)
		}
	}

	// Resolve author handle from DID
	authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
	if err != nil {
		ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
		authorHandle = "unknown"
	}

	// Get parent comment URI if this is a reply
	var parentURI *string
	if parent, ok := event.Record["parent"].(string); ok && parent != "" {
		parentURI = &parent
	}

	// Find the post ID by URI
	post, err := ec.db.GetPostByURI(postURI)
	if err != nil {
		ec.logger.Error("Failed to find post for comment", "error", err, "post_uri", postURI)
		return fmt.Errorf("failed to find post: %w", err)
	}

	// Create AppView comment data
	comment := &AppViewComment{
		ID:           uuid.New(),
		AtprotoURI:   event.URI,
		AuthorDID:    event.Repo,
		AuthorHandle: authorHandle,
		PostID:       post.ID,
		Content:      text,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}

	// Set parent ID if this is a reply
	if parentURI != nil {
		parentComment, err := ec.db.GetCommentByURI(*parentURI)
		if err == nil {
			comment.ParentID = &parentComment.ID
		}
	}

	// Store in AppView database
	if err := ec.db.CreateComment(comment); err != nil {
		return fmt.Errorf("failed to create comment in AppView: %w", err)
	}

	// Update post comment count
	if err := ec.db.queries.UpdatePostCommentCount(context.Background(), pgtype.UUID{Bytes: post.ID, Valid: true}); err != nil {
		ec.logger.Warn("Failed to update post comment count", "error", err, "post_id", post.ID)
	}

	// Update user's last seen time
	if err := ec.db.queries.UpdateUserLastSeen(context.Background(), event.Repo); err != nil {
		ec.logger.Warn("Failed to update user last seen time", "error", err, "did", event.Repo)
	}

	ec.logger.Info("Comment stored in AppView database",
		"uri", event.URI,
		"text", text,
		"created_at", createdAt,
	)

	return nil
}

// handleCommentUpdated processes comment update events
func (ec *EventConsumer) handleCommentUpdated(event AtprotoEvent) error {
	ec.logger.Info("HashPost comment updated",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Extract comment data from the record
	text, ok := event.Record["text"].(string)
	if !ok {
		return fmt.Errorf("invalid text field in comment record")
	}

	// Create updated comment data
	comment := &AppViewComment{
		AtprotoURI: event.URI,
		Content:    text,
		UpdatedAt:  event.Timestamp,
	}

	// Update comment in AppView database
	if err := ec.db.UpdateComment(event.URI, comment); err != nil {
		return fmt.Errorf("failed to update comment in AppView: %w", err)
	}

	ec.logger.Info("Comment updated in AppView database", "uri", event.URI)
	return nil
}

// handleCommentDeleted processes comment deletion events
func (ec *EventConsumer) handleCommentDeleted(event AtprotoEvent) error {
	ec.logger.Info("HashPost comment deleted",
		"repo", event.Repo,
		"uri", event.URI,
	)

	// Get comment before deleting to update post count
	comment, err := ec.db.GetCommentByURI(event.URI)
	if err != nil {
		ec.logger.Warn("Failed to get comment before deletion", "error", err, "uri", event.URI)
	} else {
		// Update post comment count after deletion
		postUUID := comment.PostID
		if err := ec.db.queries.UpdatePostCommentCount(context.Background(), postUUID); err != nil {
			ec.logger.Warn("Failed to update post comment count after deletion", "error", err, "post_id", comment.PostID)
			// Don't fail the event processing, just log the error
		}
	}

	// Delete comment from AppView database
	if err := ec.db.DeleteCommentByURI(event.URI); err != nil {
		return fmt.Errorf("failed to delete comment from AppView: %w", err)
	}

	ec.logger.Info("Comment deleted from AppView database", "uri", event.URI)
	return nil
}

// isLocalUser determines if a user is local to HashPost PDS
func (ec *EventConsumer) isLocalUser(did string) bool {
	// For now, we'll use a simple heuristic based on DID format
	// In production, this would check against our PDS server DID
	// or query our local database to see if the user exists there

	// Simple check: if DID contains our server identifier, it's local
	// This is a placeholder - in production, we'd have proper PDS identification
	return !ec.isExternalDID(did)
}

// isExternalDID checks if a DID belongs to an external PDS
func (ec *EventConsumer) isExternalDID(did string) bool {
	// For development, we'll consider DIDs that don't contain our local identifiers as external
	// In production, this would be more sophisticated
	return !ec.containsLocalIdentifier(did)
}

// containsLocalIdentifier checks if a DID contains local identifiers
func (ec *EventConsumer) containsLocalIdentifier(did string) bool {
	// Check for local development identifiers
	localIdentifiers := []string{
		"hashpost",
		"localhost",
		"127.0.0.1",
	}

	for _, identifier := range localIdentifiers {
		if strings.Contains(did, identifier) {
			return true
		}
	}
	return false
}

// resolvePDSSource resolves the PDS endpoint for a given DID
func (ec *EventConsumer) resolvePDSSource(did string) string {
	// For now, return a placeholder PDS endpoint
	// In production, this would resolve the actual PDS endpoint from the DID document
	return "https://external-pds.example.com"
}

// handleVoteCreated processes vote creation events
func (ec *EventConsumer) handleVoteCreated(event AtprotoEvent) error {
	ec.logger.Info("HashPost vote created",
		"repo", event.Repo,
		"uri", event.URI,
		"subject", event.Record["subject"],
		"direction", event.Record["direction"],
	)

	// Extract vote data from the record
	subject, ok := event.Record["subject"].(string)
	if !ok {
		return fmt.Errorf("invalid subject field in vote record")
	}

	direction, ok := event.Record["direction"].(string)
	if !ok {
		return fmt.Errorf("invalid direction field in vote record")
	}

	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction value: %s", direction)
	}

	// Determine if this is a vote on a post or comment
	var postID uuid.UUID
	var commentID *uuid.UUID

	// Try to find the subject as a post first
	post, err := ec.db.GetPostByURI(subject)
	if err == nil {
		postID = post.ID
	} else {
		// Try to find as a comment
		comment, err := ec.db.GetCommentByURI(subject)
		if err != nil {
			ec.logger.Warn("Could not find post or comment for vote subject", "subject", subject)
			return nil // Don't fail the event processing
		}
		commentID = &comment.ID
		postID = comment.PostID.Bytes
	}

	// Create vote record in AppView database
	vote := &AppViewVote{
		ID:        uuid.New(),
		UserDID:   event.Repo,
		Subject:   subject,
		Direction: direction,
		CreatedAt: event.Timestamp,
	}

	if err := ec.db.CreateVote(vote); err != nil {
		return fmt.Errorf("failed to create vote in AppView: %w", err)
	}

	// Update vote counts
	if commentID != nil {
		// Update comment vote counts
		if err := ec.db.queries.UpdateCommentVoteCounts(context.Background(), *commentID); err != nil {
			ec.logger.Warn("Failed to update comment vote counts", "error", err, "comment_id", *commentID)
		}
	} else {
		// Update post vote counts
		if err := ec.db.queries.UpdatePostVoteCounts(context.Background(), postID); err != nil {
			ec.logger.Warn("Failed to update post vote counts", "error", err, "post_id", postID)
		}
	}

	ec.logger.Info("Vote stored in AppView database",
		"uri", event.URI,
		"subject", subject,
		"direction", direction,
	)

	return nil
}

// handleSubscriptionCreated processes subscription creation events
func (ec *EventConsumer) handleSubscriptionCreated(event AtprotoEvent) error {
	ec.logger.Info("HashPost subscription created",
		"repo", event.Repo,
		"uri", event.URI,
		"subject", event.Record["subject"],
	)

	// Extract subscription data from the record
	subject, ok := event.Record["subject"].(string)
	if !ok {
		return fmt.Errorf("invalid subject field in subscription record")
	}

	// Find the subforum by URI
	subforum, err := ec.db.GetSubforumByURI(subject)
	if err != nil {
		ec.logger.Warn("Could not find subforum for subscription subject", "subject", subject)
		return nil // Don't fail the event processing
	}

	// Create subscription record in AppView database
	subscription := &AppViewSubscription{
		ID:         uuid.New(),
		UserDID:    event.Repo,
		SubforumID: subforum.ID,
		CreatedAt:  event.Timestamp,
	}

	if err := ec.db.CreateSubscription(subscription); err != nil {
		return fmt.Errorf("failed to create subscription in AppView: %w", err)
	}

	// Update subforum subscriber count
	if err := ec.db.queries.UpdateSubforumSubscriberCount(context.Background(), subforum.Slug); err != nil {
		ec.logger.Warn("Failed to update subforum subscriber count", "error", err, "subforum_slug", subforum.Slug)
	}

	ec.logger.Info("Subscription stored in AppView database",
		"uri", event.URI,
		"subject", subject,
		"subforum_id", subforum.ID,
	)

	return nil
}

// Close closes the NATS connection
func (ec *EventConsumer) Close() error {
	if ec.nc != nil {
		return ec.nc.Drain()
	}
	return nil
}
