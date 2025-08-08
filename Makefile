# Ask Terminal Makefile

BINARY_NAME=askterminal
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=$(HOME)/.config/askterminal

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"
MAIN_FILE=main.go

.PHONY: all build clean test install uninstall dev help

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

# Install the binary to system
install: build
	@echo "Installing to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@mkdir -p $(CONFIG_DIR)
	@echo "Installation complete!"
	@echo "Run: $(BINARY_NAME)"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Binary removed from $(INSTALL_DIR)"
	@read -p "Remove config directory $(CONFIG_DIR)? (y/N): " confirm && \
		if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
			rm -rf $(CONFIG_DIR); \
			echo "Config directory removed"; \
		fi

# Development mode - build and run
dev: build
	@echo "Running in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Multi-platform build complete!"

# Quick install for development
quick-install: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/bin/ || \
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/ || \
	echo "Could not install to standard locations. Try: sudo make install"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Security check (requires gosec)
security:
	@echo "Running security check..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Release preparation
release: clean test build-all
	@echo "Preparing release..."
	@cd $(BUILD_DIR) && \
	for file in $(BINARY_NAME)-*; do \
		if [[ $$file == *.exe ]]; then \
			zip "$${file%.exe}.zip" "$$file"; \
		else \
			tar -czf "$$file.tar.gz" "$$file"; \
		fi; \
	done
	@echo "Release packages created in $(BUILD_DIR)/"

# Show help
help:
	@echo "Ask Terminal - Available commands:"
	@echo ""
	@echo "Building:"
	@echo "  make build      - Build the binary"
	@echo "  make build-all  - Build for multiple platforms"
	@echo "  make clean      - Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  make dev        - Build and run"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter (requires golangci-lint)"
	@echo "  make security   - Run security check (requires gosec)"
	@echo ""
	@echo "Installation:"
	@echo "  make install    - Install to system (/usr/local/bin)"
	@echo "  make uninstall  - Remove from system"
	@echo "  make quick-install - Install to user directory"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps       - Install Go dependencies"
	@echo ""
	@echo "Release:"
	@echo "  make release    - Prepare release packages"
	@echo ""
	@echo "Usage after install:"
	@echo "  askterminal     - Launch the application"
	@echo "  askterminal --help - Show application help"