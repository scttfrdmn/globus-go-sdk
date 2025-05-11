<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Release Plan: v0.9.17

## Overview

This document outlines the release plan for Globus Go SDK v0.9.17, which focuses on API stability improvements and completing Phase 2 of the API Stability Implementation Plan.

## Key Components

### 1. API Stability Tools

Version 0.9.17 introduces comprehensive API stability tools:

- **API Signature Generation**: `cmd/apigen` generates API signatures for packages
- **API Compatibility Checking**: `cmd/apicompare` compares API signatures between versions
- **Deprecation Tracking**: `cmd/depreport` generates reports of deprecated features
- **Contract Testing**: Core interfaces now have behavioral contract verification

### 2. Enhanced Testing Framework

- **Contract Testing**: Added in `pkg/core/contracts` to verify interface implementations
- **Compatibility Testing**: Added test cases for service clients to verify API stability
- **CI/CD Integration**: Added GitHub Actions workflows for API compatibility verification

### 3. Documentation Improvements

- **API Stability Documentation**: Added detailed documentation for Phase 2 implementation:
  - `API_STABILITY_PHASE2_SUMMARY.md`: Overview of Phase 2 implementation
  - `API_DEPRECATION_SYSTEM.md`: Documentation for the deprecation system
  - `CONTRACT_TESTING.md`: Guide to the contract testing system

### 4. Code Reorganization

- **Command-Line Tools**: Moved utility scripts to proper `cmd/` directories
- **Internal Package**: Created `internal/verification` package for common functionality
- **Debugging Tools**: Restructured debug code for proper package organization

## Testing Plan

Before release, the following tests will be conducted:

1. **Go Tests**: Run comprehensive test suite with `go test ./...`
2. **API Compatibility**: Verify no breaking changes from v0.9.16
   ```
   go run cmd/apigen/main.go -dir ./pkg -v v0.9.16 -o api-v0.9.16.json
   go run cmd/apigen/main.go -dir ./pkg -v v0.9.17 -o api-v0.9.17.json
   go run cmd/apicompare/main.go -old api-v0.9.16.json -new api-v0.9.17.json -level minor
   ```
3. **Contract Tests**: Verify all interface contracts pass
4. **Pre-commit Hooks**: Ensure all pre-commit hooks pass
5. **Downstream Projects**: Test compatibility with downstream projects

## Implementation Status

| Component                           | Status      | Notes                                              |
| ----------------------------------- | ----------- | -------------------------------------------------- |
| API Signature Generation Tools      | Complete    | apigen tool integrated with test system            |
| API Comparison Tools                | Complete    | apicompare tool with semantic versioning support   |
| Deprecation System                  | Complete    | Runtime warnings and reporting tools implemented   |
| Contract Testing Framework          | Complete    | Interface verification for core components         |
| CI/CD Integration                   | Complete    | GitHub Actions workflow for API verification       |
| Code Reorganization                 | Complete    | Improved command-line tools and internal packages  |
| Pre-commit Hook Updates             | Complete    | Switched from golint to staticcheck               |
| Documentation                       | Complete    | Comprehensive documentation for all components     |

## Release Checklist

- [x] Update CHANGELOG.md with release notes
- [x] Add API stability documentation
- [x] Move script tools to proper cmd/ directories
- [x] Create release documentation
- [ ] Run final tests to ensure all pre-commit hooks pass
- [ ] Verify API compatibility with previous version
- [ ] Prepare pull request for merging to main
- [ ] Create git tag for version v0.9.17
- [ ] Update RELEASE_STATUS.md with new version information

## Release Timeline

- **Code Complete**: May 11, 2025
- **Testing Complete**: May 12, 2025
- **Documentation Complete**: May 12, 2025
- **Release Date**: May 13, 2025

## Post-Release Plans

After v0.9.17 is released, development will focus on:

1. **Phase 3 API Stability**: Expanding contract testing to all interfaces
2. **Enhanced Documentation**: Adding examples for new API stability tools
3. **Compute Service Improvements**: Expanding compute functionality
4. **v0.10.0 Planning**: Preparing for the more significant v0.10.0 release