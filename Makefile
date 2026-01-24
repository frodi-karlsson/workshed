.PHONY: build test test-race test-integration install clean lint lint-fix install-hooks

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

.DEFAULT_GOAL := build
