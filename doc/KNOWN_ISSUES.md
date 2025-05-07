<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Known Issues in Globus Go SDK v0.9.6

This document outlines known issues and technical debt in the current version of the Globus Go SDK.

## Build and Compilation Issues

### Import Cycles

While we've made significant progress in resolving import cycles by implementing an interface extraction pattern, there are still some unresolved circular dependencies that need to be addressed:

1. We've added the necessary interface definitions in `pkg/core/interfaces/`
2. We've created adapter types for key components like `ConnectionPool` and `Authorizer`
3. The core interfaces now properly isolate implementation details

Most import cycle issues have been resolved in v0.9.5 and compilation issues fixed in v0.9.6. Remaining areas to monitor include:

- `pkg/services/groups/membership.go` - potential duplicate method declarations need review
- `pkg/services/timers/client.go` - import dependencies should be checked for correctness
- Debug files using old API patterns that need updates for consistency

## Test Infrastructure

The test infrastructure is partially complete but requires additional work:

1. Integration tests are implemented for each service
2. Script for running tests with credentials is working
3. Verification test is implemented to validate credentials
4. Tests are properly structured to run with environment variables

## Documentation

Documentation is nearly complete, with the following areas needing attention:

1. **Integration Testing Guide**: Complete
2. **Release Checklist**: Outlined and ready for use
3. **Import Cycle Resolution**: Documented the approach but needs updates to reflect ongoing work

## Next Steps for v1.0.0 Release

As we approach v1.0.0, the following key tasks need to be completed:

1. **Code Quality and Consistency**: 
   - Review and eliminate any remaining duplicate declarations
   - Ensure consistent naming conventions across all packages
   - Perform comprehensive code review of all service implementations

2. **Enhanced Testing**:
   - Increase test coverage for all packages
   - Add more comprehensive integration tests
   - Improve test reliability and reduce dependencies on external services

3. **Advanced Features**:
   - Implement any remaining advanced features planned for v1.0
   - Ensure all API surfaces have proper implementations
   - Add support for newer Globus API features

4. **Documentation Completeness**:
   - Complete API reference documentation for all services
   - Add more examples and tutorials
   - Create comprehensive usage guides for complex workflows

## Issue Tracking

The primary issues tracked for the v1.0.0 release:

1. **High Priority**:
   - Complete API coverage for all Globus services
   - Resolve any remaining duplicate method declarations
   - Final testing of all error handling paths

2. **Medium Priority**:
   - Improve documentation for advanced usage scenarios
   - Enhance performance of large transfers and operations
   - Add more comprehensive examples

3. **Lower Priority**:
   - Additional convenience methods for common operations
   - Optimization of memory usage for large-scale operations
   - Exploring additional plugin architecture for extensibility

## Timeline

The estimated time to reach v1.0.0 is:

1. High priority issues: 3-4 weeks
2. Medium priority issues: 5-6 weeks
3. Lower priority issues: Ongoing improvements post v1.0.0

Total estimated time to v1.0.0 release: 2-3 months

## Current Status and Recommendations

For developers using the SDK in its current state:

1. All core services (Auth, Transfer, Groups, Flows, Search, Compute) are now functional for typical use cases
2. Integration tests provide good examples of proper service usage
3. The SDK is now stable enough for production use with the following considerations:
   - API surfaces may still evolve until v1.0.0
   - Advanced features might require additional error handling
   - Complex workflows should be thoroughly tested against the Globus services
4. For any issues encountered, please check the GitHub repository for the latest updates or submit an issue