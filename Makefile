.PHONY: build run clean install test

BINARY_NAME=volta-iptv
MAIN_PATH=./cmd/volta-iptv

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

install: build
	cp bin/$(BINARY_NAME) $(HOME)/.local/bin/

clean:
	rm -rf bin/
	go clean

test:
	go test ./...

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run
