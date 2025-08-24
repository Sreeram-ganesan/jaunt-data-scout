package observability

import "context"

type Span interface {
    End()
    AddEvent(name string, attrs ...any)
}

type noopSpan struct{}

func (noopSpan) End()                         {}
func (noopSpan) AddEvent(name string, attrs ...any) {}

type Tracer struct{}

func NewTracer() *Tracer { return &Tracer{} }

func (t *Tracer) Start(ctx context.Context, name string) (context.Context, Span) {
    // Wire to OTEL in future
    return ctx, noopSpan{}
}