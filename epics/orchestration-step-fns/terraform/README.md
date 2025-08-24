# Terraform — Orchestration MVP

Resources
- Step Functions state machine (ASL skeleton)
- SQS: frontier queue + DLQ with redrive policy
- S3: raw cache bucket for source responses and run manifests
- CloudWatch Logs group for Step Functions
- IAM role/policy for Step Functions access (logs, SQS, S3)

Usage
- terraform init
- terraform plan -var-file=envs/dev.tfvars
- terraform apply -var-file=envs/dev.tfvars

Notes
- State machine definition lives in sfn/definition.asl.json as a first draft (Pass states) and should be iteratively replaced with real Task integrations (Lambda or service integrations).
- Names are derived from project_prefix and environment for easy isolation.

Destroy
- terraform destroy -var-file=envs/dev.tfvars