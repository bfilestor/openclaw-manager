BINARY := bin/managerd

.PHONY: build run test lint

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run ./...
