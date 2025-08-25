package workflow

import "context"

// Runner represents a local simulation harness for the Step Functions flow.
// Implementations should model transitions and call external dependencies via interfaces.
type Runner interface {
	RunCity(ctx context.Context, city string) error
}
