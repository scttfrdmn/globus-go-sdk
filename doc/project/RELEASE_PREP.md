<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Release Preparation Guide

This document outlines the steps needed to prepare the Globus Go SDK for the v0.2.0 release.

## 1. Remaining Issues to Fix

### 1.1 Build Issues

There are several import cycle and method reference issues that need to be resolved:

- Import cycles in:
  - `pkg/core/pool` - Imports `interfaces` but doesn't use it
  - `pkg/services/transfer` - Has circular import with `benchmark`
  - `pkg/benchmark` - Incorrectly imports `transfer` causing cycles

- Method references that need fixing:
  - `authorizer.GetToken` in integration tests
  - `client.Client.BuildRequest` reference issues
  - `authorizer.AuthorizeRequest` missing methods
  - Undefined fields in transfer package

### 1.2 Test Failures

Several tests are failing:

- `TestBatchGetFlows` - Expects error types that don't match
- `TestBatchCancelRuns` - Expected error types don't match
- Integration tests are being skipped due to missing environment variables

## 2. Release Checklist

### 2.1 Documentation

- [x] Update CHANGELOG.md with all new features and fixes
- [x] Update README.md with latest features and status
- [x] Update version number in `pkg/globus.go`
- [ ] Ensure all public APIs have proper documentation

### 2.2 Testing

- [ ] Fix failing tests
- [ ] Run integration tests with proper credentials
- [ ] Run performance benchmarks
- [ ] Check for memory leaks in large transfers

### 2.3 Code Quality

- [ ] Run linters and address any issues
- [ ] Verify documentation for public APIs
- [ ] Check for any remaining TODO or FIXME comments
- [ ] Ensure error messages are consistent and helpful

### 2.4 Security

- [ ] Review token handling for security issues
- [ ] Ensure no sensitive information is logged
- [ ] Check for proper authentication flows

### 2.5 Final Verification

- [x] Verify credentials with the verify-credentials tool
- [ ] Test with all supported service APIs
- [ ] Run examples to ensure they work as expected
- [ ] Check for any compatibility issues with Go versions

## 3. Release Process

Once all the issues above have been addressed, follow these steps to create the release:

1. Merge any outstanding PRs
2. Ensure the main branch builds and passes all tests
3. Create a tag for the release:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```
4. Create a GitHub release:
   - Title: "Globus Go SDK v0.2.0"
   - Description: Copy from the CHANGELOG.md entry
   - Target: The tag created in step 3
5. Publish to the Go module proxy:
   ```bash
   GOPROXY=proxy.golang.org go list -m github.com/scttfrdmn/globus-go-sdk@v0.2.0
   ```
6. Announce the release in relevant channels

## 4. Post-Release

After the release, these tasks should be completed:

1. Start planning for the next release
2. Update the roadmap document
3. Create issues for any known problems that weren't fixed in this release
4. Gather feedback from early users
5. Address any critical bugs that are discovered