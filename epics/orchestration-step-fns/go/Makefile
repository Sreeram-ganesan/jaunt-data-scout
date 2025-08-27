.PHONY: test cover tidy lint build-tools dlq-redrive local-runner all

test:
	go test ./... -count=1

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

tidy:
	go mod tidy

lint:
	@golangci-lint run || echo "golangci-lint not installed or issues found"

build-tools: dlq-redrive local-runner

all: build-tools

dlq-redrive:
	go build -o bin/dlq-redrive ./cmd/dlq-redrive
	@echo "Built DLQ re-drive tool: bin/dlq-redrive"
	@echo "Usage: ./bin/dlq-redrive list --help"

local-runner:
	go build -o bin/local-runner ./cmd/local-runner
	@echo "Built local workflow runner: bin/local-runner"
	@echo "Usage: ./bin/local-runner [input-file.json]"
	@echo "Default input: examples/input.local-edinburgh.json"

run-edinburgh: local-runner
	@echo "Running Edinburgh workflow..."
	./bin/local-runner examples/input.local-edinburgh.json

clean:
	rm -rf bin/ output/ coverage.out