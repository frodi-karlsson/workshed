#!/bin/bash

# Install script for workshed development tools
# This script installs the pre-commit hook and provides helpful commands

set -e

echo "🛠️  Setting up workshed development environment..."

# Check if golangci-lint is installed
if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "❌ golangci-lint not found. Installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Check if goimports is installed
if ! command -v goimports >/dev/null 2>&1; then
    echo "❌ goimports not found. Installing..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# Install pre-commit hook
echo "📦 Installing pre-commit hook..."
cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "📋 Available commands:"
echo "  make lint       - Run linter on all files"
echo "  make lint-fix   - Auto-fix linting issues"
echo "  make install-hooks - Reinstall pre-commit hook"
echo ""
echo "🔧 The pre-commit hook will now run automatically before each commit."
echo "   It will lint your changed Go files and auto-fix formatting issues."
echo ""
echo "⚠️  To bypass the hook (emergency only):"
echo "   git commit --no-verify"