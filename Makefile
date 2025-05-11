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
STATICCHECK := $(GOBIN)/staticcheck
PRE_COMMIT := $(shell which pre-commit)

.PHONY: all
all: lint staticcheck lint-shell test

.PHONY: setup
setup: $(GOLANGCILINT) $(GOIMPORTS) $(GOCOV) $(GOCOVXML) $(STATICCHECK) setup-pre-commit install-bats
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

$(STATICCHECK):
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest

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

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	$(STATICCHECK) ./...

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
	rm -f cmd/verify-credentials/verify-credentials cmd/verify-credentials/verify-credentials-standalone cmd/verify-credentials/verify-credentials-sdk-api cmd/verify-credentials/main

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

# Verify credentials tool
.PHONY: verify-credentials
verify-credentials:
	@echo "Building verify-credentials tool..."
	$(GO) build -o cmd/verify-credentials/verify-credentials cmd/verify-credentials/main.go

.PHONY: verify-credentials-standalone
verify-credentials-standalone:
	@echo "Building standalone verify-credentials tool..."
	$(GO) build -tags standalone -o cmd/verify-credentials/verify-credentials-standalone cmd/verify-credentials/standalone.go

.PHONY: verify-credentials-sdk-api
verify-credentials-sdk-api:
	@echo "Building SDK API verify-credentials tool..."
	$(GO) build -tags sdk_api -o cmd/verify-credentials/verify-credentials-sdk-api cmd/verify-credentials/verify-credentials-sdk.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  setup              - Install development tools"
	@echo "  setup-pre-commit   - Install pre-commit hooks"
	@echo "  lint               - Run Go linters"
	@echo "  staticcheck        - Run staticcheck linter"
	@echo "  lint-shell         - Run shell script linters"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  test               - Run Go tests"
	@echo "  test-shell         - Run shell script tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-integration   - Run integration tests"
	@echo "  security-scan      - Run security scanning tools"
	@echo "  install-bats       - Install BATS testing framework"
	@echo "  verify-credentials        - Build the verify-credentials SDK tool"
	@echo "  verify-credentials-standalone - Build the standalone verify-credentials tool"
	@echo "  verify-credentials-sdk-api    - Build the SDK API verify-credentials tool"
	@echo "  clean              - Clean build artifacts"
	@echo "  help               - Show this help"