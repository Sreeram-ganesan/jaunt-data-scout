# Epic Work Tracker — Orchestration (Step Functions workflow for city discovery)

Epic goals
- Orchestrate MVP pipeline: DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep → DedupeCanonicalize → Persist → Finalize.
- Integrate SQS frontier + DLQ, S3 raw cache, and metrics/traces.
- Respect budgets; emit CloudWatch metrics and OTEL traces.

Acceptance criteria (for Edinburgh run with mocked connectors)
- [x] City job completes end-to-end with mocked connectors.
- [x] State transitions logged; failures routed to DLQ.
- [x] Raw responses cached to S3 by (source, request_hash).
- [x] Budgets respected (max_api_calls, max_wall_clock_hours, min_new_unique_rate).
- [x] CloudWatch metrics and OTEL traces emitted for each state.

Tasks and subtasks

1) Define Step Functions state machine with all MVP states and retry/backoff.
- [x] Draft ASL skeleton with all states
- [x] Add uniform retry/backoff on transient failures
- [x] Configure logging to CloudWatch Logs
- [x] Integrate DLQ via failure handling pathway
- [x] Parameterize with project/environment tags and names
- [x] Validate definition with aws stepfunctions validate-state-machine-definition (or terraform validate)

2) Implement budget guard + early-stop choice states.
- [x] Specify budget inputs (max_api_calls, max_wall_clock_hours, min_new_unique_rate)
- [x] Implement BudgetGuard logic (Go) with unit tests
- [x] Expose budget evaluations to Step Functions (e.g., task output or input path)
- [x] Add Choice state(s) to short-circuit when budget is exceeded
- [x] Record early-stop reason to logs/metrics

3) Wire SQS queues (frontier, DLQ) and message schema (lat/lng, radius, category).
- [x] Create frontier queue (SQS)
- [x] Create DLQ and attach redrive policy
- [x] Define message schema: {lat, lng, radius, category, correlation_id}
- [x] Implement enqueue/dequeue client (Go) with mocks and tests
- [x] Decide visibility timeout and batching strategy
- [x] Emit metrics on queue depth and failures

4) S3 layout for raw cache and manifests (per city run).
- [x] Create bucket with versioning and SSE
- [x] Define key layout: s3://<bucket>/<city>/<source>/<request_hash>.json
- [x] Write manifest pattern per run (e.g., manifests/<city>/<run_id>.json)
- [x] Implement cache get/put client (Go) with tests (mock S3)
- [x] Add lifecycle rules for cold storage/expiration

5) Emit metrics/alerts per state; add tracing context propagation.
- [x] Define metric names and dimensions (city, state, source)
- [x] Add CloudWatch metrics emission from tasks
- [x] Configure alarms for error rates / DLQ messages
- [x] Add OTEL tracing stubs (Go) with propagation across steps
- [x] Document correlation_id strategy and logging context
- [x] Ensure state transitions and durations are recorded

Integrations readiness checklist
- [x] IAM roles/policies scoped to least privilege
- [x] Environment-specific configuration (dev/stage/prod)
- [x] Runbook for common failures (DLQ, throttling)
- [x] Cost guardrails (budgets, alerts)
- [x] Observability dashboards (optional for MVP)

Additional Epic Completion Tasks (v2)
- [x] Feature flags to flip individual states mock↔real (per-state env/vars); document toggling procedure.
- [x] DLQ re-drive runbook and helper CLI; example reprocessing flow.
- [x] Execution input presets (golden inputs) per city for smoke/e2e tests.
- [x] Complete integration testing documentation and procedures.
- [x] End-to-end integration test script with validation.
- [x] Comprehensive deployment and testing instructions.

Links
- Main README: ./README.md
- Terraform readme: ./terraform/README.md
- Go TDD readme: ./go/README.md
- Integration Testing Guide: ./INTEGRATION_TESTING.md
- DLQ Runbook: ./DLQ_RUNBOOK.md
- Feature Flags Guide: ./terraform/FEATURE_FLAGS.md
- Examples and Golden Datasets: ./examples/README.md