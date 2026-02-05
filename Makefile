# Makefile for crumbs
# Build and test automation

# Variables
BINARY_NAME := cupboard
BINARY_DIR := bin
CMD_DIR := ./cmd/cupboard
GOPATH_BIN := $(shell go env GOPATH)/bin

# Build flags
GO_BUILD_FLAGS := -v
ifdef VERBOSE
	GO_TEST_FLAGS := -v
else
	GO_TEST_FLAGS :=
endif

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: build test test-unit test-integration lint clean install help

## build: Compile cupboard binary to bin/
build:
	@mkdir -p $(BINARY_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

## test: Run all tests (unit + integration)
test:
	go test $(GO_TEST_FLAGS) ./...

## test-unit: Run only unit tests (exclude integration)
test-unit:
	go test $(GO_TEST_FLAGS) $(shell go list ./... | grep -v /tests/)

## test-integration: Run only integration tests
test-integration: build
	go test $(GO_TEST_FLAGS) ./tests/...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BINARY_DIR)
	go clean

## install: Install cupboard to GOPATH/bin
install: build
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH_BIN)/$(BINARY_NAME)

## help: Show available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
	@echo ""
	@echo "Options:"
	@echo "  VERBOSE=1    Enable verbose test output"
