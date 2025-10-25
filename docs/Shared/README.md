# Shared Documentation

## Overview

This directory contains documentation for shared components and patterns used across both PDS and AppView services in the HashPost platform.

## Architecture

The HashPost platform follows a dual-server architecture pattern:
- **PDS (Personal Data Server)**: Handles atproto protocol compliance and stores canonical data including forum-specific tables
- **AppView**: Stateful web application with business logic and persistence that provides forum functionality
- **Database Separation**: PDS uses `hashpost_pds_dev`, AppView uses `hashpost_appview_dev`
- **Event-Driven Communication**: PDS publishes events → NATS JetStream → AppView consumes (limited current usage)

## Documentation Structure

### Architecture
- [Dual Server Pattern](Architecture/dual-server-pattern.md) - PDS vs AppView responsibilities and separation rationale
- [Data Flow](Architecture/data-flow.md) - End-to-end data flow and event-driven architecture

### Database
- [Separation Strategy](Database/separation-strategy.md) - Why separate databases, data ownership, consistency
- [SQLC Usage](Database/sqlc-usage.md) - SQLC patterns, query organization, code generation workflow

### Events
- [Event Streaming](Events/event-streaming.md) - Event types, payload structure, versioning
- [NATS JetStream](Events/nats-jetstream.md) - NATS setup, stream configuration, consumer patterns

## Key Concepts

### Database Separation
- **PDS Database**: Stores canonical atproto records plus forum-specific tables (posts, comments, votes, subforums)
- **AppView Database**: Stores denormalized/aggregated data optimized for forum queries
- **Event Synchronization**: Limited event processing for data sync (primarily direct database access)

### Event-Driven Architecture
- **Event Publishing**: PDS publishes events to NATS JetStream (implemented)
- **Event Consumption**: AppView consumes events and updates denormalized data (limited current usage)
- **Error Handling**: Retry logic, exponential backoff, dead letter queue (implemented)

### SQLC Integration
- **Type Safety**: Generated Go code from SQL queries
- **Query Organization**: Separate query files by domain
- **Code Generation**: Automated code generation workflow

## References

- [HashPost Architecture Overview](../designs/archive/hashpost-architecture.md)
- [Event Processing Architecture](../designs/archive/event-processing-architecture.md)
- [Database Migrations](../database/migrations/)
- [SQLC Configuration](../sqlc-pds.yaml, ../sqlc-appview.yaml)
