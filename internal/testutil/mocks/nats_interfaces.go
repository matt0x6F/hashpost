package mocks

import (
	"github.com/nats-io/nats.go"
)

// NatsConn interface for mocking nats.Conn
type NatsConn interface {
	JetStream(opts ...nats.JSOpt) (JetStreamContext, error)
	Close()
	Drain() error
}

// JetStreamContext interface for mocking nats.JetStreamContext
type JetStreamContext interface {
	StreamInfo(stream string, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
	DeleteStream(name string, opts ...nats.JSOpt) error
	Publish(subj string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
}
