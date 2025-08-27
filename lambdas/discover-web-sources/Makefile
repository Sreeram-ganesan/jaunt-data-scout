# Minimal Makefile to build, package, deploy, and test the DiscoverWebSources Lambda.
# Usage examples:
#   make package                          # Build for your chosen arch (defaults to x86_64)
#   make package ARCH=arm64               # Build for arm64
#   make deploy                           # Create or update the Lambda function
#   make invoke EVENT=./test-input.json
#   make logs                             # Tail CloudWatch logs

SHELL := /bin/bash

# Configurable parameters
FUNCTION_NAME ?= jaunt-dev-discover-web-sources
ROLE_NAME     ?= jaunt-dev-discover-web-sources-role
REGION        ?= us-east-1
RUNTIME       ?= provided.al2

# Architecture handling
LAMBDA_ARCH ?= x86_64
ifdef ARCH
  LAMBDA_ARCH := $(ARCH)
endif

# Normalize and map common synonyms
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
	@echo "  test           Run tests"
	@echo "  deploy         Create or update Lambda $(FUNCTION_NAME)"
	@echo "  invoke         Invoke function with EVENT file"
	@echo "  logs           Tail CloudWatch logs"
	@echo "  clean          Remove build artifacts"

.PHONY: package
package: build zip

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
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

.PHONY: test
test:
	go test ./... -v

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -f $(BOOTSTRAP) $(ZIP)

.PHONY: logs
logs:
	aws logs tail "/aws/lambda/$(FUNCTION_NAME)" --follow --region $(REGION)

.PHONY: invoke
invoke:
	@if [ -z "$(EVENT)" ]; then echo "Please set EVENT to a JSON file path"; exit 1; fi
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