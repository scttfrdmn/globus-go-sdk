<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK Release Status

## Current Release: v0.9.16

The Globus Go SDK is currently at version v0.9.16, released on May 10, 2025.

### Latest Release Details

Version v0.9.16 includes:
- Complete API stability indicators for all packages
- Standardized release checklist and process
- Per-package code coverage targets
- Enhanced documentation with compatibility guarantees
- Implementation of API compatibility verification tools
- Runtime deprecation warning system

### Recent Releases

- **v0.9.16** (2025-05-10): API stability improvements and tools
- **v0.9.15** (2025-05-08): Properly tagged release for connection pool functions fix
- **v0.9.14** (2025-05-07): Added comprehensive test coverage for connection pool functions
- **v0.9.13** (2025-05-07): Restored missing connection pool functions
- **v0.9.12** (2025-05-07): Documentation update for 0.9.12
- **v0.9.11** (2025-05-07): Bug fixes for GitHub Actions and client initialization
- **v0.9.10** (2025-05-07): Fixed connection pool initialization

## Upcoming Releases

### v0.9.17 (In Development)

This release will continue API stability improvements:
- Contract testing for core interfaces
- CI integration for API compatibility verification
- Improved test coverage for core packages
- Integration of deprecation tracking into release process

### v0.10.0 (Planned)

Planned features and improvements:
- Complete contract testing for all interfaces
- Comprehensive automated API compatibility verification
- Enhanced authentication mechanisms with improved MFA support
- Expanded compute service capabilities
- Improved performance for transfer operations
- Consolidated error handling across all services

### v1.0.0 (Long-term Goal)

Our road to v1.0.0 includes:
- Complete API stability throughout the SDK
- Comprehensive documentation and examples
- Full test coverage for all packages
- Formal API review process
- Migration guides for any breaking changes

## Current Status

All critical bugs have been fixed, and the SDK is stable for production use. 

We have made significant progress on API stability:
- **Phase 1 Complete**: All packages now have stability indicators and documentation
- **Phase 2 In Progress**: API compatibility tools and deprecation system are fully implemented
- **Next Steps**: Implementing contract testing and CI integration for API compatibility verification

The SDK now provides clear compatibility guarantees and tools to help maintain API stability as we progress toward v1.0.0.

## Release Process

We now follow a standardized release process as documented in `RELEASE_CHECKLIST.md`:

1. All changes go through code review via pull requests
2. Comprehensive testing including:
   - Unit tests (with code coverage targets)
   - Integration tests
   - API compatibility verification
   - Downstream project compatibility verification
3. Documentation updates, including:
   - CHANGELOG.md updates
   - API stability documentation
   - Deprecation notices
4. API verification checks:
   - Run API compatibility tools
   - Generate deprecation reports
   - Verify semantic versioning compliance
5. Proper Git tagging and GitHub releases

## How to Upgrade

To update to the latest version:
```
go get github.com/scttfrdmn/globus-go-sdk@v0.9.16
```

### API Stability Tools

Developers working on the SDK can now use our API stability tools:

1. **API Compatibility Verification**:
   ```
   go run cmd/apigen/main.go -dir ./pkg -v v0.9.16 -o api-v0.9.16.json
   go run cmd/apicompare/main.go -old api-v0.9.15.json -new api-v0.9.16.json -level minor
   ```

2. **Deprecation Reporting**:
   ```
   go run cmd/depreport/main.go -o DEPRECATED_FEATURES.md
   ```

For complete information on API stability, refer to our documentation:
- `API_STABILITY_PHASE1_SUMMARY.md` - Package stability indicators
- `API_STABILITY_PHASE2_SUMMARY.md` - API compatibility tools
- `API_DEPRECATION_SYSTEM.md` - Deprecation system
- `RELEASE_CHECKLIST.md` - Release process