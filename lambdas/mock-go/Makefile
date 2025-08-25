# Minimal Makefile to build, package, deploy, and test the Go mock Lambda.
# Usage examples:
#   make package                          # Build for your chosen arch (defaults to x86_64)
#   make package ARCH=arm64               # Build for arm64
#   make create-role                      # Create an execution role
#   make deploy                           # Create or update the Lambda function
#   make set-env STATE_NAME=WebFetch      # Update env var on the function
#   make invoke EVENT=../../epics/orchestration-step-fns/examples/input.edinburgh.json
#   make logs                             # Tail CloudWatch logs
#   make arn                              # Print the function ARN
#   make delete                           # Delete the function
#   make role-delete                      # Delete the role
#   make vars                             # Print resolved arch variables (debug)
#   make clean                            # Remove build artifacts

SHELL := /bin/bash

# Configurable parameters
FUNCTION_NAME ?= jaunt-mock
ROLE_NAME     ?= jaunt-mock-role
REGION        ?= us-east-1
RUNTIME       ?= provided.al2
STATE_NAME    ?= mock

# Architecture handling
# Prefer LAMBDA_ARCH, but allow ARCH as an alias (often set in envs)
LAMBDA_ARCH ?= x86_64
ifdef ARCH
  LAMBDA_ARCH := $(ARCH)
endif

# Normalize and map common synonyms:
# - x86_64, amd64 => GOARCH=amd64, Lambda architectures=x86_64
# - arm64, aarch64 => GOARCH=arm64, Lambda architectures=arm64
LAMBDA_ARCH_LOWER := $(shell echo "$(LAMBDA_ARCH)" | tr '[:upper:]' '[:lower:]' | tr -d ' _-')
ifeq ($(LAMBDA_ARCH_LOWER),x8664)
  GOARCH := amd64
  LAMBDA_ARCH_NORM := x86_64
else ifeq ($(LAMBDA_ARCH_LOWER),amd64)
  GOARCH := amd64
  LAMBDA_ARCH_NORM := x86_64
else ifeq ($(LAMBDA_ARCH_LOWER),arm64)
  GOARCH := arm64
  LAMBDA_ARCH_NORM := arm64
else ifeq ($(LAMBDA_ARCH_LOWER),aarch64)
  GOARCH := arm64
  LAMBDA_ARCH_NORM := arm64
else
  $(error ARCH/LAMBDA_ARCH must be one of: x86_64, amd64, arm64, aarch64)
endif

BOOTSTRAP := bootstrap
ZIP       := function.zip
OUT       ?= response.json

# Attempt to resolve account id for role arn construction
ACCOUNT_ID := $(shell aws sts get-caller-identity --query Account --output text 2>/dev/null || true)
ROLE_ARN   ?= arn:aws:iam::$(ACCOUNT_ID):role/$(ROLE_NAME)

.PHONY: help
help:
	@echo "Targets:"
	@echo "  package        Build $(BOOTSTRAP) for $(LAMBDA_ARCH_NORM) and create $(ZIP)"
	@echo "  build          Build $(BOOTSTRAP) for $(LAMBDA_ARCH_NORM)"
	@echo "  zip            Create $(ZIP) from $(BOOTSTRAP)"
	@echo "  deps           Download modules to populate go.sum"
	@echo "  create-role    Create IAM role $(ROLE_NAME) with basic logging"
	@echo "  deploy         Create or update Lambda $(FUNCTION_NAME)"
	@echo "  update-code    Update only the function code"
	@echo "  set-env        Update environment variables (ex: STATE_NAME)"
	@echo "  invoke         Invoke function with EVENT file and print logs"
	@echo "  logs           Tail CloudWatch logs for the function"
	@echo "  arn            Print the function ARN"
	@echo "  delete         Delete the function"
	@echo "  role-delete    Detach policy and delete the role"
	@echo "  fmt            go fmt"
	@echo "  tidy           go mod tidy"
	@echo "  vars           Print resolved variables"
	@echo "  clean          Remove build artifacts"

.PHONY: vars
vars:
	@echo "ARCH (alias)         : $(ARCH)"
	@echo "LAMBDA_ARCH (input)  : $(LAMBDA_ARCH)"
	@echo "LAMBDA_ARCH_NORM     : $(LAMBDA_ARCH_NORM)"
	@echo "GOARCH               : $(GOARCH)"
	@echo "FUNCTION_NAME        : $(FUNCTION_NAME)"
	@echo "ROLE_NAME            : $(ROLE_NAME)"
	@echo "REGION               : $(REGION)"
	@echo "ROLE_ARN             : $(ROLE_ARN)"

.PHONY: package
package: build zip

.PHONY: deps
deps:
	@echo "Ensuring module dependencies are downloaded..."
	GOOS=linux GOARCH=$(GOARCH) go mod download

.PHONY: build
build: deps
	@echo "Building $(FUNCTION_NAME) for linux/$(GOARCH) (Lambda: $(LAMBDA_ARCH_NORM)) ..."
	GOOS=linux GOARCH=$(GOARCH) CGO_ENABLED=0 go build -o $(BOOTSTRAP)

.PHONY: zip
zip:
	@echo "Zipping into $(ZIP) ..."
	rm -f $(ZIP)
	zip -9 $(ZIP) $(BOOTSTRAP) >/dev/null

.PHONY: create-role
create-role:
	@set -euo pipefail; \
	if aws iam get-role --role-name $(ROLE_NAME) >/dev/null 2>&1; then \
	  echo "Role $(ROLE_NAME) already exists"; \
	else \
	  echo "Creating role $(ROLE_NAME) ..."; \
	  aws iam create-role \
	    --role-name $(ROLE_NAME) \
	    --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}' >/dev/null; \
	  aws iam attach-role-policy \
	    --role-name $(ROLE_NAME) \
	    --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole >/dev/null; \
	  echo "Role created and policy attached. Waiting for IAM propagation..."; \
	  sleep 10; \
	fi

.PHONY: deploy
deploy: package
	@set -euo pipefail; \
	if aws lambda get-function --function-name $(FUNCTION_NAME) --region $(REGION) >/dev/null 2>&1; then \
	  echo "Function exists. Checking architecture ..."; \
	  CURR_ARCH=$$(aws lambda get-function-configuration --function-name $(FUNCTION_NAME) --query 'Architectures[0]' --output text --region $(REGION) 2>/dev/null || true); \
	  echo "Current: $$CURR_ARCH, Desired: $(LAMBDA_ARCH_NORM)"; \
	  if [ "$$CURR_ARCH" != "$(LAMBDA_ARCH_NORM)" ] && [ "$$CURR_ARCH" != "None" ]; then \
	    echo "Updating architecture to $(LAMBDA_ARCH_NORM) ..."; \
	    aws lambda update-function-configuration \
	      --function-name $(FUNCTION_NAME) \
	      --architectures $(LAMBDA_ARCH_NORM) \
	      --region $(REGION) >/dev/null; \
	    echo "Waiting for configuration update to complete ..."; \
	    until [ "$$(aws lambda get-function-configuration --function-name $(FUNCTION_NAME) --query 'LastUpdateStatus' --output text --region $(REGION))" = "Successful" ]; do \
	      sleep 2; \
	    done; \
	  fi; \
	  echo "Updating code for $(FUNCTION_NAME) ..."; \
	  aws lambda update-function-code \
	    --function-name $(FUNCTION_NAME) \
	    --zip-file fileb://$(ZIP) \
	    --region $(REGION) >/dev/null; \
	  echo "Updated $(FUNCTION_NAME)."; \
	else \
	  echo "Creating function $(FUNCTION_NAME) (arch $(LAMBDA_ARCH_NORM)) ..."; \
	  aws lambda create-function \
	    --function-name $(FUNCTION_NAME) \
	    --runtime $(RUNTIME) \
	    --handler $(BOOTSTRAP) \
	    --role $(ROLE_ARN) \
	    --zip-file fileb://$(ZIP) \
	    --architectures $(LAMBDA_ARCH_NORM) \
	    --region $(REGION) \
	    --environment Variables="{STATE_NAME=$(STATE_NAME)}" >/dev/null; \
	  echo "Created $(FUNCTION_NAME)."; \
	fi

.PHONY: update-code
update-code: package
	aws lambda update-function-code \
	  --function-name $(FUNCTION_NAME) \
	  --zip-file fileb://$(ZIP) \
	  --region $(REGION)

.PHONY: set-env
set-env:
	@if [ -z "$(STATE_NAME)" ]; then echo "STATE_NAME is empty; pass STATE_NAME=..."; exit 1; fi
	aws lambda update-function-configuration \
	  --function-name $(FUNCTION_NAME) \
	  --environment Variables="{STATE_NAME=$(STATE_NAME)}" \
	  --region $(REGION)

.PHONY: invoke
invoke:
	@if [ -z "$(EVENT)" ]; then echo "Please set EVENT to a JSON file path, e.g. EVENT=../../epics/orchestration-step-fns/examples/input.edinburgh.json"; exit 1; fi
	@echo "Invoking $(FUNCTION_NAME) with $(EVENT) ..."
	@aws lambda invoke \
	  --function-name $(FUNCTION_NAME) \
	  --payload fileb://$(EVENT) \
	  --log-type Tail \
	  --query 'LogResult' \
	  --output text \
	  --region $(REGION) \
	  $(OUT) | base64 --decode || true
	@echo
	@echo "Response saved to $(OUT)"

.PHONY: logs
logs:
	aws logs tail "/aws/lambda/$(FUNCTION_NAME)" --follow --region $(REGION)

.PHONY: arn
arn:
	@aws lambda get-function --function-name $(FUNCTION_NAME) --query 'Configuration.FunctionArn' --output text --region $(REGION)

.PHONY: delete
delete:
	-aws lambda delete-function --function-name $(FUNCTION_NAME) --region $(REGION)

.PHONY: role-delete
role-delete:
	-aws iam detach-role-policy --role-name $(ROLE_NAME) --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
	-aws iam delete-role --role-name $(ROLE_NAME)

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -f $(BOOTSTRAP) $(ZIP)

.PHONY: arch
arch:
	@aws lambda get-function-configuration \
	  --function-name $(FUNCTION_NAME) \
	  --query 'Architectures[0]' \
	  --output text \
	  --region $(REGION)

.PHONY: set-arch
set-arch:
	@echo "Setting architecture to $(LAMBDA_ARCH_NORM) for $(FUNCTION_NAME) ..."
	@aws lambda update-function-configuration \
	  --function-name $(FUNCTION_NAME) \
	  --architectures $(LAMBDA_ARCH_NORM) \
	  --region $(REGION) >/dev/null
	@echo "Waiting for configuration update to complete ..."
	@until [ "$$(aws lambda get-function-configuration --function-name $(FUNCTION_NAME) --query 'LastUpdateStatus' --output text --region $(REGION))" = "Successful" ]; do \
	  sleep 2; \
	done
	@echo "Done."