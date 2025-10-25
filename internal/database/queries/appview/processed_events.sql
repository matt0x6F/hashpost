-- name: CreateProcessedEvent :one
INSERT INTO processed_events (event_id, subject, sequence)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProcessedEvent :one
SELECT * FROM processed_events
WHERE event_id = $1;

-- name: IsEventProcessed :one
SELECT EXISTS(
    SELECT 1 FROM processed_events
    WHERE event_id = $1
) as exists;

-- name: CleanupOldProcessedEvents :exec
DELETE FROM processed_events
WHERE processed_at < NOW() - INTERVAL '7 days';

-- name: GetProcessedEventsCount :one
SELECT COUNT(*) as count FROM processed_events;
