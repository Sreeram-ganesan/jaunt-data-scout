.PHONY: test cover tidy lint

test:
	go test ./... -count=1

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

tidy:
	go mod tidy

lint:
	@golangci-lint run || echo "golangci-lint not installed or issues found"