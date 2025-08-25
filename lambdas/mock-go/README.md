# Mock Go Lambda for Jaunt Orchestration

This Lambda mimics downstream states and returns a small, bounded payload for integration testing with the Step Functions state machine.

It mirrors the Python mock handler’s behavior:
- Reads STATE_NAME from env and returns it in the response (defaults to "mock").
- Echoes up to 5 items from the input.
- Returns new_unique_rate from input if present; otherwise defaults to 0.2.
- Preserves job context fields like job_id, city, s3_prefix, budgets, kill_switches, early_stop, timeouts.

## Prerequisites

- Go 1.22+
- AWS CLI v2, configured (`aws configure`)
- zip utility
- IAM permissions to create roles and Lambda functions

Repo layout (relevant):
- lambdas/mock-go/main.go
- lambdas/mock-go/go.mod
- lambdas/mock-go/Makefile (this file’s instructions)
- epics/orchestration-step-fns/examples/input.edinburgh.json (sample event)

## Quick start

1) Build and package (x86_64 by default)
- cd lambdas/mock-go
- make package
- For arm64:
  - make package ARCH=arm64

2) Create the execution role (one-time)
- make create-role
- This creates role "jaunt-mock-role" with AWSLambdaBasicExecutionRole attached.
- You can override ROLE_NAME if needed: make create-role ROLE_NAME=my-mock-role

3) Deploy (create or update the function)
- make deploy
- Override variables as needed:
  - make deploy REGION=us-east-1 ARCH=arm64 FUNCTION_NAME=jaunt-mock ROLE_NAME=jaunt-mock-role STATE_NAME=mock

4) Invoke with a sample event
- Use the provided example:
  - make invoke EVENT=../../epics/orchestration-step-fns/examples/input.edinburgh.json
- Output will be saved to response.json and logs printed to the console.

5) Tail logs
- make logs

6) Get function ARN
- make arn
- Use this ARN to wire the Step Functions state machine via Terraform (see below).

## Wiring into Step Functions via Terraform

- In your Terraform tfvars (e.g., epics/orchestration-step-fns/terraform/envs/mock.tfvars), set all lambda_*_arn entries to the ARN printed by `make arn`:
  - lambda_discover_web_sources_arn = "arn:aws:lambda:us-east-1:<acct>:function:jaunt-mock"
  - ...repeat for all states...
- Ensure your ASL definition uses template placeholders for Lambda ARNs and DLQ QueueUrl (already reflected in the proposed changes):
  - "Resource": "${lambda_web_fetch_arn}", etc.
  - "QueueUrl": "${frontier_dlq_url}"

- Apply:
  - cd epics/orchestration-step-fns/terraform
  - terraform apply -var-file=envs/mock.tfvars

- Start an execution:
  - aws stepfunctions start-execution --state-machine-arn <STATE_MACHINE_ARN> \
    --name city-discovery-edinburgh-001 \
    --input file://../examples/input.edinburgh.json

## Configuration

Environment variables on the Lambda:
- STATE_NAME: string label returned in the output (e.g., mock, WebFetch, ExtractWithLLM).
  - Update via:
    - make set-env STATE_NAME=WebFetch

Architectures:
- ARCH=x86_64 (default) => GOARCH=amd64
- ARCH=arm64          => GOARCH=arm64

Runtime:
- provided.al2 with a custom bootstrap (built by Go).

Region:
- Override with REGION=... for all Make targets.

## Clean up

- Delete function:
  - make delete
- Delete role:
  - make role-delete

## Troubleshooting

- AccessDenied or role errors:
  - Wait ~10–30 seconds after creating the role (IAM propagation).
  - Ensure your current AWS profile has permissions for IAM and Lambda.

- Mismatched architecture:
  - If you created the function with x86_64 but built for arm64, re-run:
    - make deploy ARCH=x86_64
  - The Makefile sets the correct GOARCH for the chosen ARCH.

- Missing AWS account id in ROLE_ARN:
  - The Makefile attempts to derive ACCOUNT_ID. If it’s blank, pass ROLE_ARN explicitly:
    - make deploy ROLE_ARN=arn:aws:iam::<acct-id>:role/<role-name>

- Logs show “Handler not found”:
  - Ensure handler is set to "bootstrap" and runtime is "provided.al2".
  - Ensure function.zip contains the "bootstrap" executable.

## Example end-to-end

- Build and deploy for arm64:
  - make deploy ARCH=arm64 REGION=us-east-1
- Set a visible state name:
  - make set-env STATE_NAME=WebFetch
- Invoke with sample:
  - make invoke EVENT=../../epics/orchestration-step-fns/examples/input.edinburgh.json
- Observe response.json and logs output.
