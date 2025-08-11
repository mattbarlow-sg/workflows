#!/usr/bin/env just --justfile

# Install development tools
install-tools:
    @echo "Installing Go development tools..."
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "Tools installed successfully!"

# Format all Go files (basic, always available)
format:
    gofmt -s -w .
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    else \
        echo "Note: goimports not installed. Run 'just install-tools' to install it."; \
    fi

# Format specific directory
format-dir DIR:
    gofmt -s -w {{DIR}}
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w {{DIR}}; \
    else \
        echo "Note: goimports not installed. Run 'just install-tools' to install it."; \
    fi

# Check if code is formatted (for CI)
format-check:
    @echo "Checking Go formatting..."
    @if [ -n "$(gofmt -s -l .)" ]; then \
        echo "The following files need formatting:"; \
        gofmt -s -l .; \
        exit 1; \
    fi
    @if command -v goimports >/dev/null 2>&1; then \
        echo "Checking goimports..."; \
        if [ -n "$(goimports -l .)" ]; then \
            echo "The following files need import formatting:"; \
            goimports -l .; \
            exit 1; \
        fi; \
    else \
        echo "Warning: goimports not installed, skipping import check."; \
    fi
    @echo "All files are properly formatted!"

# Run linter
lint:
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run ./...; \
    else \
        echo "golangci-lint not installed. Run 'just install-tools' to install it."; \
        echo "Falling back to 'go vet'..."; \
        go vet ./...; \
    fi

# Quick lint (faster, less thorough)
lint-quick:
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fast ./...; \
    else \
        echo "golangci-lint not installed. Run 'just install-tools' to install it."; \
        echo "Falling back to 'go vet'..."; \
        go vet ./...; \
    fi

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Build the project
build:
    go build -o workflows ./cmd/workflows

# Clean build artifacts
clean:
    rm -f workflows
    rm -f coverage.out coverage.html

# Run all checks (format, lint, test)
check: format-check lint test

# Fix all issues (format, then lint with fixes)
fix: format
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fix ./...; \
    else \
        echo "golangci-lint not installed. Run 'just install-tools' to install it."; \
    fi

# Show available commands
default:
    @just --list