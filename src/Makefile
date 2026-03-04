BINARY := bin/managerd

.PHONY: build run test test-unit test-integration test-coverage lint

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test: test-unit

test-unit:
	go test ./...

test-integration:
	go test ./... -tags integration

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...
