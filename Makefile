.DEFAULT_GOAL := test

GOFILES := $(shell find . -name '*.go' -not -path './.git/*')

.PHONY: test fmt tidy lint build bench cover clean help

## test: Run all tests
test:
	go test ./...

## test-v: Run all tests with verbose output
test-v:
	go test -v ./...

## test-short: Run tests in short mode (skip integration tests)
test-short:
	go test -short ./...

## bench: Run benchmarks
bench:
	go test -bench=. -benchmem ./...

## cover: Run tests with coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format all Go files
fmt:
	gofmt -w $(GOFILES)

## tidy: Tidy go modules
tidy:
	go mod tidy

## lint: Run linter
lint:
	golangci-lint run ./...

## build: Build all packages
build:
	go build ./...

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -cache -testcache

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
