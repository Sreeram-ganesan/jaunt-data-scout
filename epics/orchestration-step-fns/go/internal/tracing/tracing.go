package tracing

import "context"

type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, func())
}