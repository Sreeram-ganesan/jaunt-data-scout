# Epic: Orchestration — Step Functions workflow for city discovery

This directory contains:
- terraform/: IaC for Step Functions state machine, SQS (frontier, DLQ), S3 raw cache, CloudWatch logs, and minimal IAM.
- go/: Test-driven Go boilerplate for orchestration helpers (budget guard, queue/cache abstractions, metrics/tracing stubs, and a sample runner).
- TRACKING.md: work breakdown with subtasks and checklists aligned to the epic acceptance criteria.

References
- Epic issue: https://github.com/Sreeram-ganesan/jaunt-data-scout/issues/3

Getting started

Terraform
- cd terraform
- terraform init
- terraform plan -var-file=envs/dev.tfvars
- terraform apply -var-file=envs/dev.tfvars

Golang
- cd go
- make test
- Start by unskipping tests under internal/... and drive implementations.

Notes
- The Step Functions definition is a placeholder (Pass states) in sfn/definition.asl.json; wire to real compute as tasks are implemented.
- Budgets and early stop logic are prototyped in Go (BudgetGuard) to support both local simulation and future Lambda tasks.