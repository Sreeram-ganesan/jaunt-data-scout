# Epic Work Tracker — Orchestration (Step Functions workflow for city discovery)

Epic goals
- Orchestrate MVP pipeline: DiscoverTargets → SeedPrimaries → ExpandNeighbors → TileSweep → DedupeCanonicalize → Persist → Finalize.
- Integrate SQS frontier + DLQ, S3 raw cache, and metrics/traces.
- Respect budgets; emit CloudWatch metrics and OTEL traces.

Acceptance criteria (for Edinburgh run with mocked connectors)
- [ ] City job completes end-to-end with mocked connectors.
- [ ] State transitions logged; failures routed to DLQ.
- [ ] Raw responses cached to S3 by (source, request_hash).
- [ ] Budgets respected (max_api_calls, max_wall_clock_hours, min_new_unique_rate).
- [ ] CloudWatch metrics and OTEL traces emitted for each state.

Tasks and subtasks

1) Define Step Functions state machine with all MVP states and retry/backoff.
- [ ] Draft ASL skeleton with all states
- [ ] Add uniform retry/backoff on transient failures
- [ ] Configure logging to CloudWatch Logs
- [ ] Integrate DLQ via failure handling pathway
- [ ] Parameterize with project/environment tags and names
- [ ] Validate definition with aws stepfunctions validate-state-machine-definition (or terraform validate)

2) Implement budget guard + early-stop choice states.
- [ ] Specify budget inputs (max_api_calls, max_wall_clock_hours, min_new_unique_rate)
- [ ] Implement BudgetGuard logic (Go) with unit tests
- [ ] Expose budget evaluations to Step Functions (e.g., task output or input path)
- [ ] Add Choice state(s) to short-circuit when budget is exceeded
- [ ] Record early-stop reason to logs/metrics

3) Wire SQS queues (frontier, DLQ) and message schema (lat/lng, radius, category).
- [ ] Create frontier queue (SQS)
- [ ] Create DLQ and attach redrive policy
- [ ] Define message schema: {lat, lng, radius, category, correlation_id}
- [ ] Implement enqueue/dequeue client (Go) with mocks and tests
- [ ] Decide visibility timeout and batching strategy
- [ ] Emit metrics on queue depth and failures

4) S3 layout for raw cache and manifests (per city run).
- [ ] Create bucket with versioning and SSE
- [ ] Define key layout: s3://<bucket>/<city>/<source>/<request_hash>.json
- [ ] Write manifest pattern per run (e.g., manifests/<city>/<run_id>.json)
- [ ] Implement cache get/put client (Go) with tests (mock S3)
- [ ] Add lifecycle rules for cold storage/expiration

5) Emit metrics/alerts per state; add tracing context propagation.
- [ ] Define metric names and dimensions (city, state, source)
- [ ] Add CloudWatch metrics emission from tasks
- [ ] Configure alarms for error rates / DLQ messages
- [ ] Add OTEL tracing stubs (Go) with propagation across steps
- [ ] Document correlation_id strategy and logging context
- [ ] Ensure state transitions and durations are recorded

Integrations readiness checklist
- [ ] IAM roles/policies scoped to least privilege
- [ ] Environment-specific configuration (dev/stage/prod)
- [ ] Runbook for common failures (DLQ, throttling)
- [ ] Cost guardrails (budgets, alerts)
- [ ] Observability dashboards (optional for MVP)

Links
- Terraform readme: ./terraform/README.md
- Go TDD readme: ./go/README.md