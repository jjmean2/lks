.PHONY: all build clean run test format help

APP_NAME=lks
BUILD_DIR=./bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

all: clean build

build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "==> Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: build
	@echo "==> Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

clean:
	@echo "==> Cleaning up..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "==> Running tests..."
	@go test -v ./...

format:
	@echo "==> Formatting code..."
	@go fmt ./...

help:
	@echo "Available commands:"
	@echo "  make          - Same as 'make all' (clean and build)"
	@echo "  make build    - Build the application"
	@echo "  make run      - Build and run the application"
	@echo "  make clean    - Remove build directory"
	@echo "  make test     - Run tests"
	@echo "  make format   - Format Go code using go fmt"
	@echo "  make help     - Show this help message"
