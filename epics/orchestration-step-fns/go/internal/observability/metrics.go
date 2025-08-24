package observability

import "sync/atomic"

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
    if last == 0 { return 1.0 }
    return float64(c.NewUnique.Load()) / float64(last)
}