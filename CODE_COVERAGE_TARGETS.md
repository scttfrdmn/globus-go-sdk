<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Code Coverage Targets for Globus Go SDK

This document establishes code coverage targets for the Globus Go SDK as part of the overall API Stability Implementation Plan. Comprehensive test coverage is essential for ensuring API stability, detecting regressions, and providing confidence in the SDK's reliability.

## Coverage Goals by Release

| Release | Overall Target | Core Packages Target | Service Packages Target |
|---------|---------------|---------------------|-------------------------|
| v0.9.16 | 60%           | 70%                 | 60%                     |
| v0.10.0 | 70%           | 80%                 | 70%                     |
| v1.0.0  | 80%           | 90%                 | 80%                     |

## Current Coverage Status (as of API Stability Implementation)

Based on the analysis performed on May 9, 2025, the current coverage levels are:

| Package                          | Current Coverage | Target (v0.9.16) | Status |
|----------------------------------|----------------|-------------------|--------|
| **Overall**                      | 50.9%           | 60%               | Below  |
| **Core Packages**                |                 |                   |        |
| pkg/core                         | 6.4%            | 70%               | Well Below |
| pkg/core/auth                    | 81.8%           | 70%               | Exceeds |
| pkg/core/authorizers             | 77.4%           | 70%               | Exceeds |
| pkg/core/config                  | 30.4%           | 70%               | Well Below |
| pkg/core/http                    | 57.4%           | 70%               | Below  |
| pkg/core/logging                 | 62.9%           | 70%               | Below  |
| pkg/core/ratelimit               | 84.1%           | 70%               | Exceeds |
| pkg/core/transport              | 17.6%           | 70%               | Well Below |
| **Service Packages**            |                 |                   |        |
| pkg/services/auth                | 57.7%           | 60%               | Below  |
| pkg/services/compute             | 47.0%           | 60%               | Below  |
| pkg/services/flows               | 57.3%           | 60%               | Below  |
| pkg/services/groups              | 57.0%           | 60%               | Below  |
| pkg/services/search              | 79.5%           | 60%               | Exceeds |
| pkg/services/timers              | 37.1%           | 60%               | Well Below |
| pkg/services/tokens              | 85.1%           | 60%               | Exceeds |
| pkg/services/transfer            | 48.1%           | 60%               | Below  |
| pkg/metrics                      | 61.0%           | 60%               | Exceeds |

## Package-Specific Coverage Requirements

### Critical Packages

These packages form the foundation of the SDK and require the highest level of testing:

1. **pkg/core** - Base client implementation (Target: 80% by v0.9.16)
   - 100% coverage for error handling
   - 100% coverage for connection pool management
   - 90% coverage for HTTP request/response processing

2. **pkg/core/auth** - Authentication and authorization (Target: 85% by v0.9.16)
   - 100% coverage for token refresh logic
   - 100% coverage for authentication error handling
   - 90% coverage for storage implementations

3. **pkg/services/auth** - Globus Auth API client (Target: 70% by v0.9.16)
   - 100% coverage for OAuth2 flow implementation
   - 90% coverage for token operations (introspect, revoke)
   - 80% coverage for MFA-related functionality

### Stability-Critical Packages

These packages are central to the API stability effort and require thorough testing:

1. **pkg/services/tokens** - Token management (Target: 85% by v0.9.16)
   - 100% coverage for token storage operations
   - 100% coverage for token refreshing logic
   - 90% coverage for background refresh functionality

2. **pkg/core/transport** - HTTP transport layer (Target: 70% by v0.9.16)
   - 90% coverage for connection management
   - 80% coverage for request/response handling
   - 70% coverage for debugging and logging

### Service Packages

Service-specific packages have the following targets for v0.9.16:

1. **pkg/services/transfer** (Target: 70%)
   - 90% coverage for core file operations
   - 80% coverage for task management
   - 70% coverage for recursive transfers
   - 60% coverage for resumable transfers (experimental)

2. **pkg/services/search** (Target: 80%)
   - 90% coverage for search operations
   - 80% coverage for index management
   - 70% coverage for advanced query features

3. **pkg/services/flows** (Target: 65%)
   - 80% coverage for flow management
   - 70% coverage for run operations
   - 60% coverage for batch operations

4. **pkg/services/compute** (Target: 65%)
   - 80% coverage for function management
   - 70% coverage for task execution
   - 60% coverage for advanced features (containers, environments)

5. **pkg/services/groups** (Target: 65%)
   - 80% coverage for group operations
   - 70% coverage for membership operations
   - 60% coverage for role management

6. **pkg/services/timers** (Target: 60%)
   - 80% coverage for timer management
   - 70% coverage for basic timer operations
   - 50% coverage for helper methods

## Implementation Plan

To achieve these coverage targets, the following steps will be implemented:

1. **Phase 1: Critical Path Coverage (v0.9.16)**
   - Focus on core client interfaces and implementations
   - Add tests for connection pooling and HTTP transport
   - Ensure high coverage of error handling and edge cases

2. **Phase 2: Service Coverage Enhancement (v0.10.0)**
   - Expand test coverage for all service clients
   - Add integration tests for service interactions
   - Ensure coverage of all public API methods

3. **Phase 3: Comprehensive Coverage (v1.0.0)**
   - Achieve full coverage targets for stable API components
   - Add contract tests for interface implementations
   - Ensure edge case and error handling coverage

## Testing Strategy

The following testing approaches will be used to achieve these coverage targets:

1. **Unit Testing**
   - Pure function testing with mocked dependencies
   - Interface implementation verification
   - Error handling verification

2. **Integration Testing**
   - Service client testing with mock servers
   - Cross-service interaction testing
   - Configuration validation

3. **Contract Testing**
   - Formal verification of interface contracts
   - Behavioral consistency testing
   - Edge case validation

4. **Regression Testing**
   - Automated test runs for all changes
   - Backward compatibility verification
   - API stability validation

## CI Integration

Coverage tracking will be integrated into the CI pipeline to:

1. Report coverage metrics for each PR
2. Block merges that significantly reduce coverage
3. Track coverage trends over time
4. Generate coverage reports for review

## Exemptions

Some code may be exempted from coverage targets for practical reasons:

1. **Examples and utilities**
   - Example code is validated through compilation but not unit tested
   - Command-line utilities focus on integration testing rather than unit testing

2. **Generated code**
   - Generated or boilerplate code may have lower coverage targets

3. **Platform-specific code**
   - Platform-specific functionality may have exemptions when testing on all platforms is impractical

All exemptions must be documented and justified.

## Monitoring and Reporting

Coverage will be monitored through:

1. Automated coverage reporting in CI
2. Weekly coverage trend analysis
3. Coverage blockers for critical packages
4. Periodic comprehensive coverage reviews

## Conclusion

These coverage targets establish a foundation for ensuring the reliability and stability of the Globus Go SDK. By progressively increasing coverage and focusing on critical components, we can build user confidence in the SDK's quality and stability.

_Version: 1.0_
_Last Updated: May 9, 2025_