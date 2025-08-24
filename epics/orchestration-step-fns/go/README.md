# Go — TDD scaffolding for orchestration helpers

Layout
- cmd/cityjob: sample CLI entrypoint for local runs
- internal/workflow: state machine helpers and budget guard
- internal/queue: frontier/DLQ abstractions (SQS-backed)
- internal/cache: raw cache client (S3-backed)
- internal/metrics: metrics façade (CloudWatch)
- internal/tracing: tracing façade (OTEL)

Running tests
- make test
- Start by unskipping tests under internal/* when implementing features.
- BudgetGuard is implemented + tested as an example of TDD flow.

Dependencies
- Go 1.22+
- github.com/stretchr/testify for assertions