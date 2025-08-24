package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBudgetGuard_APICallLimit(t *testing.T) {
	bg := BudgetGuard{
		MaxAPICalls: 10,
		StartTime:   time.Now().Add(-1 * time.Minute),
	}
	require.True(t, bg.ShouldStop(context.Background(), Stats{APICalls: 10}))
	require.True(t, bg.ShouldStop(context.Background(), Stats{APICalls: 11}))
	require.False(t, bg.ShouldStop(context.Background(), Stats{APICalls: 9}))
}

func TestBudgetGuard_WallClock(t *testing.T) {
	bg := BudgetGuard{
		MaxWallClock: 100 * time.Millisecond,
		StartTime:    time.Now().Add(-200 * time.Millisecond),
	}
	require.True(t, bg.ShouldStop(context.Background(), Stats{}))
}

func TestBudgetGuard_MinNewUniqueRate(t *testing.T) {
	bg := BudgetGuard{
		MinNewUniqueRate: 0.10,
		StartTime:        time.Now(),
	}
	require.True(t, bg.ShouldStop(context.Background(), Stats{NewUniqueItems: 9, TotalItemsSeen: 100}))
	require.False(t, bg.ShouldStop(context.Background(), Stats{NewUniqueItems: 10, TotalItemsSeen: 100}))
}

func TestBudgetGuard_ZeroOrUnsetDoesNotStop(t *testing.T) {
	bg := BudgetGuard{}
	require.False(t, bg.ShouldStop(context.Background(), Stats{}))
}