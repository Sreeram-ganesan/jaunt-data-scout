package metrics

import "context"

type Emitter interface {
	IncCounter(ctx context.Context, name string, tags map[string]string, delta int)
	ObserveDuration(ctx context.Context, name string, tags map[string]string, millis float64)
	Event(ctx context.Context, name string, tags map[string]string)
}
