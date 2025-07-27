<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v0.9.0 Release Plan

This document outlines the plan for the v0.9.0 release of the Globus Go SDK.

## Release Timeline

- **Documentation Completion**: Completed
- **Feature Freeze**: May 10, 2025
- **Code Freeze**: May 17, 2025
- **RC1 Release**: May 20, 2025
- **Testing Period**: May 20-27, 2025
- **Final Release**: May 30, 2025

## Major Changes Since v0.8.0

1. **Comprehensive Documentation**:
   - Added detailed documentation for all services
   - Created quick start guides for each service
   - Added comprehensive guides for common use cases
   - Implemented versioned documentation

2. **Enhanced Example Applications**:
   - Data pipeline example
   - Error recovery patterns example
   - Multi-service workflows example
   - Improved existing examples with better documentation

3. **Developer Experience Improvements**:
   - Git hooks for local testing
   - Improved GitHub Actions workflows
   - Consistent error handling patterns
   - Enhanced code documentation

4. **Performance Optimizations**:
   - Memory optimization in Transfer service
   - Connection pooling improvements
   - Rate limit handling enhancements

5. **Bug Fixes**:
   - Fixed string formatting issues
   - Fixed interface implementation bugs
   - Added missing interfaces package
   - Resolved rate limiter issues

## Release Checklist

### Pre-Release Tasks

- [x] Complete service documentation
- [x] Complete quick start guides
- [x] Create comprehensive guides for common use cases
- [x] Add example applications
- [x] Create FAQ section
- [x] Implement versioning for documentation
- [ ] Run comprehensive integration tests
- [ ] Update CHANGELOG.md
- [ ] Update version number in pkg/core/version.go
- [ ] Update go.mod file
- [ ] Review API compatibility

### Release Process

1. Create release branch:
   ```bash
   git checkout -b release/v0.9.0
   ```

2. Update version information:
   ```bash
   # Edit pkg/core/version.go to set version to "0.9.0"
   # Edit CHANGELOG.md to add v0.9.0 section
   ```

3. Run final tests:
   ```bash
   make test
   make integration-test
   ```

4. Create release candidate tag:
   ```bash
   git tag -a v0.9.0-rc1 -m "v0.9.0 Release Candidate 1"
   git push origin v0.9.0-rc1
   ```

5. Create release documentation:
   ```bash
   # Will be triggered by CI on tag
   ```

6. Announce release candidate for testing

7. After testing period, create final release:
   ```bash
   git tag -a v0.9.0 -m "v0.9.0 Release"
   git push origin v0.9.0
   ```

8. Create GitHub release with changelog

9. Merge to main:
   ```bash
   git checkout main
   git merge release/v0.9.0
   git push origin main
   ```

## Documentation Versioning

Documentation for v0.9.0 has been set up with versioning support. The documentation is available at:

- Latest (main branch): https://docs.globus.org/developer-tools/go-sdk/latest/
- v0.9.0: https://docs.globus.org/developer-tools/go-sdk/v0.9.0/
- v0.8.0: https://docs.globus.org/developer-tools/go-sdk/v0.8.0/

The versioning system allows users to view documentation for specific SDK versions and easily switch between them.

## Known Issues and Limitations

- Search service advanced queries functionality has some edge cases documented in the `FIXES.md` file
- Flows service pagination API may change in a future version
- Some integration tests are disabled pending endpoint access configurations

## Future Plans

Items considered but deferred to future releases:

1. **Globus Connect Server API Support**:
   - Management interfaces for Globus Connect Server
   - Endpoint creation and configuration APIs

2. **Enhanced Rate Limiting**:
   - Dynamic backoff strategies
   - Service-specific rate limit policies

3. **Advanced Security Features**:
   - Additional MFA mechanisms
   - Enhanced scope management

These features will be considered for the v0.10.0 or later releases.