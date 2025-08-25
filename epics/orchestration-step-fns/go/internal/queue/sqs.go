package queue

import "context"

// FrontierQueue abstracts interactions with the frontier and DLQ.
// Implementations should be mockable for tests.
type FrontierQueue interface {
	Enqueue(ctx context.Context, payload any) error
	Dequeue(ctx context.Context, max int) ([]any, error)
	DeadLetter(ctx context.Context, payload any, reason string) error
}
