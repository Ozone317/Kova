.PHONY: lint fmt check

fmt:
	gofmt -w .
	goimports -w .

lint:
	golangci-lint run ./...

check: fmt lint
