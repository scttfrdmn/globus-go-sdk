<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Package stability indicators throughout the SDK
  - Added doc.go files with STABILITY levels (stable, beta, alpha, experimental)
  - Added explicit API component listings for each package
  - Documented compatibility guarantees and examples
- Updated CLAUDE.md with API stability guidelines and examples
- Enhanced CHANGELOG structure to track API changes more explicitly
- Comprehensive API stability verification tools:
  - Added `cmd/apicompare` - Tool for comparing APIs between versions
  - Added `cmd/apigen` - Tool for generating API signatures from source
  - Added `cmd/depreport` - Tool for generating deprecation reports
  - Added contract testing framework in `pkg/core/contracts`
- New documentation for API stability implementation:
  - Added API_STABILITY_PHASE2_SUMMARY.md with details on implemented tools
  - Added API_DEPRECATION_SYSTEM.md explaining the deprecation system
  - Added CONTRACT_TESTING.md with contract testing documentation
- CI/CD integration for API compatibility verification
- Improved organization of verification tools:
  - Moved script utilities to proper cmd/ directories
  - Created `internal/verification` package for common code

### Changed
- Updated documentation to clarify stability levels of different components
- Core package marked as BETA due to ongoing connection pool improvements
- Transfer package components now have explicit stability indicators
- Switched from deprecated `golint` to `staticcheck` for code linting
- Improved pre-commit hooks with enhanced verification
- Restructured debug code to use proper package organization

### Deprecated
- No functionality has been deprecated in this release

### Removed
- No functionality has been removed in this release

### Fixed
- Fixed package conflicts in debug files
- Resolved function redeclarations across the codebase
- Updated auth and transfer client usage patterns
- Replaced deprecated io/ioutil with io package functions
- Fixed variable naming to avoid conflicts (e.g., `err` → `tokenErr`)
- Improved error handling in contract tests
- Fixed missing imports in compute example files

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

[Unreleased]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.15...HEAD
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