<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK Release Status

## Current Release: v0.9.15

The Globus Go SDK is currently at version v0.9.15, released on May 8, 2025.

### Latest Release Details

Version v0.9.15 includes:
- Fixed connection pool functions that were missing in previous releases
- Comprehensive testing and validation of the fix
- Properly tagged release with correct Git tags

### Recent Releases

- **v0.9.15** (2025-05-08): Properly tagged release for connection pool functions fix
- **v0.9.14** (2025-05-07): Added comprehensive test coverage for connection pool functions
- **v0.9.13** (2025-05-07): Restored missing connection pool functions
- **v0.9.12** (2025-05-07): Documentation update for 0.9.12
- **v0.9.11** (2025-05-07): Bug fixes for GitHub Actions and client initialization
- **v0.9.10** (2025-05-07): Fixed connection pool initialization
- **v0.9.9** (2025-05-07): Added API compatibility testing

## Upcoming Releases

### v0.10.0 (Planned)

Planned features:
- Enhanced authentication mechanisms
- Expanded compute service capabilities
- Improved performance for transfer operations
- Enhanced documentation and examples

## Current Status

All critical bugs have been fixed, and the SDK is stable for production use. The most recent bug fix (issue #13) has been thoroughly validated and released as v0.9.15.

## Release Process

1. All changes go through code review via pull requests
2. Comprehensive testing including:
   - Unit tests
   - Integration tests
   - Downstream project compatibility verification
3. Documentation updates
4. Proper Git tagging and GitHub releases

## How to Upgrade

To update to the latest version:
```
go get github.com/scttfrdmn/globus-go-sdk@v0.9.15
```