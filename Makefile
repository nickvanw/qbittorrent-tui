.PHONY: help test test-unit test-integration test-coverage clean lint fmt build run validate

# Default target shows help
help:
	@echo "qBittorrent TUI - Development Commands"
	@echo ""
	@echo "Testing:"
	@echo "  make test              Run all tests"
	@echo "  make test-unit         Run unit tests only"  
	@echo "  make test-integration  Run integration tests (requires Docker)"
	@echo "  make test-coverage     Generate coverage report"
	@echo ""
	@echo "Development:"
	@echo "  make build             Build the application"
	@echo "  make run               Build and run the application"
	@echo "  make fmt               Format code"
	@echo "  make lint              Run linters (requires golangci-lint)"
	@echo "  make clean             Clean all artifacts and test containers"
	@echo ""
	@echo "Setup:"
	@echo "  make deps              Download and tidy dependencies"
	@echo "  make install-tools     Install development tools"

# Build the application
build:
	@echo "🔨 Building qbt-tui..."
	@go build -o bin/qbt-tui ./cmd/qbt-tui

# Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	@go test -short -v ./...

# Run integration tests with proper setup
test-integration: docker-up
	@echo "🧪 Running integration tests..."
	@QBT_TEST_PASSWORD="testpass123" go test -v -tags=integration ./internal/api
	@$(MAKE) docker-down

# Run all tests
test: test-unit test-integration

# Generate test coverage
test-coverage:
	@echo "📊 Generating test coverage..."
	@go test -short -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📄 Coverage report saved to coverage.html"

# Docker commands
docker-up: docker-down
	@echo "🐳 Starting test containers..."
	@docker compose -f docker-compose.test.yml up -d --wait

docker-down:
	@docker compose -f docker-compose.test.yml down -v 2>/dev/null || true

# Clean everything
clean: docker-down
	@echo "🧹 Cleaning up..."
	@rm -f bin/qbt-tui coverage.out coverage.html
	@rm -rf testdata/fresh-config testdata/setup-config
	@find testdata -name "*.log" -delete 2>/dev/null || true
	@go clean -testcache

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@go mod tidy

# Run linters
lint:
	@echo "🔍 Running linters..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "❌ golangci-lint not found. Run 'make install-tools' first."; \
		exit 1; \
	fi
	@golangci-lint run

# Install development tools
install-tools:
	@echo "📦 Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Tools installed"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run the application
run: build
	@echo "🚀 Running qbt-tui..."
	@./bin/qbt-tui

# Run the application with test configuration
dev: build
	@echo "🚀 Running qbt-tui in dev mode..."
	@QBT_SERVER_URL="http://localhost:8181" \
	 QBT_SERVER_USERNAME="admin" \
	 QBT_SERVER_PASSWORD="testpass123" \
	 ./bin/qbt-tui

# Run validation suite
validate: clean lint test-coverage
	@echo "✅ All validations passed!"

# Quick check - format, vet, and unit tests
check: fmt
	@echo "⚡ Running quick checks..."
	@go vet ./...
	@go test -short ./...