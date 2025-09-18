<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing added yet

### Changed
- Nothing changed yet

### Deprecated
- Nothing deprecated yet

### Removed
- Nothing removed yet

### Fixed
- Nothing fixed yet

### Security
- Nothing security-related yet

## [3.63.0-1] - 2025-09-18

### Changed
- **Python SDK v3.63.0 synchronization**
  - **Method Rename**: `SetSubscriptionAdminVerifiedID` renamed to `SetSubscriptionAdminVerified` in Groups client
  - Updated method naming to match upstream Python SDK v3.63.0 naming convention
  - Updated all tests to use new method name for consistency

### Deprecated
- **Groups Client Method**
  - `SetSubscriptionAdminVerifiedID()` method is now deprecated in favor of `SetSubscriptionAdminVerified()`
  - Deprecated method remains functional and delegates to the new method for backward compatibility
  - Will be removed in a future major version

### Technical Details
- **Version**: Updated SDK version constant to 3.63.0
- **Backward Compatibility**: Full backward compatibility maintained through deprecated method delegation
- **Testing**: All existing tests updated to use new method names while maintaining deprecated method testing
- **Python SDK Parity**: Maintains synchronization with upstream Globus Python SDK v3.63.0

This release maintains our commitment to tracking upstream Python SDK releases while ensuring backward compatibility for existing users.

## [3.62.0-3] - 2025-08-08

### Added
- **Comprehensive Testing Enhancement Infrastructure**
  - Complete Phase 1 & Phase 2 testing infrastructure following Python SDK patterns
  - **71 comprehensive tests** across unit, functional, and integration test suites
  - **Metadata-driven testing system** with JSON test scenarios for enhanced test organization
  - **Shared testing utilities** in `pkg/testhelpers/` for consistent test patterns across services
  - **Enhanced error scenario testing** with systematic HTTP error code coverage (4xx, 5xx responses)
  - **Workflow-based functional testing** for end-to-end user journey validation
  - **Python SDK parity method testing** covering all 9 parity methods with comprehensive validation

- **New Testing Infrastructure Files**
  - `pkg/testhelpers/fixtures.go` - Shared testing infrastructure and utilities
  - `pkg/services/groups/unit/*` - Comprehensive unit testing suite with 8 test files
  - `pkg/services/groups/functional/*` - Workflow-based functional tests
  - `pkg/services/groups/integration/*` - End-to-end integration testing
  - `TESTING_ENHANCEMENT_PLAN.md` - Complete testing strategy and implementation roadmap

- **Enhanced Test Coverage**
  - **Error Scenario Testing**: Systematic testing of all HTTP error conditions with JSON error response parsing
  - **Subscription Method Testing**: Complete test coverage for v3.62.0 subscription functionality
  - **Python SDK Parity Validation**: Integration tests covering all 9 Python SDK parity methods
  - **Metadata-Driven Test Scenarios**: 15+ structured test cases with variable substitution and templates
  - **Network Error Handling**: Timeout, connection failure, and network-level error scenario testing

### Changed
- **Improved Error Handling in Groups Client**
  - Enhanced `doRequestLowLevel` method to properly handle `core.Error` types
  - Fixed JSON error response parsing with proper `GlobusError` creation
  - Improved error propagation from core HTTP client to service-specific clients

- **Enhanced Test Organization**
  - Restructured groups service testing into unit, functional, and integration test suites
  - Implemented consistent test patterns following upstream Python SDK approaches
  - Added emoji-based test logging for better test output readability

### Fixed
- **Network Timeout Test Stability**
  - Fixed hanging network timeout tests by replacing infinite blocking with controlled timeouts
  - Improved test reliability and reduced CI/CD execution time

### Technical Details
- **Files Modified**: 4 files enhanced (422 lines added)
- **New Files**: 13 new test infrastructure files (4,000+ lines)
- **Test Suite Coverage**: Unit tests, functional workflows, integration scenarios, error handling, and model validation
- **Python SDK Parity**: Complete testing of all subscription management, policy configuration, identity preferences, and membership field methods
- **Infrastructure Improvements**: Mock server enhancements, variable substitution system, dependency resolution, and test case generation utilities

This release significantly enhances the SDK's testing infrastructure to ensure robust quality assurance and maintainability, following proven patterns from the upstream Python SDK while maintaining full backward compatibility.

## [3.62.0-2] - 2025-01-27

### Fixed
- **Version consistency fix**
  - Corrected Version constant in `pkg/core/version.go` from "3.60.0" to "3.62.0"
  - Ensures consistency with v3.62.0 release tags and numbering
  - Addresses oversight from v3.62.0-1 release process

## [3.62.0-1] - 2025-01-27

### Added
- **Python SDK v3.62.0 feature synchronization**
  - Maintained synchronized versioning with Python SDK v3.62.0
  - Groups service subscription_id support
  - SetSubscriptionAdminVerifiedID() method for setting group subscription IDs (admin-only)
  - GetGroupSubscription() method for retrieving group subscription information
  - GroupSubscription type for handling subscription data

### Changed
- **Version synchronization**
  - Updated SDK version to 3.62.0 to match Python SDK v3.62.0
  - All changes maintain backward compatibility with existing v3.61.x code

## [3.61.0-1] - 2025-01-27

### Added
- **Python SDK v3.61.0 feature synchronization**
  - Maintained synchronized versioning with Python SDK v3.61.0
  - Added comprehensive deprecation warnings for legacy functionality

### Deprecated
- **Globus Connect Server v4 support**
  - SetupGridFTPV4Server() method deprecated
  - ConfigureGCSV4Endpoint() method deprecated  
  - GetGCSV4ServerList() method deprecated
  - GCSV4Config, GCSV4ServerList, GCSV4Server types deprecated
  - All GCS v4 methods will emit deprecation warnings when used
- **ComputeClient alias deprecated**
  - ComputeClient type alias deprecated in favor of compute.Client
  - NewComputeClientV2() function deprecated in favor of compute.NewClient()
  - Users encouraged to use compute.Client directly

### Changed
- Updated SDK version to v3.61.0 to maintain Python SDK synchronization

## [3.60.0-1] - 2025-01-27

### Added
- **Version synchronization with Python SDK**
  - Updated versioning to hybrid format `[PYTHON_SDK_VERSION]-[GO_SDK_BUILD]` (v3.60.0-1)
  - Implemented synchronized versioning with Python SDK v3.60.0  
  - Added comprehensive versioning strategy documentation (VERSIONING_STRATEGY.md)
  - Updated module path to github.com/scttfrdmn/globus-go-sdk/v3
- **Globus Auth Requirements Error (GARE) support**
  - Added GlobusAuthRequirementsError type for handling dependent consent errors
  - Implemented recognition of `dependent_consent_required` errors from Auth API
  - Added support for authorization parameters containing dependent scopes
  - Added helper functions: IsGlobusAuthRequirementsError(), IsConsentRequired(), IsDependentConsentRequired()
- **Unified error handling system**
  - Standardized `GlobusError` type across all services
  - Added consistent error context and debugging information
  - Implemented service-specific error codes and messages
- **Consistent client initialization patterns**
  - Unified `NewClient()` functions across all services
  - Standardized configuration and options handling
  - Enhanced client lifecycle management
- **Standardized response and pagination patterns**
  - Unified `Response[T]` wrapper structures
  - Consistent `PaginatedResponse[T]` across all services
  - Enhanced metadata handling and request tracking
- **Updated API versions to match current Globus APIs**
  - Transfer API updated to latest v0.10+ endpoints
  - Auth API aligned with current OAuth2 specifications
  - Groups API updated to v2 endpoints
  - Search API updated to v1 with latest features
  - Flows API updated to v1 endpoints
  - Compute API updated to v2 endpoints
  - Timers API updated to v1 endpoints
- **Enhanced deprecation system matching Python SDK**
  - Added deprecation warnings and migration guidance
  - Implemented deprecation lifecycle management
  - Added deprecation reporting tools
- **Keep a Changelog compliance**
  - Improved changelog structure and consistency
  - Added semantic versioning compliance
  - Enhanced release documentation standards

### Changed
- **BREAKING**: Version updated from v0.9.15 to v3.60.0 to align with Python SDK
- **BREAKING**: Unified error handling - all services now use `GlobusError` type
- **BREAKING**: Standardized client initialization - all services use consistent `NewClient()` pattern
- **BREAKING**: Consistent response structures - all services use `Response[T]` wrapper
- **BREAKING**: Updated API endpoints to match current Globus APIs
- **BREAKING**: Reorganized package structure for better consistency
- Enhanced documentation structure and consistency
- Updated examples and documentation for v3.60.0

### Deprecated
- Legacy error handling patterns (will be removed in v4.0.0)
- Old client initialization methods (will be removed in v4.0.0)
- Inconsistent response structures (will be removed in v4.0.0)

### Removed
- Legacy debugging utilities (moved to proper package structure)
- Deprecated lint tools (replaced with modern alternatives)
- Inconsistent internal APIs (replaced with unified patterns)

### Fixed
- Fixed internal consistency issues across services
- Corrected API version mismatches
- Fixed package conflicts in debug files
- Resolved function redeclarations across the codebase
- Updated auth and transfer client usage patterns
- Replaced deprecated io/ioutil with io package functions
- Fixed variable naming to avoid conflicts (e.g., `err` → `tokenErr`)
- Improved error handling in contract tests
- Fixed missing imports in compute example files

### Security
- Enhanced token handling security
- Improved credential validation mechanisms  
- Updated security practices to match current standards

### Migration from v0.9.x

This release introduces **breaking changes** that require migration:

1. **Update import paths**:
   ```go
   // OLD
   import "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
   
   // NEW  
   import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
   ```

2. **Update go.mod**:
   ```bash
   go get github.com/scttfrdmn/globus-go-sdk/v3
   ```

3. **Version tracking**: The SDK now follows Python SDK versioning with format `[PYTHON_SDK_VERSION]-[GO_SDK_BUILD]`

4. **GARE support**: New error handling for dependent consent scenarios - see auth package documentation

For detailed migration guidance, see [VERSIONING_STRATEGY.md](VERSIONING_STRATEGY.md).

## [0.9.15] - 2025-05-08

### Fixed
- Properly tagged release for the connection pool functions fix (issue #13)
  - Ensured correct Git tag pointing to the fixed code
  - Verified build works with downstream dependencies
  - Fixed tagging issues from previous release attempts

## [0.9.15] - 2025-05-08

### Fixed
- Properly tagged release for the connection pool functions fix (issue #13)
  - Ensured correct Git tag pointing to the fixed code
  - Verified build works with downstream dependencies
  - Fixed tagging issues from previous release attempts

## [0.9.14] - 2025-05-07

### Fixed
- Verified and reinforced fix for missing connection pool functions (issue #13)
  - Added comprehensive test suite to validate connection pool functions
  - Added verification script to confirm proper implementation
  - Ensured all required functions are properly defined and exported

### Added
- Comprehensive test coverage for connection pool functions
  - Added unit tests in pkg/core/connection_pool_test.go
  - Added verification script in scripts/verify_connection_pool_functions.go
  - Added test harness to simulate downstream project usage

## [0.9.13] - 2025-05-07

### Fixed
- Restored missing connection pool functions that were referenced in transport_init.go
  - Added missing SetConnectionPoolManager and EnableDefaultConnectionPool functions
  - Added GetConnectionPool and GetHTTPClientForService helper functions
  - Fixed breaking changes introduced in v0.9.11 that caused downstream projects to fail compilation

## [0.9.11] - 2025-05-07

### Fixed
- Fixed string formatting issues in example files
- Added missing ExpiresAt() method to TokenResponse in auth package
- Fixed client initialization patterns with proper error handling
- Fixed GitHub Actions workflow for API compatibility testing
- Updated API compatibility workflow to properly handle GitHub token authentication
- Fixed type references in integration tests

## [0.9.10] - 2025-05-07

### Fixed
- Fixed build error with undefined `httppool.NewHttpConnectionPoolManager` function
- Updated connection pool initialization to use the global pool manager

## [0.9.9] - 2025-05-07

### Added
- Comprehensive API compatibility testing suite
- Interface implementation verification tests
- Dependent project build test script
- Compiler-enforced API contracts using interfaces
- GitHub Actions workflow for API compatibility checks

### Changed
- Updated version to 0.9.9

## [0.9.8] - 2025-05-07

### Fixed
- Added GetVersionCheck() and SetVersionCheck() methods to Config in pkg/core/config/config.go
- Updated api_version.go to use GetVersionCheck() and SetVersionCheck() instead of direct field access
- Added SyncChecksum alias for SyncLevelChecksum in transfer package for backward compatibility
- Updated version to 0.9.8

## [0.9.7] - 2025-05-07

### Fixed
- Fixed mfaErr variable detection in auth/mfa.go
- Ensured VersionCheck field in Config struct is properly exported

## [0.9.6] - 2025-05-07

### Fixed
- Fixed duplicate tokenRequest method in auth/mfa.go
- Fixed type naming consistency with ClientConfig in transfer package
- Fixed incorrect DeleteItem structure in test and debug files
- Removed redundant Recursive field from DeleteItem that's unsupported by the API
- Fixed JSON marshaling issues with function fields in ResumableTransferOptions
- Added proper DataType setting for TransferItems in resumable transfers
- Fixed duplicate setupMockServer functions in transfer tests

## [0.9.5] - 2025-05-07

### Fixed
- Resolved import cycle issues between packages
- Restructured connection pool management to use interfaces
- Added additional pool configuration capabilities
- Created improved pool manager implementation

## [0.9.4] - 2025-05-07

### Fixed
- Added missing ClientInterface methods to Client type
- Fixed unused imports in client_with_pool.go
- Resolved interface implementation issues causing compilation errors in consuming applications

## [0.9.3] - 2025-05-07

### Fixed
- Added missing logging.go file in transport package that caused compilation errors
- Fixed "undefined: logRequest and logResponse" errors when using the SDK

## [0.9.2] - 2025-05-07

### Added
- Versioned documentation with Hugo-book theme
- GitHub Pages deployment workflows for documentation
- Comprehensive documentation for all API surfaces
- Enhanced GitHub Actions workflows with better CI/CD integration

### Fixed
- Documentation deployment issues
- Version compatibility checking in service clients
- GitHub Pages configuration
- Minor documentation formatting issues

## [0.9.1] - 2025-05-02

### Fixed
- Added missing interfaces package required by SDK consumers
- Fixed dependency issues when importing the SDK
- Added interface definitions for authorization, client operations, connection pools, and transport

## [0.9.0] - 2025-05-02

### Added
- Enhanced Compute service with workflow and task group capabilities
- Workflow management (creation, execution, status tracking)
- Dependency graph support for complex compute workflows
- Task group functionality for parallel execution
- Expanded container management capabilities
- Environment and secret management
- Improved API version compatibility checking
- Enhanced HTTP debugging with detailed request/response logging
- New example for Compute workflows and task groups

### Fixed
- Improved error handling in transport layer
- Enhanced connection pool management for better stability
- Fixed integration tests for all service clients
- Standardized error reporting formats across services
- Improved thread safety in concurrent operations

## [0.8.0] - 2025-03-15

### Added
- Compute service implementation
  - Batch job support
  - Container management
  - Dependency handling
  - Environment configuration
- Enhanced Auth package with options pattern
- Added Transport layer interfaces

### Changed
- Updated client implementation with connection pooling
- Improved error handling
- Enhanced logging with context-based logging

### Fixed
- Token refresh handling
- Race conditions in transport layer
- Authentication error handling

## [0.7.0] - 2025-01-30

### Added
- Flows service implementation
  - Flow management
  - Execution control
  - Status monitoring
- Search service implementation
  - Advanced query capabilities
  - Indexing operations
  - Result pagination
- Timers service implementation

### Changed
- Refactored Transfer service for better performance
- Improved error types and handling
- Enhanced documentation

### Fixed
- Memory leaks in Transfer operations
- Authentication token handling bugs

## [0.6.0] - 2024-12-05

### Added
- Groups service implementation
  - Group management (create, list, update, delete)
  - Membership management (add, remove, update roles)
  - Role management operations
- Transfer service implementation
  - File and directory operations
  - Task management
  - Status monitoring
- Auth service implementation
  - OAuth flow implementations
  - Token management

### Changed
- Improved SDK configuration options
- Enhanced error handling

### Fixed
- Connection handling in HTTP client
- Error propagation issues

## [0.5.0] - 2024-10-15

### Added
- Initial SDK framework
- Core client implementation
- Configuration management
- HTTP transport layer
- Basic authorization mechanisms

[Unreleased]: https://github.com/scttfrdmn/globus-go-sdk/compare/v3.60.0...HEAD
[3.60.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.15...v3.60.0
[0.9.15]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.14...v0.9.15
[0.9.14]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.13...v0.9.14
[0.9.13]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.12...v0.9.13
[0.9.12]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.11...v0.9.12
[0.9.11]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.10...v0.9.11
[0.9.10]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.9...v0.9.10
[0.9.9]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.8...v0.9.9
[0.9.8]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.7...v0.9.8
[0.9.7]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.6...v0.9.7
[0.9.6]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.5...v0.9.6
[0.9.5]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.4...v0.9.5
[0.9.4]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.3...v0.9.4
[0.9.3]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v0.5.0