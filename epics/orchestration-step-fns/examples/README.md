How to run a mock execution

1) Deploy infra (dev)
- cd epics/orchestration-step-fns/terraform
- terraform init
- terraform plan -var-file=envs/dev.tfvars
- terraform apply -var-file=envs/dev.tfvars

2) Create a single mock Lambda
- Package and deploy lambdas/mock/handler.py as a Lambda (e.g., function name: jaunt-mock).
- Capture its ARN.

3) Route all states to the mock
- Copy envs/mock.tfvars and set all lambda_*_arn entries to your mock Lambda ARN.

4) Apply with mock tfvars (optional)
- terraform apply -var-file=envs/mock.tfvars

5) Start an execution
- aws stepfunctions start-execution --state-machine-arn <STATE_MACHINE_ARN> --name city-discovery-edinburgh-001 --input file://epics/orchestration-step-fns/examples/input.edinburgh.json

Notes
- Ensure sfn/definition.asl.json uses template placeholders for Lambda ARNs and ${frontier_dlq_url} for the SQS QueueUrl (already done in this commit content).
- You can set STATE_NAME="WebFetch", etc., on the Lambda to see which state returned the output.