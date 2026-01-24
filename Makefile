.PHONY: build test test-race test-integration install clean

build:
	@mkdir -p bin
	go build -o bin/workshed ./cmd/workshed

test:
	go test -v ./... $(TESTARGS)

test-race:
	go test -race -v ./...

test-integration:
	go test -v -tags=integration ./...

install:
	go install ./cmd/workshed

clean:
	rm -rf bin
	go clean

.DEFAULT_GOAL := build
