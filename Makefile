.DEFAULT_GOAL := test

GOFILES := $(shell find . -name '*.go' -not -path './.git/*')

.PHONY: test fmt tidy lint

test:
	go test ./...

fmt:
	gofmt -w $(GOFILES)

tidy:
	go mod tidy

lint:
	golangci-lint run ./...
