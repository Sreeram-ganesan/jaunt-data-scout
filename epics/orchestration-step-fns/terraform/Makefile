# Terraform orchestrator Makefile for the Step Functions + SQS frontier project.
# Usage examples:
#   make init ENV=mock
#   make plan ENV=mock
#   make apply ENV=mock
#   make destroy ENV=mock
#   make outputs ENV=mock
#   make graph
# Variables:
#   ENV           : workspace/env label (default: mock). Uses envs/$(ENV).tfvars
#   AWS_PROFILE   : AWS profile (default: default)
#   AWS_REGION    : AWS region (default: us-east-1)
#   BACKEND_CONFIG: Optional backend config file (e.g., envs/backend.hcl)

SHELL := /bin/bash

# Environment selection
ENV          ?= mock
AWS_PROFILE  ?= default
AWS_REGION   ?= us-east-1
BACKEND_CONFIG ?=
# New: defaults for SFN execution testing
NAME         ?= city-discovery-edinburgh-001
INPUT        ?= input-mock.json

# New: Outputs (override as needed)
FRONTIER_DLQ_URL        ?= https://sqs.us-east-1.amazonaws.com/253490763512/jaunt-dev-frontier-dlq
FRONTIER_QUEUE_URL      ?= https://sqs.us-east-1.amazonaws.com/253490763512/jaunt-dev-frontier
S3_BUCKET_NAME          ?= jaunt-dev-data-scout-raw-253490763512
S3_RAW_CACHE_BUCKET     ?= jaunt-dev-data-scout-raw-253490763512
SQS_DLQ_QUEUE_URL       ?= https://sqs.us-east-1.amazonaws.com/253490763512/jaunt-dev-frontier-dlq
SQS_FRONTIER_QUEUE_URL  ?= https://sqs.us-east-1.amazonaws.com/253490763512/jaunt-dev-frontier
STATE_MACHINE_ARN       ?= arn:aws:states:us-east-1:253490763512:stateMachine:data-scout-orchestration-step-function

# Terraform locations/files
TF_DIR     := $(CURDIR)
VAR_FILE   ?= envs/$(ENV).tfvars
PLAN_FILE  ?= tfplan.$(ENV)
OUT_FILE   ?= outputs.$(ENV).json

export AWS_PROFILE
export AWS_REGION

# Helpers
# Fixed: correct -backend-config flag
INIT_BACKEND := $(if $(BACKEND_CONFIG),-backend-config=$(BACKEND_CONFIG),)

.PHONY: help
help:
	@echo "Targets:"
	@echo "  init        : terraform init (uses BACKEND_CONFIG if provided)"
	@echo "  workspace   : create/select terraform workspace $(ENV)"
	@echo "  plan        : terraform plan -var-file=$(VAR_FILE) -out=$(PLAN_FILE)"
	@echo "  apply       : terraform apply $(PLAN_FILE) (runs plan if missing)"
	@echo "  apply-auto  : terraform apply -auto-approve -var-file=$(VAR_FILE)"
	@echo "  destroy     : terraform destroy -var-file=$(VAR_FILE)"
	@echo "  outputs     : terraform output -json -> $(OUT_FILE)"
	@echo "  validate    : terraform validate"
	@echo "  fmt         : terraform fmt -recursive"
	@echo "  lockfile    : terraform providers lock for common platforms"
	@echo "  graph       : terraform graph -> graph.svg (requires graphviz 'dot')"
	@echo "  vars        : print resolved variables"
	@echo "  clean       : remove local plan/output artifacts"
	@echo "  start-exec  : start a Step Functions execution (requires STATE_MACHINE_ARN; optional INPUT and NAME)"
	@echo "  echo-outputs: print common resource outputs (URLs/ARN)"

.PHONY: vars
vars:
	@echo "ENV            : $(ENV)"
	@echo "AWS_PROFILE    : $(AWS_PROFILE)"
	@echo "AWS_REGION     : $(AWS_REGION)"
	@echo "TF_DIR         : $(TF_DIR)"
	@echo "VAR_FILE       : $(VAR_FILE)"
	@echo "PLAN_FILE      : $(PLAN_FILE)"
	@echo "BACKEND_CONFIG : $(BACKEND_CONFIG)"

.PHONY: init
init:
	@terraform -chdir=$(TF_DIR) init $(INIT_BACKEND) -upgrade

.PHONY: workspace
workspace:
	@terraform -chdir=$(TF_DIR) workspace new $(ENV) >/dev/null 2>&1 || true
	@terraform -chdir=$(TF_DIR) workspace select $(ENV)

.PHONY: plan
plan: workspace
	@if [ ! -f "$(VAR_FILE)" ]; then echo "Missing $(VAR_FILE). Create it or set VAR_FILE=..."; exit 1; fi
	@terraform -chdir=$(TF_DIR) plan -var-file=$(VAR_FILE) -out=$(PLAN_FILE)

.PHONY: apply
apply: workspace
	@if [ ! -f "$(PLAN_FILE)" ]; then echo "Plan file $(PLAN_FILE) not found. Running 'make plan'..."; $(MAKE) plan; fi
	@terraform -chdir=$(TF_DIR) apply $(PLAN_FILE)

.PHONY: apply-auto
apply-auto: workspace
	@if [ ! -f "$(VAR_FILE)" ]; then echo "Missing $(VAR_FILE). Create it or set VAR_FILE=..."; exit 1; fi
	@terraform -chdir=$(TF_DIR) apply -auto-approve -var-file=$(VAR_FILE)

.PHONY: destroy
destroy: workspace
	@if [ ! -f "$(VAR_FILE)" ]; then echo "Missing $(VAR_FILE). Create it or set VAR_FILE=..."; exit 1; fi
	@terraform -chdir=$(TF_DIR) destroy -var-file=$(VAR_FILE)

.PHONY: outputs
outputs:
	@terraform -chdir=$(TF_DIR) output -json | tee $(OUT_FILE) >/dev/null
	@$(MAKE) --no-print-directory echo-outputs

.PHONY: validate
validate:
	@terraform -chdir=$(TF_DIR) validate

.PHONY: fmt
fmt:
	@terraform -chdir=$(TF_DIR) fmt -recursive

.PHONY: lockfile
lockfile:
	@terraform -chdir=$(TF_DIR) providers lock \
	  -platform=linux_amd64 -platform=linux_arm64 \
	  -platform=darwin_amd64 -platform=darwin_arm64

.PHONY: graph
graph:
	@if command -v dot >/dev/null 2>&1; then \
	  terraform -chdir=$(TF_DIR) graph | dot -Tsvg > graph.svg && echo "Wrote graph.svg"; \
	else \
	  echo "'dot' (graphviz) not found; skipping graph.svg"; \
	fi

# New: Print outputs block
.PHONY: echo-outputs
echo-outputs:
	@echo "Outputs:"
	@echo "frontier_dlq_url = \"$(FRONTIER_DLQ_URL)\""
	@echo "frontier_queue_url = \"$(FRONTIER_QUEUE_URL)\""
	@echo "s3_bucket_name = \"$(S3_BUCKET_NAME)\""
	@echo "s3_raw_cache_bucket = \"$(S3_RAW_CACHE_BUCKET)\""
	@echo "sqs_dlq_queue_url = \"$(SQS_DLQ_QUEUE_URL)\""
	@echo "sqs_frontier_queue_url = \"$(SQS_FRONTIER_QUEUE_URL)\""
	@echo "state_machine_arn = \"$(STATE_MACHINE_ARN)\""

# New: Start a Step Functions execution from a JSON input
.PHONY: start-exec
start-exec:
	@$(MAKE) --no-print-directory echo-outputs
	@if [ -z "$(STATE_MACHINE_ARN)" ]; then echo "Set STATE_MACHINE_ARN=arn:aws:states:...:stateMachine:..."; exit 1; fi
	@if [ ! -f "$(INPUT)" ]; then echo "Input file '$(INPUT)' not found. Override with INPUT=..."; exit 1; fi
	@echo "Starting execution: $(NAME)"
	@aws stepfunctions start-execution \
	  --state-machine-arn "$(STATE_MACHINE_ARN)" \
	  --name "$(NAME)" \
	  --input file://$(INPUT) \
	  --region $(AWS_REGION)

.PHONY: clean
clean:
	@rm -f $(PLAN_FILE) $(OUT_FILE)
