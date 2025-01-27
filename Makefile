# Project variables
APP_NAME := tars
GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")
GO_FMT := gofmt -w
GO_LINT := golangci-lint run
GO_TEST := go test ./...

# Build the binary
.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/tars

# Run the application
.PHONY: run
run:
	go run ./cmd/tars

# Format code
.PHONY: fmt
fmt:
	$(GO_FMT) $(GO_FILES)

# Lint the code
.PHONY: lint
lint:
	$(GO_LINT)

# Run tests
.PHONY: test
test:
	$(GO_TEST)

# Clean up build artifacts
.PHONY: clean
clean:
	rm -rf bin/$(APP_NAME)

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build   - Build the binary"
	@echo "  run     - Run the application"
	@echo "  fmt     - Format the code"
	@echo "  lint    - Lint the code"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Install dependencies"
