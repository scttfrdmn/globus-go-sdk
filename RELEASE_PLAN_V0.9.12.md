<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Release Plan for v0.9.12

## Overview

This release addresses two important issues:
1. Fixed missing functions in `transport_init.go` that were causing compilation errors
2. Added device authentication flow support to the auth package

## Tasks

### Pre-Release
- [x] Fix issue #13: Missing functions in transport_init.go
- [x] Fix issue #12: Implement device authentication flow
- [x] Update version number to 0.9.12
- [x] Update CHANGELOG.md with new changes
- [x] Add/update documentation for device authentication
- [x] Run tests to verify changes
- [x] Prepare release notes

### Release
- [ ] Create release PR
- [ ] Merge PR to main branch
- [ ] Tag release as v0.9.12
- [ ] Create GitHub release with release notes

### Post-Release
- [ ] Announce release in appropriate channels
- [ ] Monitor for any issues reported after release

## Timeline

Proposed timeline for release:
- Development and testing: 1-2 days
- Release: When all tests pass and documentation is updated
- Post-release monitoring: 1 week

## Implementation Details

### Issue #13 Fix
The fix for the missing functions issue involved updating the `SetConnectionPoolManager` function to accept both `ConnectionPoolProvider` and `interfaces.ConnectionPoolManager` types through an adapter pattern.

### Device Authentication Flow
The device authentication flow implementation includes:
- New models: `DeviceCodeResponse`
- New error handling: `DeviceAuthError` type with specific error codes
- New API methods: 
  - `RequestDeviceCode`
  - `PollDeviceCode`
  - `CompleteDeviceFlow`
- Example implementation in `cmd/examples/device-auth/`
- Documentation in README

This feature enables CLI applications and other non-browser environments to authenticate with Globus services.