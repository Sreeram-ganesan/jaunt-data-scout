# Orchestration Terraform — Project Details

This folder provisions the orchestration scaffolding for the Data Scout MVP:
- SQS frontier queue with DLQ
- Step Functions state machine (ASL) parameterized with Lambda ARNs
- IAM role/policies for the state machine
- Optional S3 buckets (raw cache/manifests) and CloudWatch Logs for SFN

Note: Lambda functions are deployed separately (e.g., jaunt-mock). Terraform references their ARNs via tfvars.

## Folder layout (expected)

- main.tf, variables.tf, outputs.tf — root composition wiring all modules
- modules/
  - stepfn_orchestrator/ — State Machine, logging, IAM execution role
  - sqs_frontier/ — frontier queue + DLQ + redrive policy
  - s3_cache/ (optional) — S3 buckets and lifecycle (raw, extracted, manifests)
  - cloudwatch/ (optional) — dashboards/alarms
- envs/
  - mock.tfvars — example values (all lambda_*_arn pointing to the mock lambda)
  - prod.tfvars — production overrides

Your ASL definition should use template placeholders for Lambda ARNs and DLQ URL, e.g.:
- "Resource": "${lambda_web_fetch_arn}"
- "QueueUrl": "${frontier_dlq_url}"

## Inputs (tfvars)

Minimal set (example):
- name_prefix = "jaunt-mvp"
- region = "us-east-1"

Lambda ARNs (mock wiring — all point to the same mock function ARN):
- lambda_discover_web_sources_arn
- lambda_discover_targets_arn
- lambda_seed_primaries_arn
- lambda_expand_neighbors_arn
- lambda_tile_sweep_arn
- lambda_web_fetch_arn
- lambda_extract_with_llm_arn
- lambda_geocode_validate_arn
- lambda_dedupe_canonicalize_arn
- lambda_persist_arn
- lambda_rank_arn
- lambda_finalize_arn

Frontier queues:
- frontier_queue_name
- frontier_dlq_name
- frontier_max_receive_count (default 5)

Optional S3:
- s3_raw_bucket_name
- s3_extracted_bucket_name
- s3_manifests_bucket_name

## Resources created

SQS (module: sqs_frontier)
- aws_sqs_queue.frontier
- aws_sqs_queue.frontier_dlq
- aws_sqs_queue_policy (optional)
- Redrive policy linking frontier → DLQ

Step Functions (module: stepfn_orchestrator)
- aws_sfn_state_machine.orchestrator with logging (CloudWatch Logs)
- aws_iam_role.sfn_exec and aws_iam_role_policy/attachment for:
  - lambda:InvokeFunction on provided Lambda ARNs
  - sqs:SendMessage/DeleteMessage on frontier/DLQ
  - logs:CreateLogDelivery/PutResourcePolicy (for SFN logs)

Optional S3 (module: s3_cache)
- aws_s3_bucket resources with lifecycle rules:
  - raw/json (Google 30d; open-data 90d)
  - raw/html (90d)
  - extracted (90d)
  - manifests (90d)

CloudWatch (module: cloudwatch)
- dashboards/alarms for error rates, DLQ depth, SFN failures

## Using the Makefile

Prereqs:
- terraform >= 1.5
- AWS CLI configured (AWS_PROFILE) and an account with IAM/SFN/SQS/S3 permissions

Quick start (mock env):
- Create envs/mock.tfvars with Lambda ARNs and queue names (all ARNs may point to the single mock lambda ARN you deployed):
  - lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:<acct>:function:jaunt-mock"
  - ...repeat for all states...
  - frontier_queue_name = "jaunt-frontier-mock"
  - frontier_dlq_name = "jaunt-frontier-mock-dlq"

- Initialize and select workspace:
  - make init ENV=mock
  - make workspace ENV=mock

- Plan and apply:
  - make plan ENV=mock
  - make apply ENV=mock

- View outputs:
  - make outputs ENV=mock

- Update formatting / validate:
  - make fmt
  - make validate

- Destroy (teardown):
  - make destroy ENV=mock

Notes:
- Set AWS_PROFILE and AWS_REGION if needed:
  - AWS_PROFILE=dev AWS_REGION=us-east-1 make plan ENV=mock
- Backend config (remote state) can be passed via BACKEND_CONFIG:
  - make init ENV=mock BACKEND_CONFIG=envs/backend.hcl

## Tips

- Keep all lambda_*_arn in tfvars aligned with the functions deployed by the Go mock (or real connectors later).
- The ASL definition should emit state logs to CloudWatch and route failures to the DLQ.
- Use workspaces to separate mock/prod: make workspace ENV=prod then plan/apply with envs/prod.tfvars.
