# data-scout — AWS Architecture and Module Overview

Concise overview of how Data Scout is implemented on AWS: Step Functions orchestrates city discovery; SQS acts as the frontier/DLQ; Lambda runs connectors/processors; S3 stores raw/extracted/manifests; IAM secures access; CloudWatch/X-Ray provide observability. Terraform under epics/orchestration-step-fns/terraform provisions core infra.

## Current Implementation Status

- ✅ **Step Functions Orchestrator**: Complete workflow with mock/real feature flags
- ✅ **SQS Frontier & DLQ**: Message queuing with redrive capability  
- ✅ **DiscoverWebSources Lambda**: Tavily API integration for web source discovery
- 🚧 **Other Lambdas**: WebFetch, ExtractWithLLM, GeocodeValidate, etc. (planned)

The **DiscoverWebSources** Lambda is the first real implementation in the web discovery path. It can be toggled between mock and real via the `discover_web_sources` feature flag in Terraform configurations.

## System Overview (AWS)

- Orchestrator: AWS Step Functions State Machine (city job lifecycle).
- Frontier: Amazon SQS queue with DLQ for failed messages.
- Compute: AWS Lambda functions (mock provided; real connectors planned).
- Storage: Amazon S3 (raw html/json, extracted entities, manifests). Existing Postgres used by Persist stage (outside this repo).
- Observability: CloudWatch Logs/Metrics/Alarms; AWS X-Ray/OTEL traces.
- Security: IAM least-privilege; SSE-S3/KMS (optional); secrets via Secrets Manager.

## Module Breakdown (by AWS service)

1) Step Functions — Orchestrator
- Responsibility: Drive the workflow: DiscoverWebSources → WebFetch → ExtractWithLLM → GeocodeValidate → DedupeCanonicalize → Persist → Rank → Finalize. Maps/Seed/Expand/TileSweep states are planned.
- Inputs: StartExecution payload (city, budgets, kill_switches, s3_prefix, timeouts).
- Outputs: Invokes Lambda tasks; gates on budgets/early-stop; writes status to CW Logs.
- Depends on: IAM role with lambda:InvokeFunction, sqs:*, logs:*.
- Interfaces: ASL state machine, CloudWatch Logs.

2) SQS Frontier + DLQ
- Responsibility: Decouple discovery from fetching; buffer map/web tasks; redrive failures.
- Inputs: Messages from discovery states (web/maps union schema).
- Outputs: Messages consumed by Web Fetcher, Maps workers; failures to DLQ.
- Depends on: Redrive policy; IAM for send/receive/delete.
- Interfaces: SQS queue URLs.

3) Lambda — Mock Handler (this repo)
- Responsibility: Test/dummy implementation for all states; echoes context, caps items, returns status “ok”.
- Inputs: Event JSON from Step Functions or invoke.
- Outputs: status, state, items[], new_unique_rate, job context fields.
- Depends on: Runtime env var STATE_NAME; IAM basic execution.
- Interfaces: Runtime provided.al2; handler=bootstrap.
- Notes: See lambdas/mock-go (Makefile, README, main.go).

4) Lambda — Connectors/Processors (planned)
- **Tavily Connector (DiscoverWebSources): IMPLEMENTED** - discover web sources; enqueue {type:web,...} to SQS.
- Web Fetcher: robots-aware fetch; write raw html/json to S3; hand off to ExtractWithLLM.
- ExtractWithLLM: bounded JSON extraction; write extracted JSON to S3.
- GeocodeValidate: fill/validate coords; quality flags.
- DedupeCanonicalize: similarity rules; reversible merges; lineage.
- Persist: upsert to Postgres; write manifests to S3.
- Rank: compute primary/secondary scores.
- Finalize: summarize counts/metrics.
- Interfaces: Lambda invoke, S3 SDK, SQS SDK, external APIs (Tavily/OSM/etc.).

5) Storage — Amazon S3
- Responsibility: Durable object storage and manifests.
- Layout (examples):
  - raw/html/<city>/<domain>/<content_hash>.html
  - raw/json/<city>/<source>/<request_hash>.json
  - extracted/<city>/<run_id>/<content_hash>.json
  - manifests/<city>/<run_id>.json
- Policies: SSE-S3/KMS; lifecycle (30–90 days) per data class (see Terraform notes).

6) IAM
- Responsibility: Execution roles/policies for SFN and Lambda; SQS access; S3 access; logs.
- Practices: Least privilege; per-function roles where practical.

7) Observability — CloudWatch + X-Ray/OTEL
- Responsibility: Logs per Lambda; SFN execution logs; metrics/alarms (DLQ depth, failures, durations); traces across states.
- Interfaces: CloudWatch Logs/Metrics, X-Ray segments, OTEL SDKs.

8) Secrets & Config
- Responsibility: Store API keys and sensitive config in AWS Secrets Manager; non-secret config in SSM Parameter Store or env vars.
- Rotation: Prefer managed rotation where available.

## Data Flow (AWS)

- StepFunctions (city job) → optional Discover states enqueue to SQS Frontier.
- Lambda WebFetcher/Maps workers dequeue → fetch external content → write to S3 (raw) → invoke ExtractWithLLM → write to S3 (extracted).
- GeocodeValidate → DedupeCanonicalize → Persist (to Postgres) → Rank → Finalize.
- Failures: retries with backoff; poison messages to DLQ; manual redrive.

## Reliability & Scaling

- SQS decoupling with DLQ; maxReceiveCount guards.
- Lambda concurrency scales per state; idempotent keys (request_hash/content_hash).
- Step Functions retries/catch blocks; early-stop gates via budget/new_unique_rate.
- S3 lifecycle to control storage costs.

## Security

- IAM least privilege; encrypted in-transit and at-rest (S3 SSE; KMS optional).
- Secrets in Secrets Manager; no secrets in logs.
- Optional VPC egress allowlist for fetchers.

## Observability

- CloudWatch: logs, dashboards, alarms (SFN failure rates, DLQ depth, Lambda errors).
- Tracing: X-Ray/OTEL spans across WebFetch → ExtractWithLLM → Persist.
- Budget/cost metrics: LLM tokens, HTTP bytes, calls per connector.

## Repo Mapping

- lambdas/mock-go: mock Lambda, Makefile to package/deploy/invoke; used to wire all states during infra bring-up.
- epics/orchestration-step-fns/terraform: SQS frontier + DLQ, Step Functions, IAM, optional S3/logs; Makefile to init/plan/apply.
- epics/orchestration-step-fns/documentation: PlantUML system design diagram and README.

## Appendix

- Replace mock with real Lambdas incrementally; keep ASL placeholders and DLQ wiring unchanged.
- Keep state outputs schema-stable (job_id, city, s3_prefix, budgets, kill_switches, early_stop, timeouts, items, new_unique_rate).
