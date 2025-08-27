.PHONY: test cover tidy lint build-tools dlq-redrive

test:
	go test ./... -count=1

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

tidy:
	go mod tidy

lint:
	@golangci-lint run || echo "golangci-lint not installed or issues found"

build-tools: dlq-redrive

dlq-redrive:
	go build -o bin/dlq-redrive ./cmd/dlq-redrive
	@echo "Built DLQ re-drive tool: bin/dlq-redrive"
	@echo "Usage: ./bin/dlq-redrive list --help"