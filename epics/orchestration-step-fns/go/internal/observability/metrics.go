package observability

import (
	"context"
	"sync/atomic"
)

// Counters provides atomic counters for basic metrics tracking
type Counters struct {
	Calls               atomic.Int64
	Errors              atomic.Int64
	Backoffs            atomic.Int64
	NewUnique           atomic.Int64
	LastNCalls          atomic.Int64
	ExtractorTokenCount atomic.Int64
	HTTPBytesIn         atomic.Int64
}

func (c *Counters) RecordCall()             { c.Calls.Add(1); c.LastNCalls.Add(1) }
func (c *Counters) RecordError()            { c.Errors.Add(1) }
func (c *Counters) RecordBackoff()          { c.Backoffs.Add(1) }
func (c *Counters) RecordNewUnique(n int64) { c.NewUnique.Add(n) }
func (c *Counters) RecordTokens(n int64)    { c.ExtractorTokenCount.Add(n) }
func (c *Counters) RecordHTTPBytes(n int64) { c.HTTPBytesIn.Add(n) }

func (c *Counters) NewUniqueRate() float64 {
	last := c.LastNCalls.Load()
	if last == 0 {
		return 1.0
	}
	return float64(c.NewUnique.Load()) / float64(last)
}

// Typed metric helpers using EMF for CloudWatch
func CountCall(ctx context.Context, service, state, connector, city string) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "Calls",
		Unit:        "Count",
		Value:       1,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func CountError(ctx context.Context, service, state, connector, city string) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "Errors",
		Unit:        "Count",
		Value:       1,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func RecordDurationMS(ctx context.Context, service, state, connector, city string, durationMS float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "Duration",
		Unit:        "Milliseconds",
		Value:       durationMS,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func RecordHTTPBytesIn(ctx context.Context, service, state, connector, city string, bytes float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "HTTPBytesIn",
		Unit:        "Bytes",
		Value:       bytes,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func RecordTokensInOut(ctx context.Context, service, state, connector, city string, tokensIn, tokensOut float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "TokensIn",
		Unit:        "Count",
		Value:       tokensIn,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
	emitEMF(ctx, &EMFMetric{
		MetricName:  "TokensOut",
		Unit:        "Count",
		Value:       tokensOut,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func RecordTokenCostEstimate(ctx context.Context, service, state, connector, city string, cost float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "TokenCostEstimate",
		Unit:        "None",
		Value:       cost,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func RecordNewUniqueRate(ctx context.Context, service, state, connector, city string, rate float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "NewUniqueRate",
		Unit:        "Percent",
		Value:       rate * 100,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

func BudgetCapGauge(ctx context.Context, service, state, connector, city string, utilization float64) {
	emitEMF(ctx, &EMFMetric{
		MetricName:  "BudgetCapUtilization",
		Unit:        "Percent",
		Value:       utilization * 100,
		Service:     service,
		State:       state,
		Connector:   connector,
		City:        city,
		RunID:       extractRunID(ctx),
		Split:       extractSplit(ctx),
		CorrelationID: FromContext(ctx),
	})
}

// Helper functions to extract metadata from context
func extractRunID(ctx context.Context) string {
	if runID, ok := ctx.Value("run_id").(string); ok {
		return runID
	}
	return "unknown"
}

func extractSplit(ctx context.Context) string {
	if split, ok := ctx.Value("split").(string); ok {
		return split
	}
	return "unknown"
}
