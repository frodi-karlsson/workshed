.PHONY: build test test-race test-integration test-e2e test-e2e-update test-all check help install clean lint lint-fix install-hooks

build:
	@mkdir -p bin
	go build -o bin/workshed ./cmd/workshed

test:
	go test -v ./... $(TESTARGS)

test-race:
	go test -race -v ./...

test-integration:
	go test -v -tags=integration ./...

test-e2e:
	go test -v -timeout=5s ./internal/tui/...

test-e2e-update:
	UPDATE_SNAPS=true go test -v -timeout=5s ./internal/tui/...

test-all:
	@echo "Running unit tests..."
	go test ./...
	@echo ""
	@echo "Running integration tests..."
	go test -v -tags=integration ./...
	@echo ""
	@echo "Running e2e tests..."
	go test -v -timeout=5s ./internal/tui/...

check:
	@echo "Running all checks..."
	@echo ""
	@echo "--- Lint ---"
	golangci-lint run ./...
	@echo ""
	@echo "--- Unit Tests ---"
	go test -timeout=30s ./...
	@echo ""
	@echo "--- Integration Tests ---"
	go test -v -tags=integration ./...
	@echo ""
	@echo "--- E2E Tests ---"
	go test -v -timeout=5s ./internal/tui/...

help:
	@echo "Available targets:"
	@echo ""
	@echo "  build          Build the workshed binary"
	@echo "  test           Run unit tests"
	@echo "  test-race      Run unit tests with race detector"
	@echo "  test-integration   Run integration tests"
	@echo "  test-e2e       Run e2e TUI tests with timeout"
	@echo "  test-e2e-update    Update e2e TUI test snapshots"
	@echo "  test-all       Run unit, integration, and e2e tests"
	@echo "  check          Run lint, unit, integration, and e2e tests"
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
