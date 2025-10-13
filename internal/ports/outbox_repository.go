package ports

import "context"

// OutboxRepository exposes operations for background notification workers.
type OutboxRepository interface {
	// UpdateOutboxStatus updates the status field of an outbox event.
	// Implementations should also maintain processed timestamps when the new status is "processed".
	UpdateOutboxStatus(ctx context.Context, eventID int64, status string) error
}
