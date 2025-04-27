## SPDX-License-Identifier: Apache-2.0
## Copyright (c) 2025 Scott Friedman and Project Contributors

SHELL := /bin/bash
GO := go
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
GOLANGCILINT := $(GOBIN)/golangci-lint
GOIMPORTS := $(GOBIN)/goimports
GOCOV := $(GOBIN)/gocov
GOCOVXML := $(GOBIN)/gocov-xml
PRE_COMMIT := $(shell which pre-commit)

.PHONY: all
all: lint lint-shell test

.PHONY: setup
setup: $(GOLANGCILINT) $(GOIMPORTS) $(GOCOV) $(GOCOVXML) setup-pre-commit install-bats
	$(GO) mod download

.PHONY: setup-pre-commit
setup-pre-commit:
	@if [ -z "$(PRE_COMMIT)" ]; then \
		echo "Installing pre-commit..."; \
		pip install pre-commit; \
	fi
	pre-commit install

$(GOLANGCILINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) latest

$(GOIMPORTS):
	$(GO) install golang.org/x/tools/cmd/goimports@latest

$(GOCOV):
	$(GO) install github.com/axw/gocov/gocov@latest

$(GOCOVXML):
	$(GO) install github.com/AlekSi/gocov-xml@latest

.PHONY: lint
lint: $(GOLANGCILINT)
	$(GOLANGCILINT) run --config .golangci.yml

.PHONY: fmt
fmt: $(GOIMPORTS)
	$(GOIMPORTS) -w .
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test -race ./...

.PHONY: test-coverage
test-coverage: $(GOCOV) $(GOCOVXML)
	$(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCOV) convert coverage.txt > coverage.json
	$(GOCOVXML) < coverage.json > coverage.xml
	$(GO) tool cover -html=coverage.txt -o coverage.html

.PHONY: test-integration
test-integration:
	$(GO) test -v -tags=integration ./...

.PHONY: clean
clean:
	$(GO) clean
	rm -f coverage.txt coverage.json coverage.xml coverage.html

# Shell script linting and testing
.PHONY: lint-shell
lint-shell:
	@echo "Linting shell scripts..."
	@./scripts/lint_shell_scripts.sh

.PHONY: install-bats
install-bats:
	@echo "Installing BATS testing framework..."
	@./scripts/install_bats.sh

.PHONY: test-shell
test-shell: install-bats
	@echo "Running shell script tests..."
	@./scripts/run_shell_tests.sh

# Security scanning
.PHONY: security-scan
security-scan:
	@echo "Running security scan..."
	@./scripts/run_security_scan.sh

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  setup              - Install development tools"
	@echo "  setup-pre-commit   - Install pre-commit hooks"
	@echo "  lint               - Run Go linters"
	@echo "  lint-shell         - Run shell script linters"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  test               - Run Go tests"
	@echo "  test-shell         - Run shell script tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-integration   - Run integration tests"
	@echo "  security-scan      - Run security scanning tools"
	@echo "  install-bats       - Install BATS testing framework"
	@echo "  clean              - Clean build artifacts"
	@echo "  help               - Show this help"