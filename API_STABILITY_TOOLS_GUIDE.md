# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# API Stability Tools Guide

This document provides guidance on using the API stability tools developed as part of Phase 2 of the API Stability Implementation Plan for the Globus Go SDK.

## Table of Contents

- [Overview](#overview)
- [Tool Categories](#tool-categories)
- [API Compatibility Tools](#api-compatibility-tools)
  - [API Generator (`apigen`)](#api-generator-apigen)
  - [API Comparator (`apicompare`)](#api-comparator-apicompare)
  - [Verification Script](#verification-script)
- [Deprecation System](#deprecation-system)
  - [Runtime Deprecation Warnings](#runtime-deprecation-warnings)
  - [Deprecation Report Generator](#deprecation-report-generator)
- [Contract Testing](#contract-testing)
  - [Interface Contracts](#interface-contracts)
  - [Behavior Verification](#behavior-verification)
- [Continuous Integration](#continuous-integration)
  - [API Stability Workflow](#api-stability-workflow)
  - [Coverage Reporting](#coverage-reporting)
- [Best Practices](#best-practices)

## Overview

The API stability tools ensure that changes to the Globus Go SDK maintain compatibility guarantees based on semantic versioning principles. These tools help developers identify potential breaking changes, manage deprecations, and verify that implementations conform to interface contracts.

## Tool Categories

The API stability system consists of three main categories of tools:

1. **API Compatibility Tools** - Extract and compare API signatures between versions
2. **Deprecation System** - Mark, track, and report deprecated features
3. **Contract Testing** - Verify that implementations conform to interface contracts

## API Compatibility Tools

### API Generator (`apigen`)

Located at `cmd/apigen/main.go`, this tool extracts API signatures from Go source code.

**Usage:**

```bash
go run ./cmd/apigen/main.go -dir ./pkg -version "v0.9.16" -output api-signatures.json
```

**Arguments:**

- `-dir` - Directory to scan for Go files (default: ".")
- `-version` - Version of the API being scanned (default: "current")
- `-output` - Output file for API signatures (default: "api-signatures.json")

**Output:**

The tool generates a JSON file containing signatures for all exported APIs, including:
- Functions and methods
- Types (structs, interfaces, etc.)
- Constants and variables

### API Comparator (`apicompare`)

Located at `cmd/apicompare/main.go`, this tool compares API signatures between versions to detect breaking changes.

**Usage:**

```bash
go run ./cmd/apicompare/main.go -old old-api.json -new new-api.json -level minor -output comparison.json
```

**Arguments:**

- `-old` - Previous version API signatures file
- `-new` - Current version API signatures file
- `-level` - Compatibility level: patch, minor, or major (default: "minor")
- `-output` - Output file for comparison results (optional)

**Compatibility Levels:**

- `patch` - No API changes allowed
- `minor` - Additions allowed, no removals or breaking changes
- `major` - Any changes allowed, but breaking changes are reported

### Verification Script

Located at `scripts/verify_api_compatibility.sh`, this script automates the process of comparing API signatures between versions.

**Usage:**

```bash
./scripts/verify_api_compatibility.sh --prev-version v0.9.15 --current-version v0.9.16 --level minor
```

**Arguments:**

- `--prev-version` - Previous version (default: latest release tag)
- `--current-version` - Current version (default: current commit)
- `--level` - Compatibility level: patch, minor, or major (default: "minor")
- `--output-dir` - Output directory for results (default: "./api-compatibility-results")

**Output:**

The script generates a comprehensive report in Markdown format that includes:
- Summary of API changes
- Detailed list of additions, removals, and changes
- Breaking changes (if any)
- Deprecation analysis

## Deprecation System

### Runtime Deprecation Warnings

Located in `pkg/core/deprecation`, this package provides utilities for marking and logging deprecation warnings at runtime.

**Example Usage:**

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/deprecation"

func OldFunction() {
    // Log a deprecation warning
    deprecation.LogWarning(
        ctx,               // Context
        "OldFunction",     // Feature name
        "v0.9.0",          // Deprecated in version
        "v1.0.0",          // Will be removed in version
        "Use NewFunction instead" // Usage guidance
    )
    
    // Implement the function...
}
```

**Features:**

- Configurable warning frequency
- Integration with SDK logging system
- Version-based deprecation tracking

### Deprecation Report Generator

Located at `cmd/depreport/main.go`, this tool scans the codebase for deprecated features and generates a report.

**Usage:**

```bash
go run ./cmd/depreport/main.go -dir ./pkg -o DEPRECATED_FEATURES.md
```

**Arguments:**

- `-dir` - Source directory to scan (default: ".")
- `-o` - Output file for the report (default: stdout)

**Output:**

The tool generates a Markdown report that includes:
- Features with planned removal dates
- Features without planned removal dates
- Usage guidance for deprecated features
- Summary statistics

## Contract Testing

### Interface Contracts

Located in `pkg/core/contracts`, this package provides utilities for verifying that implementations conform to interface contracts.

**Example Usage:**

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/contracts"

func TestCustomTransport(t *testing.T) {
    transport := NewCustomTransport()
    contracts.VerifyTransportContract(t, transport)
}
```

### Behavior Verification

The contract tests verify not just type conformance but also behavioral conformance:

- **Transport Contracts** - Verify HTTP methods, error handling, and context cancellation
- **Client Contracts** - Verify base client behaviors
- **Config Contracts** - Verify configuration behaviors

## Continuous Integration

### API Stability Workflow

The API stability workflow is defined in `.github/workflows/api-stability.yml` and runs on:
- Pushes to main branch
- Pull requests targeting main branch
- Manual triggers with configurable compatibility level

**Jobs:**

1. **API Compatibility Check** - Extracts and compares API signatures
2. **Code Coverage** - Generates coverage reports and checks against thresholds
3. **Compatibility Testing** - Runs comprehensive compatibility tests

### Coverage Reporting

The code coverage workflow is defined in `.github/workflows/codecov.yml` and provides:
- Package-level coverage metrics
- Threshold-based warnings
- HTML and Markdown reports
- PR comments with coverage information

## Best Practices

When making changes to the SDK:

1. **Check Compatibility Before Releases**
   ```bash
   ./scripts/verify_api_compatibility.sh --level minor
   ```

2. **Mark Deprecated Features Properly**
   ```go
   // Deprecated: Use NewFunction instead. Will be removed in v1.0.0.
   func OldFunction() {
       deprecation.LogWarning(ctx, "OldFunction", "v0.9.0", "v1.0.0", "Use NewFunction instead")
       // Implementation...
   }
   ```

3. **Write Contract Tests for New Interfaces**
   ```go
   // In pkg/core/contracts/your_interface_contract.go
   func VerifyYourInterfaceContract(t *testing.T, impl interfaces.YourInterface) {
       // Verify behaviors...
   }
   ```

4. **Run Verification Before Merging**
   ```bash
   # Run the API stability workflow locally
   go test -v ./pkg/core/contracts/...
   go run ./cmd/depreport/main.go
   ./scripts/verify_api_compatibility.sh
   ```

5. **Follow Semantic Versioning**
   - Patch (0.9.x): Bug fixes, no API changes
   - Minor (0.x.0): New features, no breaking changes
   - Major (x.0.0): Breaking changes allowed