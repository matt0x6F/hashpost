package pds

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil/mocks"
	"github.com/nats-io/nats.go"
)

// EventType represents the type of atproto event
type EventType string

const (
	EventTypeRecordCreated    EventType = "record.created"
	EventTypeRecordUpdated    EventType = "record.updated"
	EventTypeRecordDeleted    EventType = "record.deleted"
	EventTypeIdentityResolved EventType = "identity.resolved"
	EventTypeSessionCreated   EventType = "session.created"
)

// AtprotoEvent represents an atproto event to be streamed
type AtprotoEvent struct {
	Type       EventType              `json:"type"`
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection,omitempty"`
	Record     map[string]interface{} `json:"record,omitempty"`
	URI        string                 `json:"uri,omitempty"`
	CID        string                 `json:"cid,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Use interfaces for testability
type NatsConn = mocks.NatsConn
type JetStreamContext = mocks.JetStreamContext

// natsConnWrapper wraps *nats.Conn to implement NatsConn interface
type natsConnWrapper struct {
	*nats.Conn
}

func (w *natsConnWrapper) JetStream(opts ...nats.JSOpt) (JetStreamContext, error) {
	js, err := w.Conn.JetStream(opts...)
	if err != nil {
		return nil, err
	}
	return &jetStreamWrapper{js}, nil
}

// jetStreamWrapper wraps nats.JetStreamContext to implement JetStreamContext interface
type jetStreamWrapper struct {
	nats.JetStreamContext
}

// EventStreamer handles publishing atproto events to NATS Jetstream
type EventStreamer struct {
	nc      NatsConn
	js      JetStreamContext
	logger  *slog.Logger
	stream  string
	subject string
}

// NewEventStreamer creates a new event streamer
func NewEventStreamer(natsURL string, logger *slog.Logger) (*EventStreamer, error) {
	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Wrap the connection to implement our interface
	wrappedConn := &natsConnWrapper{Conn: nc}

	// Create JetStream context
	js, err := wrappedConn.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	streamer := &EventStreamer{
		nc:      wrappedConn,
		js:      js,
		logger:  logger,
		stream:  "HASHPOST_EVENTS",
		subject: "hashpost.events.record.created", // Use exact subject for debugging
	}

	// Create the stream
	if err := streamer.createStream(); err != nil {
		nc.Close()
		logger.Error("Failed to create NATS stream", "error", err)
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	// For debugging, let's force recreate the stream to ensure it's properly configured
	if err := streamer.recreateStream(); err != nil {
		logger.Warn("Failed to recreate stream", "error", err)
		// Don't fail here, just log the warning
	}

	return streamer, nil
}

// createStream creates the NATS JetStream stream for atproto events
func (es *EventStreamer) createStream() error {
	// Check if stream already exists
	streamInfo, err := es.js.StreamInfo(es.stream)
	if err == nil {
		// Stream exists, log its info
		es.logger.Info("NATS stream already exists",
			"stream", es.stream,
			"subjects", streamInfo.Config.Subjects,
			"state", streamInfo.State,
		)
		return nil
	}

	es.logger.Info("Stream does not exist, creating new stream", "stream", es.stream, "error", err)

	// Create stream configuration
	streamConfig := &nats.StreamConfig{
		Name:      es.stream,
		Subjects:  []string{es.subject},
		Retention: nats.LimitsPolicy, // Keep messages until they expire
		MaxAge:    24 * time.Hour,    // Keep events for 24 hours
		Storage:   nats.FileStorage,
		Replicas:  1,
	}

	// Create the stream
	_, err = es.js.AddStream(streamConfig)
	if err != nil {
		es.logger.Error("Failed to create NATS stream", "error", err, "stream", es.stream)
		return fmt.Errorf("failed to add stream: %w", err)
	}

	es.logger.Info("Created NATS JetStream", "stream", es.stream, "subject", es.subject)
	return nil
}

// recreateStream deletes and recreates the stream to ensure proper configuration
func (es *EventStreamer) recreateStream() error {
	// Delete the existing stream
	err := es.js.DeleteStream(es.stream)
	if err != nil {
		es.logger.Warn("Failed to delete existing stream", "error", err)
	} else {
		es.logger.Info("Deleted existing stream", "stream", es.stream)
	}

	// Create stream configuration with more explicit settings
	es.logger.Info("Recreating stream with subject pattern", "subject", es.subject)

	streamConfig := &nats.StreamConfig{
		Name:      es.stream,
		Subjects:  []string{es.subject},
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour,
		Storage:   nats.FileStorage,
		Replicas:  1,
	}

	// Create the stream
	_, err = es.js.AddStream(streamConfig)
	if err != nil {
		es.logger.Error("Failed to recreate NATS stream", "error", err, "stream", es.stream)
		return fmt.Errorf("failed to recreate stream: %w", err)
	}

	es.logger.Info("Recreated NATS JetStream", "stream", es.stream, "subject", es.subject)
	return nil
}

// PublishRecordEvent publishes a record-related event
func (es *EventStreamer) PublishRecordEvent(ctx context.Context, eventType EventType, repo, collection string, record map[string]interface{}, uri, cid string) error {
	event := &AtprotoEvent{
		Type:       eventType,
		Repo:       repo,
		Collection: collection,
		Record:     record,
		URI:        uri,
		CID:        cid,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"source": "hashpost-pds",
		},
	}

	return es.publishEvent(ctx, event)
}

// PublishIdentityEvent publishes an identity-related event
func (es *EventStreamer) PublishIdentityEvent(ctx context.Context, eventType EventType, handle, did string) error {
	event := &AtprotoEvent{
		Type:      eventType,
		Repo:      did,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"source": "hashpost-pds",
			"handle": handle,
			"did":    did,
		},
	}

	return es.publishEvent(ctx, event)
}

// PublishSessionEvent publishes a session-related event
func (es *EventStreamer) PublishSessionEvent(ctx context.Context, eventType EventType, did, sessionID string) error {
	event := &AtprotoEvent{
		Type:      eventType,
		Repo:      did,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"source":     "hashpost-pds",
			"session_id": sessionID,
		},
	}

	return es.publishEvent(ctx, event)
}

// publishEvent publishes an event to NATS JetStream
func (es *EventStreamer) publishEvent(ctx context.Context, event *AtprotoEvent) error {
	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create subject based on event type
	// For debugging, let's try publishing to the exact stream subject pattern
	// The stream is configured for "hashpost.events.*" so let's try that exact pattern
	subject := "hashpost.events.record.created"

	// Try a different approach - let's try publishing to the exact wildcard pattern
	// that the stream is configured for
	if subject == "hashpost.events.record.created" {
		subject = "hashpost.events.record.created" // Keep the same for now
	}

	es.logger.Debug("Attempting to publish event",
		"subject", subject,
		"stream", es.stream,
		"data_length", len(data),
	)

	// Check if stream exists before publishing
	streamInfo, err := es.js.StreamInfo(es.stream)
	if err != nil {
		es.logger.Error("Stream does not exist", "error", err, "stream", es.stream)
		return fmt.Errorf("stream does not exist: %w", err)
	}

	es.logger.Debug("Stream exists",
		"stream", streamInfo.Config.Name,
		"subjects", streamInfo.Config.Subjects,
		"state", streamInfo.State,
	)

	// Try publishing synchronously first
	ack, err := es.js.Publish(subject, data)
	if err != nil {
		es.logger.Error("Failed to publish event to NATS",
			"error", err,
			"subject", subject,
			"stream", es.stream,
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	es.logger.Debug("Event published successfully",
		"stream", ack.Stream,
		"sequence", ack.Sequence,
		"subject", subject,
	)

	es.logger.Debug("Published atproto event",
		"type", event.Type,
		"repo", event.Repo,
		"subject", subject,
	)

	return nil
}

// Close closes the NATS connection
func (es *EventStreamer) Close() error {
	if es.nc != nil {
		return es.nc.Drain()
	}
	return nil
}
