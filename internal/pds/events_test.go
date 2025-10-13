package pds

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/matt0x6f/hashpost/internal/testutil/mocks"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventStreamer_PublishRecordEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJetStream := mocks.NewMockJetStreamContext(ctrl)
	mockNatsConn := mocks.NewMockNatsConn(ctrl)

	// Mock NATS connection and JetStream context
	mockNatsConn.EXPECT().JetStream().Return(mockJetStream, nil).AnyTimes()
	mockNatsConn.EXPECT().Close().AnyTimes()
	mockNatsConn.EXPECT().Drain().AnyTimes()

	logger := testutil.CreateMockLogger()
	eventStreamer := &EventStreamer{
		nc:      mockNatsConn,
		js:      mockJetStream,
		logger:  logger,
		stream:  "HASHPOST_EVENTS",
		subject: "hashpost.events.record.created",
	}

	// Expect StreamInfo to be called and return an existing stream
	mockJetStream.EXPECT().StreamInfo("HASHPOST_EVENTS", gomock.Any()).Return(&nats.StreamInfo{
		Config: nats.StreamConfig{
			Name:     "HASHPOST_EVENTS",
			Subjects: []string{"hashpost.events.record.created"},
		},
	}, nil).AnyTimes()

	// Expect Publish to be called
	mockJetStream.EXPECT().Publish("hashpost.events.record.created", gomock.Any(), gomock.Any()).Return(&nats.PubAck{
		Stream:   "HASHPOST_EVENTS",
		Sequence: 1,
	}, nil).Do(func(subject string, data []byte, opts ...nats.PubOpt) {
		assert.Equal(t, "hashpost.events.record.created", subject)
		// Basic check that data is not empty
		assert.NotEmpty(t, data)
		// Further validation of data can be done by unmarshaling and comparing
		var publishedEvent AtprotoEvent
		err := json.Unmarshal(data, &publishedEvent)
		require.NoError(t, err)
		assert.Equal(t, EventTypeRecordCreated, publishedEvent.Type)
		assert.Equal(t, "did:plc:test-user", publishedEvent.Repo)
		assert.Equal(t, "app.bsky.feed.post", publishedEvent.Collection)
		assert.Equal(t, "at://did:plc:test-user/app.bsky.feed.post/123", publishedEvent.URI)
		assert.Equal(t, "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", publishedEvent.CID)
		assert.Equal(t, "Hello, world!", publishedEvent.Record["text"])
		assert.Equal(t, "hashpost-pds", publishedEvent.Metadata["source"])
	}).Times(1)

	err := eventStreamer.PublishRecordEvent(
		context.Background(),
		EventTypeRecordCreated,
		"did:plc:test-user",
		"app.bsky.feed.post",
		map[string]interface{}{"text": "Hello, world!"},
		"at://did:plc:test-user/app.bsky.feed.post/123",
		"bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
	)
	require.NoError(t, err)
}

func TestEventStreamer_PublishIdentityEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJetStream := mocks.NewMockJetStreamContext(ctrl)
	mockNatsConn := mocks.NewMockNatsConn(ctrl)

	mockNatsConn.EXPECT().JetStream().Return(mockJetStream, nil).AnyTimes()
	mockNatsConn.EXPECT().Close().AnyTimes()
	mockNatsConn.EXPECT().Drain().AnyTimes()

	logger := testutil.CreateMockLogger()
	eventStreamer := &EventStreamer{
		nc:      mockNatsConn,
		js:      mockJetStream,
		logger:  logger,
		stream:  "HASHPOST_EVENTS",
		subject: "hashpost.events.record.created",
	}

	// Expect StreamInfo to be called and return an existing stream
	mockJetStream.EXPECT().StreamInfo("HASHPOST_EVENTS", gomock.Any()).Return(&nats.StreamInfo{
		Config: nats.StreamConfig{
			Name:     "HASHPOST_EVENTS",
			Subjects: []string{"hashpost.events.*"}, // Stream configured for wildcard
		},
	}, nil).AnyTimes()

	// Expect Publish to be called
	mockJetStream.EXPECT().Publish("hashpost.events.record.created", gomock.Any(), gomock.Any()).Return(&nats.PubAck{
		Stream:   "HASHPOST_EVENTS",
		Sequence: 1,
	}, nil).Do(func(subject string, data []byte, opts ...nats.PubOpt) {
		assert.Equal(t, "hashpost.events.record.created", subject)
		var publishedEvent AtprotoEvent
		err := json.Unmarshal(data, &publishedEvent)
		require.NoError(t, err)
		assert.Equal(t, EventTypeIdentityResolved, publishedEvent.Type)
		assert.Equal(t, "did:plc:test-user", publishedEvent.Repo)
		assert.Equal(t, "testuser.hashpost.local", publishedEvent.Metadata["handle"])
		assert.Equal(t, "did:plc:test-user", publishedEvent.Metadata["did"])
	}).Times(1)

	err := eventStreamer.PublishIdentityEvent(
		context.Background(),
		EventTypeIdentityResolved,
		"testuser.hashpost.local",
		"did:plc:test-user",
	)
	require.NoError(t, err)
}

func TestEventStreamer_PublishSessionEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJetStream := mocks.NewMockJetStreamContext(ctrl)
	mockNatsConn := mocks.NewMockNatsConn(ctrl)

	mockNatsConn.EXPECT().JetStream().Return(mockJetStream, nil).AnyTimes()
	mockNatsConn.EXPECT().Close().AnyTimes()
	mockNatsConn.EXPECT().Drain().AnyTimes()

	logger := testutil.CreateMockLogger()
	eventStreamer := &EventStreamer{
		nc:      mockNatsConn,
		js:      mockJetStream,
		logger:  logger,
		stream:  "HASHPOST_EVENTS",
		subject: "hashpost.events.record.created",
	}

	// Expect StreamInfo to be called and return an existing stream
	mockJetStream.EXPECT().StreamInfo("HASHPOST_EVENTS", gomock.Any()).Return(&nats.StreamInfo{
		Config: nats.StreamConfig{
			Name:     "HASHPOST_EVENTS",
			Subjects: []string{"hashpost.events.*"}, // Stream configured for wildcard
		},
	}, nil).AnyTimes()

	// Expect Publish to be called
	mockJetStream.EXPECT().Publish("hashpost.events.record.created", gomock.Any(), gomock.Any()).Return(&nats.PubAck{
		Stream:   "HASHPOST_EVENTS",
		Sequence: 1,
	}, nil).Do(func(subject string, data []byte, opts ...nats.PubOpt) {
		assert.Equal(t, "hashpost.events.record.created", subject)
		var publishedEvent AtprotoEvent
		err := json.Unmarshal(data, &publishedEvent)
		require.NoError(t, err)
		assert.Equal(t, EventTypeSessionCreated, publishedEvent.Type)
		assert.Equal(t, "did:plc:test-user", publishedEvent.Repo)
		assert.Equal(t, "session-123", publishedEvent.Metadata["session_id"])
	}).Times(1)

	err := eventStreamer.PublishSessionEvent(
		context.Background(),
		EventTypeSessionCreated,
		"did:plc:test-user",
		"session-123",
	)
	require.NoError(t, err)
}

func TestEventStreamer_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNatsConn := mocks.NewMockNatsConn(ctrl)
	mockJetStream := mocks.NewMockJetStreamContext(ctrl)

	mockNatsConn.EXPECT().Drain().Return(nil).Times(1)

	logger := testutil.CreateMockLogger()
	eventStreamer := &EventStreamer{
		nc:      mockNatsConn,
		js:      mockJetStream,
		logger:  logger,
		stream:  "HASHPOST_EVENTS",
		subject: "hashpost.events.record.created",
	}

	err := eventStreamer.Close()
	require.NoError(t, err)
}
