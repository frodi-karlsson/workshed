.PHONY: build test test-race test-integration test-all check help install clean lint lint-fix install-hooks

build:
	@mkdir -p bin
	go build -o bin/workshed ./cmd/workshed

test:
	go test -v ./... $(TESTARGS)

test-race:
	go test -race -v ./...

test-integration:
	go test -v -tags=integration ./...

test-all:
	@echo "Running unit tests..."
	go test ./...
	@echo ""
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

check:
	@echo "Running all checks..."
	@echo ""
	@echo "--- Lint ---"
	golangci-lint run ./...
	@echo ""
	@echo "--- Unit Tests ---"
	go test ./...
	@echo ""
	@echo "--- Integration Tests ---"
	go test -v -tags=integration ./...

help:
	@echo "Available targets:"
	@echo ""
	@echo "  build          Build the workshed binary"
	@echo "  test           Run unit tests"
	@echo "  test-race      Run unit tests with race detector"
	@echo "  test-integration   Run integration tests"
	@echo "  test-all       Run unit and integration tests"
	@echo "  check          Run lint, unit tests, and integration tests"
	@echo "  install        Install workshed to \$$GOPATH/bin"
	@echo "  clean          Remove build artifacts"
	@echo "  lint           Run golangci-lint"
	@echo "  lint-fix       Run golangci-lint with auto-fixes"
	@echo "  install-hooks  Install pre-commit hook"
	@echo ""
	@echo "  help           Show this help message"

install:
	go install ./cmd/workshed

clean:
	rm -rf bin
	go clean

lint:
	@echo "🔍 Running golangci-lint on all Go files..."
	golangci-lint run ./...

lint-fix:
	@echo "🔧 Running golangci-lint with auto-fixes..."
	golangci-lint run --fix ./...
	@echo "🎨 Auto-fixing formatting..."
	gofmt -w .
	@echo "📦 Auto-fixing imports..."
	goimports -w .
	@echo "✅ Auto-fixes applied"

install-hooks:
	@echo "📦 Installing pre-commit hook..."
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"

.DEFAULT_GOAL := help
