package budget

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Connector string

const (
	GoogleText    Connector = "google.text"
	GoogleNearby  Connector = "google.nearby"
	GoogleDetails Connector = "google.details"
	Overpass      Connector = "overpass"
	OTM           Connector = "otm"
	Wiki          Connector = "wiki"
	TavilyAPI     Connector = "tavily.api"
	WebFetch      Connector = "web.fetch"
	LLMTokens     Connector = "llm.tokens"
	Nominatim     Connector = "nominatim"
)

type Split string

const (
	Primaries   Split = "primaries"
	Secondaries Split = "secondaries"
)

type Bucket struct {
	capacity int64
	tokens   int64
	refill   int64
	period   time.Duration
	last     time.Time
	mu       sync.Mutex
}

func newBucket(capacity, refill int64, period time.Duration) *Bucket {
	return &Bucket{
		capacity: capacity,
		tokens:   capacity,
		refill:   refill,
		period:   period,
		last:     time.Now(),
	}
}

func (b *Bucket) take(n int64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last)
	if elapsed >= b.period && b.refill > 0 {
		steps := int64(elapsed / b.period)
		b.tokens = min64(b.capacity, b.tokens+steps*b.refill)
		b.last = now
	}
	if b.tokens >= n {
		b.tokens -= n
		return true
	}
	return false
}

func (b *Bucket) put(n int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens = min64(b.capacity, b.tokens+n)
}

type Guard struct {
	mu         sync.RWMutex
	buckets    map[Connector]*Bucket
	splitQuota map[Split]int64 // tokens reserved per split in current window
	splitUsed  map[Split]int64
	splitRatio float64 // 0.7 => 70% primaries
}

type Config struct {
	Budgets map[Connector]struct {
		Capacity int64         `yaml:"capacity"`
		Refill   int64         `yaml:"refill"`
		Period   time.Duration `yaml:"period"`
	} `yaml:"budgets"`
	SplitRatio float64 `yaml:"split_ratio"`
}

func NewGuard(cfg Config) *Guard {
	b := make(map[Connector]*Bucket, len(cfg.Budgets))
	for c, v := range cfg.Budgets {
		period := v.Period
		if period == 0 {
			period = time.Minute
		}
		b[c] = newBucket(v.Capacity, v.Refill, period)
	}
	return &Guard{
		buckets:    b,
		splitQuota: map[Split]int64{Primaries: 0, Secondaries: 0},
		splitUsed:  map[Split]int64{Primaries: 0, Secondaries: 0},
		splitRatio: cfg.SplitRatio,
	}
}

// Rebalance should be called at run start and periodically to set split quotas based on current token stocks.
func (g *Guard) Rebalance() {
	g.mu.Lock()
	defer g.mu.Unlock()
	var total int64
	for _, b := range g.buckets {
		total += b.tokens
	}
	g.splitQuota[Primaries] = int64(float64(total) * g.splitRatio)
	g.splitQuota[Secondaries] = total - g.splitQuota[Primaries]
	g.splitUsed[Primaries] = 0
	g.splitUsed[Secondaries] = 0
}

type AcquireOpts struct {
	Connector Connector
	Split     Split
	Tokens    int64
	Deadline  time.Duration
}

var ErrBudgetExceeded = errors.New("budget exceeded")

// Acquire tries to take tokens from the connector bucket while honoring the 70/30 split.
func (g *Guard) Acquire(ctx context.Context, opts AcquireOpts) error {
	if opts.Tokens <= 0 {
		opts.Tokens = 1
	}
	if opts.Deadline <= 0 {
		opts.Deadline = 2 * time.Second
	}
	deadline := time.After(opts.Deadline)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return ErrBudgetExceeded
		default:
			if g.tryAcquire(opts) {
				return nil
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (g *Guard) tryAcquire(opts AcquireOpts) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	b, ok := g.buckets[opts.Connector]
	if !ok {
		return false
	}

	// enforce split
	if g.splitQuota[opts.Split] > 0 && g.splitUsed[opts.Split]+opts.Tokens > g.splitQuota[opts.Split] {
		return false
	}

	if b.take(opts.Tokens) {
		g.splitUsed[opts.Split] += opts.Tokens
		return true
	}
	return false
}

func (g *Guard) Release(connector Connector, tokens int64, split Split) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.buckets[connector]; ok {
		b.put(tokens)
	}
	if tokens < 0 {
		tokens = 0
	}
	if g.splitUsed[split] >= tokens {
		g.splitUsed[split] -= tokens
	} else {
		g.splitUsed[split] = 0
	}
}

type ProgressWindow struct {
	LastNNewUnique int64
	LastNCalls     int64
}

func (pw ProgressWindow) NewUniqueRate() float64 {
	if pw.LastNCalls == 0 {
		return 1.0
	}
	return float64(pw.LastNNewUnique) / float64(pw.LastNCalls)
}

// EarlyStop returns true if the new_unique_rate over the last window falls below threshold.
func EarlyStop(pw ProgressWindow, threshold float64) bool {
	return pw.NewUniqueRate() < threshold
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
